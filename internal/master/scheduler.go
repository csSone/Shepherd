// Package master provides intelligent scheduling for model execution in distributed architecture.
package master

import (
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/shepherd-project/shepherd/Shepherd/internal/logger"
	"github.com/shepherd-project/shepherd/Shepherd/internal/node"
)

// ModelRequest 表示模型运行请求
type ModelRequest struct {
	ModelName      string            `json:"modelName"`      // 模型名称
	ModelSize      int64             `json:"modelSize"`      // 模型大小（字节）
	RequiredMemory int64             `json:"requiredMemory"` // 所需内存（字节）
	RequireGPU     bool              `json:"requireGPU"`     // 是否需要GPU
	GPUBackend     string            `json:"gpuBackend"`     // GPU后端类型（cuda/rocm）
	GPUMemory      int64             `json:"gpuMemory"`      // 所需GPU内存（字节）
	ContextSize    int               `json:"contextSize"`    // 上下文大小
	Priority       int               `json:"priority"`       // 请求优先级（0-10）
	Timeout        time.Duration     `json:"timeout"`        // 超时时间
	Metadata       map[string]string `json:"metadata"`       // 请求元数据
}

// SchedulingStrategy 定义调度策略接口
type SchedulingStrategy interface {
	// SelectNode 从候选节点中选择最适合的节点
	SelectNode(candidates []*node.NodeInfo, modelReq *ModelRequest) (*node.NodeInfo, error)
	// Name 返回策略名称
	Name() string
}

// ResourceBasedStrategy 基于资源的调度策略（选择资源最充足的节点）
type ResourceBasedStrategy struct{}

// LoadBalancedStrategy 基于负载均衡的调度策略（选择负载最低的节点）
type LoadBalancedStrategy struct{}

// LocalityStrategy 基于本地性的调度策略（优先选择已有模型的节点）
type LocalityStrategy struct {
	// 模型到节点的映射缓存
	modelCache map[string][]string // model name -> node IDs
	mu         sync.RWMutex
}

// NewLocalityStrategy 创建新的本地性策略
func NewLocalityStrategy() *LocalityStrategy {
	return &LocalityStrategy{
		modelCache: make(map[string][]string),
	}
}

// UpdateModelCache 更新模型缓存
func (s *LocalityStrategy) UpdateModelCache(modelName string, nodeIDs []string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.modelCache[modelName] = nodeIDs
}

// Name 实现SchedulingStrategy接口
func (s *ResourceBasedStrategy) Name() string {
	return "ResourceBased"
}

// SelectNode 实现基于资源的节点选择
func (s *ResourceBasedStrategy) SelectNode(candidates []*node.NodeInfo, modelReq *ModelRequest) (*node.NodeInfo, error) {
	if len(candidates) == 0 {
		return nil, fmt.Errorf("没有可用的候选节点")
	}

	// 过滤满足资源需求的节点
	validNodes := make([]*node.NodeInfo, 0)
	for _, node := range candidates {
		if s.validateNodeResources(node, modelReq) {
			validNodes = append(validNodes, node)
		}
	}

	if len(validNodes) == 0 {
		return nil, fmt.Errorf("没有节点满足资源需求")
	}

	// 按资源总量排序，选择资源最充足的节点
	sort.Slice(validNodes, func(i, j int) bool {
		scoreI := s.calculateResourceScore(validNodes[i])
		scoreJ := s.calculateResourceScore(validNodes[j])
		return scoreI > scoreJ
	})

	return validNodes[0], nil
}

// validateNodeResources 验证节点是否满足资源需求
func (s *ResourceBasedStrategy) validateNodeResources(node *node.NodeInfo, req *ModelRequest) bool {
	if node.Resources == nil {
		return false
	}

	// 检查内存需求
	availableMemory := node.Resources.MemoryTotal - node.Resources.MemoryUsed
	if req.RequiredMemory > 0 && availableMemory < req.RequiredMemory {
		return false
	}

	// 检查GPU需求
	if req.RequireGPU {
		if node.Capabilities == nil || !node.Capabilities.GPU || node.Capabilities.GPUCount == 0 {
			return false
		}

		// 检查GPU后端类型（如果有特定需求）
		if req.GPUBackend != "" {
			// 这里简化处理，实际应该检查GPU类型
			if len(node.Capabilities.GPUNames) > 0 {
				// 简单检查GPU名称是否包含后端类型
				hasBackend := false
				for _, gpuName := range node.Capabilities.GPUNames {
					if containsIgnoreCase(gpuName, req.GPUBackend) {
						hasBackend = true
						break
					}
				}
				if !hasBackend {
					return false
				}
			}
		}

		// 检查GPU内存需求
		if req.GPUMemory > 0 && node.Resources.GPUInfo != nil {
			hasEnoughGPUMemory := false
			for _, gpu := range node.Resources.GPUInfo {
				availableGPUMemory := gpu.TotalMemory - gpu.UsedMemory
				if availableGPUMemory >= req.GPUMemory {
					hasEnoughGPUMemory = true
					break
				}
			}
			if !hasEnoughGPUMemory {
				return false
			}
		}
	}

	return true
}

