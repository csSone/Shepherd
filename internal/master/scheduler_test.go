// Package master provides intelligent scheduling for model execution in distributed architecture.
package master

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/shepherd-project/shepherd/Shepherd/internal/config"
	"github.com/shepherd-project/shepherd/Shepherd/internal/logger"
	"github.com/shepherd-project/shepherd/Shepherd/internal/node"
)

// TestResourceBasedStrategy 测试基于资源的调度策略
func TestResourceBasedStrategy(t *testing.T) {
	strategy := &ResourceBasedStrategy{}

	// 创建测试节点
	nodes := createTestNodes()

	// 测试选择资源最充足的节点
	req := &ModelRequest{
		ModelName:      "test-model",
		ModelSize:      1000000000, // 1GB
		RequiredMemory: 2000000000, // 2GB
		RequireGPU:     true,
		GPUMemory:      1000000000, // 1GB
	}

	selectedNode, err := strategy.SelectNode(nodes, req)
	require.NoError(t, err)
	assert.NotNil(t, selectedNode)

	// 应该选择node3，因为它有最多的资源
	assert.Equal(t, "node3", selectedNode.ID)
}

// TestLoadBalancedStrategy 测试基于负载均衡的调度策略
func TestLoadBalancedStrategy(t *testing.T) {
	strategy := &LoadBalancedStrategy{}

	// 创建测试节点
	nodes := createTestNodes()

	// 测试选择负载最低的节点
	req := &ModelRequest{
		ModelName:      "test-model",
		ModelSize:      1000000000, // 1GB
		RequiredMemory: 2000000000, // 2GB
		RequireGPU:     true,
		GPUMemory:      1000000000, // 1GB
	}

	selectedNode, err := strategy.SelectNode(nodes, req)
	require.NoError(t, err)
	assert.NotNil(t, selectedNode)

	// 应该选择node1，因为它的负载最低
	assert.Equal(t, "node1", selectedNode.ID)
}

// TestLocalityStrategy 测试基于本地性的调度策略
func TestLocalityStrategy(t *testing.T) {
	strategy := NewLocalityStrategy()

	// 创建测试节点
	nodes := createTestNodes()

	// 更新模型缓存，模拟node1已经有该模型
	strategy.UpdateModelCache("test-model", []string{"node1"})

	// 测试选择已有模型的节点
	req := &ModelRequest{
		ModelName:      "test-model",
		ModelSize:      1000000000, // 1GB
		RequiredMemory: 2000000000, // 2GB
		RequireGPU:     true,
		GPUMemory:      1000000000, // 1GB
	}

	selectedNode, err := strategy.SelectNode(nodes, req)
	require.NoError(t, err)
	assert.NotNil(t, selectedNode)

	// 应该选择node1，因为它已经有该模型
	assert.Equal(t, "node1", selectedNode.ID)
}

// TestScheduler 测试调度器
func TestScheduler(t *testing.T) {
	// 创建测试用的logger
	log, err := logger.NewLogger(&config.LogConfig{
		Level:  "info",
		Format: "text",
		Output: "stdout",
	}, "test")
	require.NoError(t, err)

	// 创建测试用的NodeManager
	nodeManager := NewNodeManager(log)

	// 创建调度器
	scheduler := NewScheduler(nodeManager, log)

	// 创建测试节点
	nodes := createTestNodes()

	// 注册测试节点
	for _, node := range nodes {
		err := nodeManager.RegisterNode(node)
		require.NoError(t, err)
	}

	// 测试调度
	req := &ModelRequest{
		ModelName:      "test-model",
		ModelSize:      1000000000, // 1GB
		RequiredMemory: 2000000000, // 2GB
		RequireGPU:     true,
		GPUMemory:      1000000000, // 1GB
	}

	selectedNode, err := scheduler.Schedule(req)
	require.NoError(t, err)
	assert.NotNil(t, selectedNode)
	assert.Equal(t, "node1", selectedNode.ID) // 默认使用负载均衡策略，应该选择node1
}

