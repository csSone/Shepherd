// Package master provides node management for the master node in distributed architecture.
//
// Deprecated: This package is part of the old distributed architecture.
// New code should use github.com/shepherd-project/shepherd/Shepherd/internal/node.Node with Master role instead.
// This package will be removed in a future release.
package master

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/shepherd-project/shepherd/Shepherd/internal/logger"
	"github.com/shepherd-project/shepherd/Shepherd/internal/node"
)

// NodeManager 管理连接到 Master 的 Client 节点
//
// Deprecated: Use node.Node with Master role instead. This type will be removed in a future release.
type NodeManager struct {
	nodes  map[string]*node.NodeInfo
	mu     sync.RWMutex
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
	log    *logger.Logger

	// 配置参数
	timeout       time.Duration // 心跳超时时间
	checkInterval time.Duration // 检查间隔

	// 事件通道
	eventChan chan *node.NodeEvent

	// 运行状态
	running bool
}

// NewNodeManager 创建新的节点管理器
func NewNodeManager(log *logger.Logger) *NodeManager {
	ctx, cancel := context.WithCancel(context.Background())

	return &NodeManager{
		nodes:         make(map[string]*node.NodeInfo),
		ctx:           ctx,
		cancel:        cancel,
		log:           log,
		timeout:       15 * time.Second,                // 15秒超时（3个心跳周期）
		checkInterval: 5 * time.Second,                 // 5秒检查间隔
		eventChan:     make(chan *node.NodeEvent, 100), // 缓冲100个事件
		running:       false,
	}
}

// Start 启动节点管理器
func (nm *NodeManager) Start() {
	nm.mu.Lock()
	defer nm.mu.Unlock()

	if nm.running {
		return
	}

	nm.running = true
	nm.log.Info("启动节点管理器")

	// 启动超时检查协程
	nm.wg.Add(1)
	go nm.timeoutChecker()

	// 启动事件广播协程
	nm.wg.Add(1)
	go nm.eventBroadcaster()
}

// Stop 停止节点管理器
func (nm *NodeManager) Stop() {
	nm.mu.Lock()
	defer nm.mu.Unlock()

	if !nm.running {
		return
	}

	nm.running = false
	nm.log.Info("停止节点管理器")

	// 取消上下文
	nm.cancel()

	// 等待所有协程结束
	nm.wg.Wait()

	// 关闭事件通道
	close(nm.eventChan)
}

// RegisterNode 注册新节点
func (nm *NodeManager) RegisterNode(nodeInfo *node.NodeInfo) error {
	if nodeInfo == nil {
		return fmt.Errorf("节点信息不能为空")
	}

	if nodeInfo.ID == "" {
		return fmt.Errorf("节点ID不能为空")
	}

	nm.mu.Lock()
	defer nm.mu.Unlock()

	// 检查节点是否已存在
	if existing, exists := nm.nodes[nodeInfo.ID]; exists {
		// 更新现有节点
		existing.Name = nodeInfo.Name
		existing.Address = nodeInfo.Address
		existing.Port = nodeInfo.Port
		existing.Version = nodeInfo.Version
		existing.Capabilities = nodeInfo.Capabilities
		existing.Resources = nodeInfo.Resources
		existing.Status = node.NodeStatusOnline
		existing.LastSeen = time.Now()
		existing.UpdatedAt = time.Now()

		nm.log.Info(fmt.Sprintf("节点重新注册: %s (%s) from %s:%d", nodeInfo.ID, nodeInfo.Name, nodeInfo.Address, nodeInfo.Port))

		// 发送节点更新事件
		nm.sendEvent(&node.NodeEvent{
			ID:        fmt.Sprintf("node-update-%s", nodeInfo.ID),
			NodeID:    nodeInfo.ID,
			Type:      "node_update",
			Severity:  "info",
			Message:   fmt.Sprintf("节点 %s 重新注册", nodeInfo.ID),
			Details:   map[string]interface{}{"node": existing},
			Timestamp: time.Now(),
		})
	} else {
		// 添加新节点
		nodeInfo.Status = node.NodeStatusOnline
		nodeInfo.LastSeen = time.Now()
		nodeInfo.CreatedAt = time.Now()
		nodeInfo.UpdatedAt = time.Now()
		nm.nodes[nodeInfo.ID] = nodeInfo

		nm.log.Info(fmt.Sprintf("新节点注册: %s (%s) from %s:%d", nodeInfo.ID, nodeInfo.Name, nodeInfo.Address, nodeInfo.Port))

		// 发送节点注册事件
		nm.sendEvent(&node.NodeEvent{
			ID:        fmt.Sprintf("node-register-%s", nodeInfo.ID),
			NodeID:    nodeInfo.ID,
			Type:      "node_register",
			Severity:  "info",
			Message:   fmt.Sprintf("新节点 %s 注册", nodeInfo.ID),
			Details:   map[string]interface{}{"node": nodeInfo},
			Timestamp: time.Now(),
		})
	}

	return nil
}

