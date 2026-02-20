// Package node provides distributed node management implementation.
package node

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// Node represents a distributed node in the Shepherd system
type Node struct {
	// Basic information
	id       string
	name     string
	role     NodeRole
	status   NodeStatus
	address  string
	port     int
	version  string
	tags     []string
	metadata map[string]string

	// Capabilities and resources
	capabilities *NodeCapabilities
	resources    *NodeResources
	config       *NodeConfig

	// Runtime state
	createdAt time.Time
	updatedAt time.Time
	lastSeen  time.Time
	startedAt *time.Time
	stoppedAt *time.Time

	// Concurrency control
	mu      sync.RWMutex
	running bool
	ctx     context.Context
	cancel  context.CancelFunc

	// Subsystems
	resource  *ResourceMonitor
	heartbeat *HeartbeatManager
	commands  *CommandExecutor
}

// NewNode creates a new Node instance
func NewNode(config *NodeConfig) (*Node, error) {
	if config == nil {
		return nil, fmt.Errorf("配置不能为空")
	}

	if config.ID == "" {
		return nil, fmt.Errorf("节点ID不能为空")
	}

	if config.Role == "" {
		config.Role = NodeRoleStandalone
	}

	ctx, cancel := context.WithCancel(context.Background())

	node := &Node{
		id:        config.ID,
		name:      config.Name,
		role:      config.Role,
		status:    NodeStatusOffline,
		address:   config.Address,
		port:      config.Port,
		version:   "1.0.0", // TODO: 从构建信息获取
		tags:      make([]string, 0),
		metadata:  make(map[string]string),
		config:    config,
		createdAt: time.Now(),
		updatedAt: time.Now(),
		ctx:       ctx,
		cancel:    cancel,
	}

	// Initialize capabilities
	node.capabilities = &NodeCapabilities{
		SupportsLlama:  true,
		SupportsPython: true,
		GPU:            false,
		CPUCount:       1,
		Memory:         1024 * 1024 * 1024, // 1GB default
	}

	// Initialize resources
	node.resources = &NodeResources{
		CPUTotal:    1000,                    // 1 core in millicores
		MemoryTotal: 1024 * 1024 * 1024,      // 1GB
		DiskTotal:   10 * 1024 * 1024 * 1024, // 10GB
		GPUInfo:     make([]GPUInfo, 0),
		LoadAverage: make([]float64, 3),
	}

	return node, nil
}

// Start 启动节点
func (n *Node) Start() error {
	n.mu.Lock()
	defer n.mu.Unlock()

	if n.running {
		return fmt.Errorf("节点已经在运行")
	}

	n.status = NodeStatusOnline
	n.running = true
	now := time.Now()
	n.startedAt = &now
	n.updatedAt = now
	n.lastSeen = now

	// Initialize subsystems
	if err := n.initSubsystems(); err != nil {
		n.status = NodeStatusError
		n.running = false
		return fmt.Errorf("初始化子系统失败: %w", err)
	}

	// Start subsystems
	if err := n.startSubsystems(); err != nil {
		n.status = NodeStatusError
		n.running = false
		return fmt.Errorf("启动子系统失败: %w", err)
	}

	return nil
}

// Stop 停止节点
func (n *Node) Stop() error {
	n.mu.Lock()
	defer n.mu.Unlock()

	if !n.running {
		return nil
	}

	n.status = NodeStatusOffline
	n.running = false
	now := time.Now()
	n.stoppedAt = &now
	n.updatedAt = now

	// Stop subsystems
	n.stopSubsystems()

	// Cancel context
	n.cancel()

	return nil
}

// GetID 获取节点ID
func (n *Node) GetID() string {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return n.id
}

// GetName 获取节点名称
func (n *Node) GetName() string {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return n.name
}

// SetName 设置节点名称
func (n *Node) SetName(name string) {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.name = name
	n.updatedAt = time.Now()
}

// GetRole 获取节点角色
func (n *Node) GetRole() NodeRole {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return n.role
}

// SetRole 设置节点角色
func (n *Node) SetRole(role NodeRole) error {
	n.mu.Lock()
	defer n.mu.Unlock()

	if n.running {
		return fmt.Errorf("节点运行时不能更改角色")
	}

	n.role = role
	n.updatedAt = time.Now()
	return nil
}