// TestSchedulerWithCustomStrategy 测试自定义策略的调度器
func TestSchedulerWithCustomStrategy(t *testing.T) {
	// 创建测试用的logger
	log := createTestLogger(t)

	// 创建测试用的NodeManager
	nodeManager := NewNodeManager(log)

	// 创建调度器
	scheduler := NewScheduler(nodeManager, log)

	// 设置基于资源的策略
	scheduler.SetStrategy(&ResourceBasedStrategy{})

	// 创建测试节点
	nodes := createTestNodes()

	// 注册测试节点
	for _, node := range nodes {
		err := nodeManager.RegisterNode(node)
		require.NoError(t, err)
	}

	// 测试调度
	req := &ModelRequest{
		ModelName:      "test-model",
		ModelSize:      1000000000, // 1GB
		RequiredMemory: 2000000000, // 2GB
		RequireGPU:     true,
		GPUMemory:      1000000000, // 1GB
	}

	selectedNode, err := scheduler.Schedule(req)
	require.NoError(t, err)
	assert.NotNil(t, selectedNode)
	assert.Equal(t, "node3", selectedNode.ID) // 使用资源策略，应该选择node3
}

// TestSchedulerNoAvailableNodes 测试没有可用节点的情况
func TestSchedulerNoAvailableNodes(t *testing.T) {
	// 创建测试用的logger
	log := createTestLogger(t)

	// 创建测试用的NodeManager
	nodeManager := NewNodeManager(log)

	// 创建调度器
	scheduler := NewScheduler(nodeManager, log)

	// 测试调度（没有注册任何节点）
	req := &ModelRequest{
		ModelName:      "test-model",
		ModelSize:      1000000000, // 1GB
		RequiredMemory: 2000000000, // 2GB
		RequireGPU:     true,
		GPUMemory:      1000000000, // 1GB
	}

	selectedNode, err := scheduler.Schedule(req)
	assert.Error(t, err)
	assert.Nil(t, selectedNode)
	assert.Contains(t, err.Error(), "没有可用的在线节点")
}

// TestSchedulerInsufficientResources 测试资源不足的情况
func TestSchedulerInsufficientResources(t *testing.T) {
	// 创建测试用的logger
	log := createTestLogger(t)

	// 创建测试用的NodeManager
	nodeManager := NewNodeManager(log)

	// 创建调度器
	scheduler := NewScheduler(nodeManager, log)

	// 创建资源不足的测试节点
	insufficientNode := &node.NodeInfo{
		ID:      "insufficient-node",
		Name:    "Insufficient Node",
		Address: "127.0.0.1",
		Port:    8080,
		Role:    node.NodeRoleClient,
		Status:  node.NodeStatusOnline,
		Capabilities: &node.NodeCapabilities{
			GPU:      false, // 没有GPU
			CPUCount: 4,
			Memory:   4000000000, // 4GB
		},
		Resources: &node.NodeResources{
			CPUTotal:    4000,
			CPUUsed:     1000,
			MemoryTotal: 4000000000,
			MemoryUsed:  2000000000, // 已使用2GB，只剩2GB可用
		},
		Metadata:  make(map[string]string),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		LastSeen:  time.Now(),
	}

	// 注册测试节点
	err := nodeManager.RegisterNode(insufficientNode)
	require.NoError(t, err)

	// 测试调度（需要GPU和大量内存）
	req := &ModelRequest{
		ModelName:      "large-model",
		ModelSize:      10000000000, // 10GB
		RequiredMemory: 4000000000,  // 4GB
		RequireGPU:     true,
		GPUMemory:      4000000000, // 4GB
	}

	selectedNode, err := scheduler.Schedule(req)
	assert.Error(t, err)
	assert.Nil(t, selectedNode)
	assert.Contains(t, err.Error(), "节点选择失败")
}