// HandleHeartbeat 处理心跳消息
func (nm *NodeManager) HandleHeartbeat(nodeID string, heartbeat *node.HeartbeatMessage) error {
	if heartbeat == nil {
		return fmt.Errorf("心跳消息不能为空")
	}

	if heartbeat.NodeID != nodeID {
		return fmt.Errorf("心跳消息节点ID不匹配")
	}

	nm.mu.Lock()
	defer nm.mu.Unlock()

	nodeInfo, exists := nm.nodes[nodeID]
	if !exists {
		return fmt.Errorf("节点不存在: %s", nodeID)
	}

	// 更新节点状态和资源信息
	nodeInfo.Status = heartbeat.Status
	nodeInfo.LastSeen = heartbeat.Timestamp
	nodeInfo.UpdatedAt = time.Now()
	nodeInfo.Resources = heartbeat.Resources
	nodeInfo.Capabilities = heartbeat.Capabilities

	return nil
}

// CheckTimeouts 检查超时节点并标记为离线
func (nm *NodeManager) CheckTimeouts() {
	nm.mu.Lock()
	defer nm.mu.Unlock()

	now := time.Now()
	timeoutNodes := make([]string, 0)

	for nodeID, nodeInfo := range nm.nodes {
		if nodeInfo.Status == node.NodeStatusOnline && now.Sub(nodeInfo.LastSeen) > nm.timeout {
			timeoutNodes = append(timeoutNodes, nodeID)
		}
	}

	// 标记超时节点为离线
	for _, nodeID := range timeoutNodes {
		nodeInfo := nm.nodes[nodeID]
		nodeInfo.Status = node.NodeStatusOffline
		nodeInfo.UpdatedAt = time.Now()

		nm.log.Warn(fmt.Sprintf("节点超时离线: %s (%s)", nodeID, nodeInfo.Name))

		// 发送节点离线事件
		nm.sendEvent(&node.NodeEvent{
			ID:        fmt.Sprintf("node-timeout-%s", nodeID),
			NodeID:    nodeID,
			Type:      "node_timeout",
			Severity:  "warning",
			Message:   fmt.Sprintf("节点 %s 超时离线", nodeID),
			Details:   map[string]interface{}{"node": nodeInfo},
			Timestamp: time.Now(),
		})
	}
}

// GetNode 获取单个节点
func (nm *NodeManager) GetNode(nodeID string) (*node.NodeInfo, error) {
	nm.mu.RLock()
	defer nm.mu.RUnlock()

	nodeInfo, exists := nm.nodes[nodeID]
	if !exists {
		return nil, fmt.Errorf("节点不存在: %s", nodeID)
	}

	// 返回节点信息的副本
	return nm.copyNodeInfo(nodeInfo), nil
}

// ListNodes 获取所有节点列表
func (nm *NodeManager) ListNodes() []*node.NodeInfo {
	nm.mu.RLock()
	defer nm.mu.RUnlock()

	nodes := make([]*node.NodeInfo, 0, len(nm.nodes))
	for _, nodeInfo := range nm.nodes {
		nodes = append(nodes, nm.copyNodeInfo(nodeInfo))
	}
	return nodes
}

// ListOnlineNodes 获取在线节点列表
func (nm *NodeManager) ListOnlineNodes() []*node.NodeInfo {
	nm.mu.RLock()
	defer nm.mu.RUnlock()

	nodes := make([]*node.NodeInfo, 0)
	for _, nodeInfo := range nm.nodes {
		if nodeInfo.Status == node.NodeStatusOnline {
			nodes = append(nodes, nm.copyNodeInfo(nodeInfo))
		}
	}
	return nodes
}

// UpdateNodeStatus 更新节点状态
func (nm *NodeManager) UpdateNodeStatus(nodeID string, status node.NodeStatus) error {
	nm.mu.Lock()
	defer nm.mu.Unlock()

	nodeInfo, exists := nm.nodes[nodeID]
	if !exists {
		return fmt.Errorf("节点不存在: %s", nodeID)
	}

	oldStatus := nodeInfo.Status
	nodeInfo.Status = status
	nodeInfo.UpdatedAt = time.Now()

	if oldStatus != status {
		nm.log.Info(fmt.Sprintf("节点状态更新: %s (%s -> %s)", nodeID, oldStatus, status))

		// 发送状态变更事件
		nm.sendEvent(&node.NodeEvent{
			ID:        fmt.Sprintf("node-status-%s", nodeID),
			NodeID:    nodeID,
			Type:      "status_change",
			Severity:  "info",
			Message:   fmt.Sprintf("节点 %s 状态变更: %s -> %s", nodeID, oldStatus, status),
			Details:   map[string]interface{}{"oldStatus": oldStatus, "newStatus": status},
			Timestamp: time.Now(),
		})
	}

	return nil
}