// calculateResourceScore 计算节点资源评分
func (s *ResourceBasedStrategy) calculateResourceScore(node *node.NodeInfo) float64 {
	if node.Resources == nil {
		return 0
	}

	score := 0.0

	// CPU资源评分（绝对可用CPU核心数，考虑核心数权重）
	availableCPU := float64(node.Resources.CPUTotal - node.Resources.CPUUsed)
	cpuScore := availableCPU * 0.00001 // 将millicores转换为合理的分数
	score += cpuScore

	// 内存资源评分（绝对可用内存，考虑GB权重）
	if node.Resources.MemoryTotal > 0 {
		availableMem := float64(node.Resources.MemoryTotal - node.Resources.MemoryUsed)
		memScore := availableMem / 1000000000 * 0.01 // 将字节转换为GB并计算分数
		score += memScore
	}

	// GPU资源评分（GPU数量是最重要的因素）
	if node.Capabilities != nil && node.Capabilities.GPU && node.Capabilities.GPUCount > 0 {
		gpuScore := float64(node.Capabilities.GPUCount) * 10 // 每个GPU加10分
		score += gpuScore

		// GPU内存评分（绝对可用GPU内存）
		if node.Resources.GPUInfo != nil {
			availableGPUMem := int64(0)
			for _, gpu := range node.Resources.GPUInfo {
				availableGPUMem += (gpu.TotalMemory - gpu.UsedMemory)
			}
			gpuMemScore := float64(availableGPUMem) / 1000000000 * 0.1 // 将字节转换为GB并计算分数
			score += gpuMemScore
		}
	}

	return score
}

// Name 实现SchedulingStrategy接口
func (s *LoadBalancedStrategy) Name() string {
	return "LoadBalanced"
}

// SelectNode 实现基于负载均衡的节点选择
func (s *LoadBalancedStrategy) SelectNode(candidates []*node.NodeInfo, modelReq *ModelRequest) (*node.NodeInfo, error) {
	if len(candidates) == 0 {
		return nil, fmt.Errorf("没有可用的候选节点")
	}

	// 过滤满足资源需求的节点
	validNodes := make([]*node.NodeInfo, 0)
	for _, node := range candidates {
		if s.validateNodeResources(node, modelReq) {
			validNodes = append(validNodes, node)
		}
	}

	if len(validNodes) == 0 {
		return nil, fmt.Errorf("没有节点满足资源需求")
	}

	// 按负载排序，选择负载最低的节点
	sort.Slice(validNodes, func(i, j int) bool {
		loadI := s.calculateLoad(validNodes[i])
		loadJ := s.calculateLoad(validNodes[j])
		return loadI < loadJ // 负载越低越好
	})

	return validNodes[0], nil
}

// validateNodeResources 验证节点是否满足资源需求（复用ResourceBasedStrategy的逻辑）
func (s *LoadBalancedStrategy) validateNodeResources(node *node.NodeInfo, req *ModelRequest) bool {
	if node.Resources == nil {
		return false
	}

	// 检查内存需求
	availableMemory := node.Resources.MemoryTotal - node.Resources.MemoryUsed
	if req.RequiredMemory > 0 && availableMemory < req.RequiredMemory {
		return false
	}

	// 检查GPU需求
	if req.RequireGPU {
		if node.Capabilities == nil || !node.Capabilities.GPU || node.Capabilities.GPUCount == 0 {
			return false
		}

		// 检查GPU后端类型（如果有特定需求）
		if req.GPUBackend != "" {
			// 简化处理，实际应该检查GPU类型
			if len(node.Capabilities.GPUNames) > 0 {
				hasBackend := false
				for _, gpuName := range node.Capabilities.GPUNames {
					if containsIgnoreCase(gpuName, req.GPUBackend) {
						hasBackend = true
						break
					}
				}
				if !hasBackend {
					return false
				}
			}
		}

		// 检查GPU内存需求
		if req.GPUMemory > 0 && node.Resources.GPUInfo != nil {
			hasEnoughGPUMemory := false
			for _, gpu := range node.Resources.GPUInfo {
				availableGPUMemory := gpu.TotalMemory - gpu.UsedMemory
				if availableGPUMemory >= req.GPUMemory {
					hasEnoughGPUMemory = true
					break
				}
			}
			if !hasEnoughGPUMemory {
				return false
			}
		}
	}

	return true
}