// TestSchedulerUpdateModelCache 测试更新模型缓存
func TestSchedulerUpdateModelCache(t *testing.T) {
	// 创建测试用的logger
	log := createTestLogger(t)

	// 创建测试用的NodeManager
	nodeManager := NewNodeManager(log)

	// 创建调度器
	scheduler := NewScheduler(nodeManager, log)

	// 更新模型缓存
	scheduler.UpdateModelCache("test-model", []string{"node1", "node2"})

	// 验证缓存已更新
	localityStrategy := scheduler.localityStrategy
	require.NotNil(t, localityStrategy)

	localityStrategy.mu.RLock()
	defer localityStrategy.mu.RUnlock()

	nodeIDs, exists := localityStrategy.modelCache["test-model"]
	assert.True(t, exists)
	assert.Equal(t, []string{"node1", "node2"}, nodeIDs)
}

// TestSchedulerGetAvailableStrategies 测试获取可用策略
func TestSchedulerGetAvailableStrategies(t *testing.T) {
	// 创建测试用的logger
	log := createTestLogger(t)

	// 创建测试用的NodeManager
	nodeManager := NewNodeManager(log)

	// 创建调度器
	scheduler := NewScheduler(nodeManager, log)

	// 获取可用策略
	strategies := scheduler.GetAvailableStrategies()
	assert.Len(t, strategies, 3)
	assert.Contains(t, strategies, "ResourceBased")
	assert.Contains(t, strategies, "LoadBalanced")
	assert.Contains(t, strategies, "Locality")
}

// TestSchedulerSetAndGetStrategy 测试设置和获取策略
func TestSchedulerSetAndGetStrategy(t *testing.T) {
	// 创建测试用的logger
	log := createTestLogger(t)

	// 创建测试用的NodeManager
	nodeManager := NewNodeManager(log)

	// 创建调度器
	scheduler := NewScheduler(nodeManager, log)

	// 测试默认策略
	defaultStrategy := scheduler.GetStrategy()
	assert.NotNil(t, defaultStrategy)
	assert.Equal(t, "LoadBalanced", defaultStrategy.Name())

	// 设置新策略
	newStrategy := &ResourceBasedStrategy{}
	scheduler.SetStrategy(newStrategy)

	// 验证策略已更改
	currentStrategy := scheduler.GetStrategy()
	assert.NotNil(t, currentStrategy)
	assert.Equal(t, "ResourceBased", currentStrategy.Name())
}

// createTestLogger 创建测试用的logger
func createTestLogger(t *testing.T) *logger.Logger {
	log, err := logger.NewLogger(&config.LogConfig{
		Level:  "info",
		Format: "text",
		Output: "stdout",
	}, "test")
	require.NoError(t, err)
	return log
}