// GetStatus 获取节点状态
func (n *Node) GetStatus() NodeStatus {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return n.status
}

// SetStatus 设置节点状态
func (n *Node) SetStatus(status NodeStatus) {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.status = status
	n.updatedAt = time.Now()
}

// GetAddress 获取节点地址
func (n *Node) GetAddress() string {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return n.address
}

// GetPort 获取节点端口
func (n *Node) GetPort() int {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return n.port
}

// GetVersion 获取节点版本
func (n *Node) GetVersion() string {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return n.version
}

// GetTags 获取节点标签
func (n *Node) GetTags() []string {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return append([]string{}, n.tags...)
}

// SetTags 设置节点标签
func (n *Node) SetTags(tags []string) {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.tags = append([]string{}, tags...)
	n.updatedAt = time.Now()
}

// AddTag 添加节点标签
func (n *Node) AddTag(tag string) {
	n.mu.Lock()
	defer n.mu.Unlock()

	for _, existing := range n.tags {
		if existing == tag {
			return // 已存在
		}
	}

	n.tags = append(n.tags, tag)
	n.updatedAt = time.Now()
}

// RemoveTag 移除节点标签
func (n *Node) RemoveTag(tag string) {
	n.mu.Lock()
	defer n.mu.Unlock()

	for i, existing := range n.tags {
		if existing == tag {
			n.tags = append(n.tags[:i], n.tags[i+1:]...)
			break
		}
	}

	n.updatedAt = time.Now()
}

// GetMetadata 获取节点元数据
func (n *Node) GetMetadata() map[string]string {
	n.mu.RLock()
	defer n.mu.RUnlock()

	metadata := make(map[string]string)
	for k, v := range n.metadata {
		metadata[k] = v
	}
	return metadata
}

// SetMetadata 设置节点元数据
func (n *Node) SetMetadata(metadata map[string]string) {
	n.mu.Lock()
	defer n.mu.Unlock()

	n.metadata = make(map[string]string)
	for k, v := range metadata {
		n.metadata[k] = v
	}
	n.updatedAt = time.Now()
}

// GetCapabilities 获取节点能力
func (n *Node) GetCapabilities() *NodeCapabilities {
	n.mu.RLock()
	defer n.mu.RUnlock()

	if n.capabilities == nil {
		return nil
	}

	// Return a copy
	cap := *n.capabilities
	return &cap
}

// SetCapabilities 设置节点能力
func (n *Node) SetCapabilities(capabilities *NodeCapabilities) {
	n.mu.Lock()
	defer n.mu.Unlock()

	if capabilities == nil {
		n.capabilities = nil
	} else {
		// Create a copy
		n.capabilities = &NodeCapabilities{}
		*n.capabilities = *capabilities
	}
	n.updatedAt = time.Now()
}

// GetResources 获取节点资源信息
func (n *Node) GetResources() *NodeResources {
	n.mu.RLock()
	defer n.mu.RUnlock()

	if n.resources == nil {
		return nil
	}

	// Return a copy
	res := *n.resources
	return &res
}

// SetResources 设置节点资源信息
func (n *Node) SetResources(resources *NodeResources) {
	n.mu.Lock()
	defer n.mu.Unlock()

	if resources == nil {
		n.resources = nil
	} else {
		// Create a copy
		n.resources = &NodeResources{}
		*n.resources = *resources
	}
	n.updatedAt = time.Now()
}

// GetConfig 获取节点配置
func (n *Node) GetConfig() *NodeConfig {
	n.mu.RLock()
	defer n.mu.RUnlock()

	if n.config == nil {
		return nil
	}

	// Return a copy
	cfg := *n.config
	return &cfg
}

// IsRunning 检查节点是否正在运行
func (n *Node) IsRunning() bool {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return n.running
}

// GetCreatedAt 获取节点创建时间
func (n *Node) GetCreatedAt() time.Time {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return n.createdAt
}

// GetUpdatedAt 获取节点更新时间
func (n *Node) GetUpdatedAt() time.Time {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return n.updatedAt
}

// GetLastSeen 获取节点最后活跃时间
func (n *Node) GetLastSeen() time.Time {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return n.lastSeen
}

// UpdateLastSeen 更新节点最后活跃时间
func (n *Node) UpdateLastSeen() {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.lastSeen = time.Now()
}