// calculateLoad 计算节点负载评分
func (s *LoadBalancedStrategy) calculateLoad(node *node.NodeInfo) float64 {
	if node.Resources == nil {
		return 100.0 // 如果没有资源信息，认为负载很高
	}

	load := 0.0

	// CPU负载
	if node.Resources.CPUTotal > 0 {
		cpuLoad := float64(node.Resources.CPUUsed) / float64(node.Resources.CPUTotal)
		load += cpuLoad * 0.4 // 40%权重
	}

	// 内存负载
	if node.Resources.MemoryTotal > 0 {
		memLoad := float64(node.Resources.MemoryUsed) / float64(node.Resources.MemoryTotal)
		load += memLoad * 0.4 // 40%权重
	}

	// 系统平均负载
	if node.Resources.LoadAverage != nil && len(node.Resources.LoadAverage) > 0 {
		// 使用1分钟平均负载
		sysLoad := node.Resources.LoadAverage[0]
		// 假设CPU核心数
		cpuCores := float64(node.Resources.CPUTotal) / 1000.0
		if cpuCores > 0 {
			sysLoadScore := sysLoad / cpuCores
			load += sysLoadScore * 0.2 // 20%权重
		}
	}

	return load
}

// Name 实现SchedulingStrategy接口
func (s *LocalityStrategy) Name() string {
	return "Locality"
}

// SelectNode 实现基于本地性的节点选择
func (s *LocalityStrategy) SelectNode(candidates []*node.NodeInfo, modelReq *ModelRequest) (*node.NodeInfo, error) {
	if len(candidates) == 0 {
		return nil, fmt.Errorf("没有可用的候选节点")
	}

	s.mu.RLock()
	// 查找已有模型的节点
	preferredNodeIDs, exists := s.modelCache[modelReq.ModelName]
	s.mu.RUnlock()

	// 如果有节点已有该模型，优先从这些节点中选择
	if exists && len(preferredNodeIDs) > 0 {
		preferredNodes := make([]*node.NodeInfo, 0)
		nodeIDMap := make(map[string]*node.NodeInfo)

		// 构建节点ID映射
		for _, node := range candidates {
			nodeIDMap[node.ID] = node
		}

		// 查找首选节点
		for _, nodeID := range preferredNodeIDs {
			if node, ok := nodeIDMap[nodeID]; ok {
				// 确保节点满足资源需求
				if s.validateNodeResources(node, modelReq) {
					preferredNodes = append(preferredNodes, node)
				}
			}
		}

		// 如果有满足条件的首选节点，按负载选择最优的
		if len(preferredNodes) > 0 {
			lbStrategy := &LoadBalancedStrategy{}
			return lbStrategy.SelectNode(preferredNodes, modelReq)
		}
	}

	// 如果没有节点已有该模型，回退到负载均衡策略
	lbStrategy := &LoadBalancedStrategy{}
	return lbStrategy.SelectNode(candidates, modelReq)
}

// validateNodeResources 验证节点是否满足资源需求（复用LoadBalancedStrategy的逻辑）
func (s *LocalityStrategy) validateNodeResources(node *node.NodeInfo, req *ModelRequest) bool {
	if node.Resources == nil {
		return false
	}

	// 检查内存需求
	availableMemory := node.Resources.MemoryTotal - node.Resources.MemoryUsed
	if req.RequiredMemory > 0 && availableMemory < req.RequiredMemory {
		return false
	}

	// 检查GPU需求
	if req.RequireGPU {
		if node.Capabilities == nil || !node.Capabilities.GPU || node.Capabilities.GPUCount == 0 {
			return false
		}

		// 检查GPU后端类型（如果有特定需求）
		if req.GPUBackend != "" {
			// 简化处理，实际应该检查GPU类型
			if len(node.Capabilities.GPUNames) > 0 {
				hasBackend := false
				for _, gpuName := range node.Capabilities.GPUNames {
					if containsIgnoreCase(gpuName, req.GPUBackend) {
						hasBackend = true
						break
					}
				}
				if !hasBackend {
					return false
				}
			}
		}

		// 检查GPU内存需求
		if req.GPUMemory > 0 && node.Resources.GPUInfo != nil {
			hasEnoughGPUMemory := false
			for _, gpu := range node.Resources.GPUInfo {
				availableGPUMemory := gpu.TotalMemory - gpu.UsedMemory
				if availableGPUMemory >= req.GPUMemory {
					hasEnoughGPUMemory = true
					break
				}
			}
			if !hasEnoughGPUMemory {
				return false
			}
		}
	}

	return true
}