// createTestNodes 创建测试用的节点列表
func createTestNodes() []*node.NodeInfo {
	now := time.Now()

	// Node1: 低负载，中等资源
	node1 := &node.NodeInfo{
		ID:      "node1",
		Name:    "Node 1",
		Address: "127.0.0.1",
		Port:    8081,
		Role:    node.NodeRoleClient,
		Status:  node.NodeStatusOnline,
		Capabilities: &node.NodeCapabilities{
			GPU:      true,
			GPUCount: 1,
			GPUNames: []string{"NVIDIA RTX 3080"},
			CPUCount: 8,
			Memory:   16000000000, // 16GB
		},
		Resources: &node.NodeResources{
			CPUTotal:    8000,
			CPUUsed:     1000, // 低CPU使用率
			MemoryTotal: 16000000000,
			MemoryUsed:  4000000000,               // 使用4GB，还剩12GB
			LoadAverage: []float64{0.5, 0.6, 0.7}, // 低负载
			GPUInfo: []node.GPUInfo{
				{
					Index:       0,
					Name:        "NVIDIA RTX 3080",
					Vendor:      "NVIDIA",
					TotalMemory: 10000000000, // 10GB
					UsedMemory:  2000000000,  // 使用2GB
				},
			},
		},
		Metadata:  make(map[string]string),
		CreatedAt: now,
		UpdatedAt: now,
		LastSeen:  now,
	}

	// Node2: 中等负载，中等资源
	node2 := &node.NodeInfo{
		ID:      "node2",
		Name:    "Node 2",
		Address: "127.0.0.1",
		Port:    8082,
		Role:    node.NodeRoleClient,
		Status:  node.NodeStatusOnline,
		Capabilities: &node.NodeCapabilities{
			GPU:      true,
			GPUCount: 1,
			GPUNames: []string{"NVIDIA RTX 3080"},
			CPUCount: 8,
			Memory:   16000000000, // 16GB
		},
		Resources: &node.NodeResources{
			CPUTotal:    8000,
			CPUUsed:     4000, // 中等CPU使用率
			MemoryTotal: 16000000000,
			MemoryUsed:  8000000000,               // 使用8GB，还剩8GB
			LoadAverage: []float64{1.0, 1.2, 1.4}, // 中等负载
			GPUInfo: []node.GPUInfo{
				{
					Index:       0,
					Name:        "NVIDIA RTX 3080",
					Vendor:      "NVIDIA",
					TotalMemory: 10000000000, // 10GB
					UsedMemory:  5000000000,  // 使用5GB
				},
			},
		},
		Metadata:  make(map[string]string),
		CreatedAt: now,
		UpdatedAt: now,
		LastSeen:  now,
	}

	// Node3: 高资源，高负载
	node3 := &node.NodeInfo{
		ID:      "node3",
		Name:    "Node 3",
		Address: "127.0.0.1",
		Port:    8083,
		Role:    node.NodeRoleClient,
		Status:  node.NodeStatusOnline,
		Capabilities: &node.NodeCapabilities{
			GPU:      true,
			GPUCount: 2,
			GPUNames: []string{"NVIDIA RTX 3090", "NVIDIA RTX 3090"},
			CPUCount: 16,
			Memory:   32000000000, // 32GB
		},
		Resources: &node.NodeResources{
			CPUTotal:    16000,
			CPUUsed:     8000, // 高CPU使用率
			MemoryTotal: 32000000000,
			MemoryUsed:  16000000000,              // 使用16GB，还剩16GB
			LoadAverage: []float64{2.0, 2.4, 2.8}, // 高负载
			GPUInfo: []node.GPUInfo{
				{
					Index:       0,
					Name:        "NVIDIA RTX 3090",
					Vendor:      "NVIDIA",
					TotalMemory: 12000000000, // 12GB
					UsedMemory:  6000000000,  // 使用6GB
				},
				{
					Index:       1,
					Name:        "NVIDIA RTX 3090",
					Vendor:      "NVIDIA",
					TotalMemory: 12000000000, // 12GB
					UsedMemory:  6000000000,  // 使用6GB
				},
			},
		},
		Metadata:  make(map[string]string),
		CreatedAt: now,
		UpdatedAt: now,
		LastSeen:  now,
	}

	// Node4: 无GPU节点
	node4 := &node.NodeInfo{
		ID:      "node4",
		Name:    "Node 4",
		Address: "127.0.0.1",
		Port:    8084,
		Role:    node.NodeRoleClient,
		Status:  node.NodeStatusOnline,
		Capabilities: &node.NodeCapabilities{
			GPU:      false, // 无GPU
			GPUCount: 0,
			CPUCount: 8,
			Memory:   16000000000, // 16GB
		},
		Resources: &node.NodeResources{
			CPUTotal:    8000,
			CPUUsed:     2000,
			MemoryTotal: 16000000000,
			MemoryUsed:  4000000000,
			LoadAverage: []float64{0.8, 0.9, 1.0},
		},
		Metadata:  make(map[string]string),
		CreatedAt: now,
		UpdatedAt: now,
		LastSeen:  now,
	}

	return []*node.NodeInfo{node1, node2, node3, node4}
}