// GetUptime 获取节点运行时长
func (n *Node) GetUptime() time.Duration {
	n.mu.RLock()
	defer n.mu.RUnlock()

	if n.startedAt == nil {
		return 0
	}

	if n.stoppedAt != nil {
		return n.stoppedAt.Sub(*n.startedAt)
	}

	return time.Since(*n.startedAt)
}

// ToInfo 转换为NodeInfo
func (n *Node) ToInfo() *NodeInfo {
	n.mu.RLock()
	defer n.mu.RUnlock()

	info := &NodeInfo{
		ID:           n.id,
		Name:         n.name,
		Address:      n.address,
		Port:         n.port,
		Role:         n.role,
		Status:       n.status,
		Version:      n.version,
		Tags:         append([]string{}, n.tags...),
		Capabilities: n.GetCapabilities(),
		Resources:    n.GetResources(),
		CreatedAt:    n.createdAt,
		UpdatedAt:    n.updatedAt,
		LastSeen:     n.lastSeen,
	}

	// Create metadata copy
	info.Metadata = make(map[string]string)
	for k, v := range n.metadata {
		info.Metadata[k] = v
	}

	return info
}

// String 返回节点的字符串表示
func (n *Node) String() string {
	return fmt.Sprintf("Node{id:%s, name:%s, role:%s, status:%s, address:%s:%d}",
		n.GetID(), n.GetName(), n.GetRole(), n.GetStatus(), n.GetAddress(), n.GetPort())
}

// initSubsystems 初始化子系统
func (n *Node) initSubsystems() error {
	// 初始化资源监控器
	resourceConfig := &ResourceMonitorConfig{
		Interval: 5 * time.Second,
		Callback: func(resources *NodeResources) {
			// 当资源信息更新时，自动更新节点的资源信息
			n.SetResources(resources)
		},
	}
	n.resource = NewResourceMonitor(resourceConfig)

	// TODO: 实现其他子系统初始化
	// n.heartbeat = NewHeartbeatManager(n)
	// n.health = NewHealthManager(n)
	// n.commands = NewCommandManager(n)
	// n.events = NewEventManager(n)
	// n.metrics = NewMetricsManager(n)
	return nil
}

// startSubsystems 启动子系统
func (n *Node) startSubsystems() error {
	// 启动资源监控器
	if n.resource != nil {
		if err := n.resource.Start(); err != nil {
			return fmt.Errorf("启动资源监控器失败: %w", err)
		}
	}

	// TODO: 实现其他子系统启动
	// if n.heartbeat != nil {
	//     if err := n.heartbeat.Start(); err != nil {
	//         return err
	//     }
	// }
	// ... 其他子系统
	return nil
}

// stopSubsystems 停止子系统
func (n *Node) stopSubsystems() {
	// 停止资源监控器
	if n.resource != nil {
		if err := n.resource.Stop(); err != nil {
			// 停止失败只记录日志，不影响其他子系统停止
			// TODO: 添加日志记录
		}
	}

	// TODO: 实现其他子系统停止
	// if n.heartbeat != nil {
	//     n.heartbeat.Stop()
	// }
	// ... 其他子系统
}

// Context 获取节点上下文
func (n *Node) Context() context.Context {
	return n.ctx
}

// GetResourceMonitor 获取资源监控器
func (n *Node) GetResourceMonitor() *ResourceMonitor {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return n.resource
}

// GetResourceSnapshot 获取当前资源快照（便捷方法）
func (n *Node) GetResourceSnapshot() *NodeResources {
	n.mu.RLock()
	defer n.mu.RUnlock()

	if n.resource == nil {
		return n.resources
	}

	return n.resource.GetSnapshot()
}

// GetGPUInfo 获取GPU信息（便捷方法）
func (n *Node) GetGPUInfo() []GPUInfo {
	n.mu.RLock()
	defer n.mu.RUnlock()

	if n.resource == nil {
		return make([]GPUInfo, 0)
	}

	return n.resource.GetGPUInfo()
}

// GetLlamacppInfo 获取llama.cpp信息（便捷方法）
func (n *Node) GetLlamacppInfo() *LlamacppInfo {
	n.mu.RLock()
	defer n.mu.RUnlock()

	if n.resource == nil {
		return nil
	}

	return n.resource.GetLlamacppInfo()
}