// Scheduler 智能调度器
type Scheduler struct {
	nodeManager      *NodeManager
	strategy         SchedulingStrategy
	defaultStrategy  SchedulingStrategy
	localityStrategy *LocalityStrategy
	log              *logger.Logger
	mu               sync.RWMutex
}

// NewScheduler 创建新的调度器
func NewScheduler(nodeManager *NodeManager, log *logger.Logger) *Scheduler {
	localityStrategy := NewLocalityStrategy()
	return &Scheduler{
		nodeManager:      nodeManager,
		defaultStrategy:  &LoadBalancedStrategy{}, // 默认使用负载均衡策略
		localityStrategy: localityStrategy,
		log:              log,
	}
}

// SetStrategy 设置调度策略
func (s *Scheduler) SetStrategy(strategy SchedulingStrategy) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.strategy = strategy
	s.log.Infof("调度策略已更改为: %s", strategy.Name())
}

// GetStrategy 获取当前调度策略
func (s *Scheduler) GetStrategy() SchedulingStrategy {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.strategy != nil {
		return s.strategy
	}
	return s.defaultStrategy
}

// UpdateModelCache 更新模型本地性缓存
func (s *Scheduler) UpdateModelCache(modelName string, nodeIDs []string) {
	s.localityStrategy.UpdateModelCache(modelName, nodeIDs)
	s.log.Debugf("更新模型缓存: %s -> %v", modelName, nodeIDs)
}

// Schedule 调度模型运行请求到合适的节点
func (s *Scheduler) Schedule(modelReq *ModelRequest) (*node.NodeInfo, error) {
	s.mu.RLock()
	strategy := s.strategy
	s.mu.RUnlock()

	// 如果没有指定策略，使用默认策略
	if strategy == nil {
		strategy = s.defaultStrategy
	}

	// 获取在线节点列表
	candidates := s.nodeManager.ListOnlineNodes()
	if len(candidates) == 0 {
		return nil, fmt.Errorf("没有可用的在线节点")
	}

	s.log.Debugf("开始调度模型 %s，候选节点数: %d，使用策略: %s",
		modelReq.ModelName, len(candidates), strategy.Name())

	// 使用指定策略选择节点
	selectedNode, err := strategy.SelectNode(candidates, modelReq)
	if err != nil {
		return nil, fmt.Errorf("节点选择失败: %w", err)
	}

	s.log.Infof("成功调度模型 %s 到节点 %s (%s:%d)",
		modelReq.ModelName, selectedNode.ID, selectedNode.Address, selectedNode.Port)

	return selectedNode, nil
}

// GetAvailableStrategies 获取所有可用的调度策略
func (s *Scheduler) GetAvailableStrategies() []string {
	return []string{
		(&ResourceBasedStrategy{}).Name(),
		(&LoadBalancedStrategy{}).Name(),
		(&LocalityStrategy{}).Name(),
	}
}

// containsIgnoreCase 忽略大小写的字符串包含检查
func containsIgnoreCase(s, substr string) bool {
	s = toLower(s)
	substr = toLower(substr)
	return contains(s, substr)
}

// toLower 转换为小写
func toLower(s string) string {
	result := make([]rune, len(s))
	for i, r := range s {
		if r >= 'A' && r <= 'Z' {
			result[i] = r + ('a' - 'A')
		} else {
			result[i] = r
		}
	}
	return string(result)
}

// contains 检查字符串是否包含子串
func contains(s, substr string) bool {
	return len(s) >= len(substr) && findSubstring(s, substr)
}

// findSubstring 查找子串
func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		match := true
		for j := 0; j < len(substr); j++ {
			if s[i+j] != substr[j] {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}