// GetEventChannel 获取事件通道
func (nm *NodeManager) GetEventChannel() <-chan *node.NodeEvent {
	return nm.eventChan
}

// GetNodeCount 获取节点统计信息
func (nm *NodeManager) GetNodeCount() (total, online, offline, busy int) {
	nm.mu.RLock()
	defer nm.mu.RUnlock()

	total = len(nm.nodes)
	for _, nodeInfo := range nm.nodes {
		switch nodeInfo.Status {
		case node.NodeStatusOnline:
			online++
		case node.NodeStatusOffline:
			offline++
		case node.NodeStatusBusy:
			busy++
		}
	}
	return
}

// timeoutChecker 定期检查节点超时的协程
func (nm *NodeManager) timeoutChecker() {
	defer nm.wg.Done()

	ticker := time.NewTicker(nm.checkInterval)
	defer ticker.Stop()

	for {
		select {
		case <-nm.ctx.Done():
			return
		case <-ticker.C:
			nm.CheckTimeouts()
		}
	}
}

// eventBroadcaster 事件广播协程
func (nm *NodeManager) eventBroadcaster() {
	defer nm.wg.Done()

	for {
		select {
		case <-nm.ctx.Done():
			return
		case event, ok := <-nm.eventChan:
			if !ok {
				// 通道已关闭
				return
			}
			// 这里可以添加事件广播逻辑，比如发送给订阅者、写入日志等
			nm.log.Debug(fmt.Sprintf("节点事件: %s - %s", event.Type, event.Message))
		}
	}
}

// sendEvent 发送事件到事件通道
func (nm *NodeManager) sendEvent(event *node.NodeEvent) {
	select {
	case nm.eventChan <- event:
		// 事件发送成功
	default:
		// 事件通道已满，丢弃事件并记录日志
		nm.log.Warn("事件通道已满，丢弃事件: " + event.Type)
	}
}

// copyNodeInfo 创建节点信息的副本
func (nm *NodeManager) copyNodeInfo(original *node.NodeInfo) *node.NodeInfo {
	if original == nil {
		return nil
	}

	// 创建副本
	copy := &node.NodeInfo{
		ID:        original.ID,
		Name:      original.Name,
		Address:   original.Address,
		Port:      original.Port,
		Role:      original.Role,
		Status:    original.Status,
		Version:   original.Version,
		Tags:      make([]string, len(original.Tags)),
		CreatedAt: original.CreatedAt,
		UpdatedAt: original.UpdatedAt,
		LastSeen:  original.LastSeen,
	}

	// 复制切片
	for i, tag := range original.Tags {
		copy.Tags[i] = tag
	}

	// 复制metadata
	if original.Metadata != nil {
		copy.Metadata = make(map[string]string)
		for k, v := range original.Metadata {
			copy.Metadata[k] = v
		}
	}

	// 复制capabilities
	if original.Capabilities != nil {
		copy.Capabilities = &node.NodeCapabilities{}
		*copy.Capabilities = *original.Capabilities
		if len(original.Capabilities.GPUNames) > 0 {
			copy.Capabilities.GPUNames = make([]string, len(original.Capabilities.GPUNames))
			for i, name := range original.Capabilities.GPUNames {
				copy.Capabilities.GPUNames[i] = name
			}
		}
		if len(original.Capabilities.CondaEnvs) > 0 {
			copy.Capabilities.CondaEnvs = make([]string, len(original.Capabilities.CondaEnvs))
			for i, env := range original.Capabilities.CondaEnvs {
				copy.Capabilities.CondaEnvs[i] = env
			}
		}
	}

	// 复制resources
	if original.Resources != nil {
		copy.Resources = &node.NodeResources{}
		*copy.Resources = *original.Resources
		if len(original.Resources.GPUInfo) > 0 {
			copy.Resources.GPUInfo = make([]node.GPUInfo, len(original.Resources.GPUInfo))
			for i, gpu := range original.Resources.GPUInfo {
				copy.Resources.GPUInfo[i] = gpu
			}
		}
		if len(original.Resources.LoadAverage) > 0 {
			copy.Resources.LoadAverage = make([]float64, len(original.Resources.LoadAverage))
			for i, load := range original.Resources.LoadAverage {
				copy.Resources.LoadAverage[i] = load
			}
		}
	}

	return copy
}
