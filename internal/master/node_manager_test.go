package master

import (
	"testing"
	"time"

	"github.com/shepherd-project/shepherd/Shepherd/internal/config"
	"github.com/shepherd-project/shepherd/Shepherd/internal/logger"
	"github.com/shepherd-project/shepherd/Shepherd/internal/node"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewNodeManager 测试创建节点管理器
func TestNewNodeManager(t *testing.T) {
	log, _ := logger.NewLogger(&config.LogConfig{
		Level:  "info",
		Format: "text",
		Output: "stdout",
	}, "test")
	nm := NewNodeManager(log)

	assert.NotNil(t, nm)
	assert.Equal(t, 15*time.Second, nm.timeout)
	assert.Equal(t, 5*time.Second, nm.checkInterval)
	assert.NotNil(t, nm.eventChan)
	assert.False(t, nm.running)
}

// TestNodeManager_StartStop 测试启动和停止节点管理器
func TestNodeManager_StartStop(t *testing.T) {
	log, _ := logger.NewLogger(&config.LogConfig{
		Level:  "info",
		Format: "text",
		Output: "stdout",
	}, "test")
	nm := NewNodeManager(log)

	// 测试启动
	nm.Start()
	assert.True(t, nm.running)

	// 测试重复启动
	nm.Start()
	assert.True(t, nm.running) // 应该仍然是true

	// 测试停止
	nm.Stop()
	assert.False(t, nm.running)

	// 测试重复停止
	nm.Stop()
	assert.False(t, nm.running) // 应该仍然是false
}

// TestNodeManager_RegisterNode 测试节点注册
func TestNodeManager_RegisterNode(t *testing.T) {
	log, _ := logger.NewLogger(&config.LogConfig{
		Level:  "info",
		Format: "text",
		Output: "stdout",
	}, "test")
	nm := NewNodeManager(log)

	// 创建测试节点
	nodeInfo := &node.NodeInfo{
		ID:      "test-node-1",
		Name:    "Test Node 1",
		Address: "192.168.1.100",
		Port:    8080,
		Role:    node.NodeRoleClient,
		Status:  node.NodeStatusOffline,
		Version: "1.0.0",
		Tags:    []string{"test", "client"},
		Metadata: map[string]string{
			"region": "us-west-1",
		},
		Capabilities: &node.NodeCapabilities{
			GPU:            true,
			GPUCount:       2,
			CPUCount:       8,
			Memory:         16 * 1024 * 1024 * 1024, // 16GB
			SupportsLlama:  true,
			SupportsPython: true,
		},
	}

	// 注册新节点
	err := nm.RegisterNode(nodeInfo)
	require.NoError(t, err)

	// 验证节点已注册
	retrieved, err := nm.GetNode("test-node-1")
	require.NoError(t, err)
	assert.Equal(t, "test-node-1", retrieved.ID)
	assert.Equal(t, "Test Node 1", retrieved.Name)
	assert.Equal(t, "192.168.1.100", retrieved.Address)
	assert.Equal(t, 8080, retrieved.Port)
	assert.Equal(t, node.NodeRoleClient, retrieved.Role)
	assert.Equal(t, node.NodeStatusOnline, retrieved.Status) // 应该被设置为在线
	assert.Equal(t, "1.0.0", retrieved.Version)
	assert.Equal(t, []string{"test", "client"}, retrieved.Tags)
	assert.Equal(t, "us-west-1", retrieved.Metadata["region"])
	assert.True(t, retrieved.Capabilities.GPU)
	assert.Equal(t, 2, retrieved.Capabilities.GPUCount)

	// 测试重复注册（更新现有节点）
	updatedNodeInfo := &node.NodeInfo{
		ID:      "test-node-1",
		Name:    "Test Node 1 Updated",
		Address: "192.168.1.101",
		Port:    8081,
		Role:    node.NodeRoleClient,
		Status:  node.NodeStatusOffline,
		Version: "1.0.1",
	}

	err = nm.RegisterNode(updatedNodeInfo)
	require.NoError(t, err)

	// 验证节点信息已更新
	retrieved, err = nm.GetNode("test-node-1")
	require.NoError(t, err)
	assert.Equal(t, "Test Node 1 Updated", retrieved.Name)
	assert.Equal(t, "192.168.1.101", retrieved.Address)
	assert.Equal(t, 8081, retrieved.Port)
	assert.Equal(t, "1.0.1", retrieved.Version)
	assert.Equal(t, node.NodeStatusOnline, retrieved.Status) // 应该仍然是在线
}

// TestNodeManager_RegisterNodeInvalidInput 测试无效输入的节点注册
func TestNodeManager_RegisterNodeInvalidInput(t *testing.T) {
	log := func() *logger.Logger {
		log, _ := logger.NewLogger(&config.LogConfig{
			Level:  "info",
			Format: "text",
			Output: "stdout",
		}, "test")
		return log
	}()
	nm := NewNodeManager(log)

	// 测试nil节点
	err := nm.RegisterNode(nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "节点信息不能为空")

	// 测试空ID节点
	nodeInfo := &node.NodeInfo{
		Name: "Test Node",
	}
	err = nm.RegisterNode(nodeInfo)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "节点ID不能为空")
}

// TestNodeManager_HandleHeartbeat 测试心跳处理
func TestNodeManager_HandleHeartbeat(t *testing.T) {
	log := func() *logger.Logger {
		log, _ := logger.NewLogger(&config.LogConfig{
			Level:  "info",
			Format: "text",
			Output: "stdout",
		}, "test")
		return log
	}()
	nm := NewNodeManager(log)

	// 注册测试节点
	nodeInfo := &node.NodeInfo{
		ID:     "test-node-1",
		Name:   "Test Node 1",
		Status: node.NodeStatusOnline,
	}
	err := nm.RegisterNode(nodeInfo)
	require.NoError(t, err)

	// 创建心跳消息
	heartbeat := &node.HeartbeatMessage{
		NodeID:    "test-node-1",
		Timestamp: time.Now(),
		Status:    node.NodeStatusBusy,
		Role:      node.NodeRoleClient,
		Resources: &node.NodeResources{
			CPUUsed:  500,
			CPUTotal: 1000,
		},
		Capabilities: &node.NodeCapabilities{
			GPU: true,
		},
		Sequence: 1,
	}

	// 处理心跳
	err = nm.HandleHeartbeat("test-node-1", heartbeat)
	require.NoError(t, err)

	// 验证节点信息已更新
	retrieved, err := nm.GetNode("test-node-1")
	require.NoError(t, err)
	assert.Equal(t, node.NodeStatusBusy, retrieved.Status)
	assert.Equal(t, int64(500), retrieved.Resources.CPUUsed)
	assert.True(t, retrieved.Capabilities.GPU)

	// 测试不存在的节点
	nonExistentHeartbeat := &node.HeartbeatMessage{
		NodeID:    "non-existent-node",
		Timestamp: time.Now(),
		Status:    node.NodeStatusOnline,
		Role:      node.NodeRoleClient,
		Sequence:  1,
	}
	err = nm.HandleHeartbeat("non-existent-node", nonExistentHeartbeat)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "节点不存在")

	// 测试nil心跳
	err = nm.HandleHeartbeat("test-node-1", nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "心跳消息不能为空")

	// 测试节点ID不匹配
	heartbeat.NodeID = "different-node"
	err = nm.HandleHeartbeat("test-node-1", heartbeat)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "心跳消息节点ID不匹配")
}

// TestNodeManager_CheckTimeouts 测试超时检查
func TestNodeManager_CheckTimeouts(t *testing.T) {
	log := func() *logger.Logger {
		log, _ := logger.NewLogger(&config.LogConfig{
			Level:  "info",
			Format: "text",
			Output: "stdout",
		}, "test")
		return log
	}()
	nm := NewNodeManager(log)

	// 注册测试节点
	nodeInfo1 := &node.NodeInfo{
		ID:     "test-node-1",
		Name:   "Test Node 1",
		Status: node.NodeStatusOnline,
	}
	nodeInfo2 := &node.NodeInfo{
		ID:     "test-node-2",
		Name:   "Test Node 2",
		Status: node.NodeStatusOnline,
	}

	err := nm.RegisterNode(nodeInfo1)
	require.NoError(t, err)
	err = nm.RegisterNode(nodeInfo2)
	require.NoError(t, err)

	// 手动设置LastSeen时间
	nodeInfo1, _ = nm.GetNode("test-node-1")
	nodeInfo2, _ = nm.GetNode("test-node-2")

	// 直接修改节点的LastSeen（通过修改实现，仅用于测试）
	nm.mu.Lock()
	nm.nodes["test-node-1"].LastSeen = time.Now().Add(-10 * time.Second) // 10秒前
	nm.nodes["test-node-2"].LastSeen = time.Now().Add(-20 * time.Second) // 20秒前（超过15秒超时）
	nm.mu.Unlock()

	// 执行超时检查
	nm.CheckTimeouts()

	// 验证状态
	retrieved1, _ := nm.GetNode("test-node-1")
	assert.Equal(t, node.NodeStatusOnline, retrieved1.Status) // 10秒前，应该仍是在线

	retrieved2, _ := nm.GetNode("test-node-2")
	assert.Equal(t, node.NodeStatusOffline, retrieved2.Status) // 20秒前，应该被标记为离线
}

// TestNodeManager_GetNode 测试获取单个节点
func TestNodeManager_GetNode(t *testing.T) {
	log := func() *logger.Logger {
		log, _ := logger.NewLogger(&config.LogConfig{
			Level:  "info",
			Format: "text",
			Output: "stdout",
		}, "test")
		return log
	}()
	nm := NewNodeManager(log)

	// 测试获取不存在的节点
	_, err := nm.GetNode("non-existent-node")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "节点不存在")

	// 注册测试节点
	nodeInfo := &node.NodeInfo{
		ID:   "test-node-1",
		Name: "Test Node 1",
	}
	err = nm.RegisterNode(nodeInfo)
	require.NoError(t, err)

	// 获取节点
	retrieved, err := nm.GetNode("test-node-1")
	require.NoError(t, err)
	assert.Equal(t, "test-node-1", retrieved.ID)
	assert.Equal(t, "Test Node 1", retrieved.Name)

	// 验证返回的是副本（修改返回值不会影响原值）
	retrieved.Name = "Modified Name"
	original, _ := nm.GetNode("test-node-1")
	assert.Equal(t, "Test Node 1", original.Name) // 原值不应该改变
}

// TestNodeManager_ListNodes 测试获取所有节点列表
func TestNodeManager_ListNodes(t *testing.T) {
	log := func() *logger.Logger {
		log, _ := logger.NewLogger(&config.LogConfig{
			Level:  "info",
			Format: "text",
			Output: "stdout",
		}, "test")
		return log
	}()
	nm := NewNodeManager(log)

	// 初始状态应该为空
	nodes := nm.ListNodes()
	assert.Len(t, nodes, 0)

	// 注册测试节点
	nodeInfo1 := &node.NodeInfo{ID: "test-node-1", Name: "Test Node 1"}
	nodeInfo2 := &node.NodeInfo{ID: "test-node-2", Name: "Test Node 2"}

	err := nm.RegisterNode(nodeInfo1)
	require.NoError(t, err)
	err = nm.RegisterNode(nodeInfo2)
	require.NoError(t, err)

	// 获取节点列表
	nodes = nm.ListNodes()
	assert.Len(t, nodes, 2)

	// 验证返回的是副本
	nodes[0].Name = "Modified Name"
	original, _ := nm.GetNode("test-node-1")
	assert.NotEqual(t, "Modified Name", original.Name) // 原值不应该改变
}

// TestNodeManager_ListOnlineNodes 测试获取在线节点列表
func TestNodeManager_ListOnlineNodes(t *testing.T) {
	log := func() *logger.Logger {
		log, _ := logger.NewLogger(&config.LogConfig{
			Level:  "info",
			Format: "text",
			Output: "stdout",
		}, "test")
		return log
	}()
	nm := NewNodeManager(log)

	// 初始状态应该为空
	nodes := nm.ListOnlineNodes()
	assert.Len(t, nodes, 0)

	// 注册测试节点
	nodeInfo1 := &node.NodeInfo{ID: "test-node-1", Name: "Test Node 1"}
	nodeInfo2 := &node.NodeInfo{ID: "test-node-2", Name: "Test Node 2"}

	err := nm.RegisterNode(nodeInfo1)
	require.NoError(t, err)
	err = nm.RegisterNode(nodeInfo2)
	require.NoError(t, err)

	// 初始状态都应该是在线
	nodes = nm.ListOnlineNodes()
	assert.Len(t, nodes, 2)

	// 将一个节点设置为离线
	err = nm.UpdateNodeStatus("test-node-1", node.NodeStatusOffline)
	require.NoError(t, err)

	// 验证只有在线节点被返回
	nodes = nm.ListOnlineNodes()
	assert.Len(t, nodes, 1)
	assert.Equal(t, "test-node-2", nodes[0].ID)
}

// TestNodeManager_UpdateNodeStatus 测试更新节点状态
func TestNodeManager_UpdateNodeStatus(t *testing.T) {
	log := func() *logger.Logger {
		log, _ := logger.NewLogger(&config.LogConfig{
			Level:  "info",
			Format: "text",
			Output: "stdout",
		}, "test")
		return log
	}()
	nm := NewNodeManager(log)

	// 注册测试节点
	nodeInfo := &node.NodeInfo{
		ID:     "test-node-1",
		Name:   "Test Node 1",
		Status: node.NodeStatusOnline,
	}
	err := nm.RegisterNode(nodeInfo)
	require.NoError(t, err)

	// 更新节点状态
	err = nm.UpdateNodeStatus("test-node-1", node.NodeStatusBusy)
	require.NoError(t, err)

	// 验证状态已更新
	retrieved, err := nm.GetNode("test-node-1")
	require.NoError(t, err)
	assert.Equal(t, node.NodeStatusBusy, retrieved.Status)

	// 测试更新为相同状态（不应该触发事件）
	err = nm.UpdateNodeStatus("test-node-1", node.NodeStatusBusy)
	require.NoError(t, err)

	// 测试更新不存在的节点
	err = nm.UpdateNodeStatus("non-existent-node", node.NodeStatusOffline)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "节点不存在")
}

// TestNodeManager_GetEventChannel 测试获取事件通道
func TestNodeManager_GetEventChannel(t *testing.T) {
	log := func() *logger.Logger {
		log, _ := logger.NewLogger(&config.LogConfig{
			Level:  "info",
			Format: "text",
			Output: "stdout",
		}, "test")
		return log
	}()
	nm := NewNodeManager(log)

	// 获取事件通道
	eventChan := nm.GetEventChannel()
	assert.NotNil(t, eventChan)

	// 启动节点管理器
	nm.Start()
	defer nm.Stop()

	// 注册节点应该产生事件
	nodeInfo := &node.NodeInfo{
		ID:   "test-node-1",
		Name: "Test Node 1",
	}
	err := nm.RegisterNode(nodeInfo)
	require.NoError(t, err)

	// 检查事件（有超时保护）
	select {
	case event := <-eventChan:
		assert.Equal(t, "node_register", event.Type)
		assert.Equal(t, "test-node-1", event.NodeID)
	case <-time.After(100 * time.Millisecond):
		t.Fatal("未收到预期的事件")
	}
}

// TestNodeManager_GetNodeCount 测试获取节点统计
func TestNodeManager_GetNodeCount(t *testing.T) {
	log := func() *logger.Logger {
		log, _ := logger.NewLogger(&config.LogConfig{
			Level:  "info",
			Format: "text",
			Output: "stdout",
		}, "test")
		return log
	}()
	nm := NewNodeManager(log)

	// 初始状态应该为0
	total, online, offline, busy := nm.GetNodeCount()
	assert.Equal(t, 0, total)
	assert.Equal(t, 0, online)
	assert.Equal(t, 0, offline)
	assert.Equal(t, 0, busy)

	// 注册测试节点
	nodeInfo1 := &node.NodeInfo{ID: "test-node-1", Name: "Test Node 1"}
	nodeInfo2 := &node.NodeInfo{ID: "test-node-2", Name: "Test Node 2"}
	nodeInfo3 := &node.NodeInfo{ID: "test-node-3", Name: "Test Node 3"}

	err := nm.RegisterNode(nodeInfo1)
	require.NoError(t, err)
	err = nm.RegisterNode(nodeInfo2)
	require.NoError(t, err)
	err = nm.RegisterNode(nodeInfo3)
	require.NoError(t, err)

	// 初始状态都应该是在线
	total, online, offline, busy = nm.GetNodeCount()
	assert.Equal(t, 3, total)
	assert.Equal(t, 3, online)
	assert.Equal(t, 0, offline)
	assert.Equal(t, 0, busy)

	// 将一些节点设置为不同状态
	err = nm.UpdateNodeStatus("test-node-1", node.NodeStatusOffline)
	require.NoError(t, err)
	err = nm.UpdateNodeStatus("test-node-2", node.NodeStatusBusy)
	require.NoError(t, err)

	// 验证统计
	total, online, offline, busy = nm.GetNodeCount()
	assert.Equal(t, 3, total)
	assert.Equal(t, 1, online)
	assert.Equal(t, 1, offline)
	assert.Equal(t, 1, busy)
}

// TestNodeManager_EventBroadcast 测试事件广播
func TestNodeManager_EventBroadcast(t *testing.T) {
	log := func() *logger.Logger {
		log, _ := logger.NewLogger(&config.LogConfig{
			Level:  "info",
			Format: "text",
			Output: "stdout",
		}, "test")
		return log
	}()
	nm := NewNodeManager(log)

	// 启动节点管理器
	nm.Start()
	defer nm.Stop()

	// 获取事件通道
	eventChan := nm.GetEventChannel()

	// 注册节点
	nodeInfo := &node.NodeInfo{
		ID:   "test-node-1",
		Name: "Test Node 1",
	}
	err := nm.RegisterNode(nodeInfo)
	require.NoError(t, err)

	// 应该收到注册事件
	select {
	case event := <-eventChan:
		assert.Equal(t, "node_register", event.Type)
		assert.Equal(t, "test-node-1", event.NodeID)
	case <-time.After(100 * time.Millisecond):
		t.Fatal("未收到注册事件")
	}

	// 更新节点状态
	err = nm.UpdateNodeStatus("test-node-1", node.NodeStatusBusy)
	require.NoError(t, err)

	// 应该收到状态变更事件
	select {
	case event := <-eventChan:
		assert.Equal(t, "status_change", event.Type)
		assert.Equal(t, "test-node-1", event.NodeID)
	case <-time.After(100 * time.Millisecond):
		t.Fatal("未收到状态变更事件")
	}
}

// TestNodeManager_CopyNodeInfo 测试节点信息深拷贝
func TestNodeManager_CopyNodeInfo(t *testing.T) {
	log := func() *logger.Logger {
		log, _ := logger.NewLogger(&config.LogConfig{
			Level:  "info",
			Format: "text",
			Output: "stdout",
		}, "test")
		return log
	}()
	nm := NewNodeManager(log)

	// 创建复杂节点信息
	original := &node.NodeInfo{
		ID:      "test-node-1",
		Name:    "Test Node 1",
		Address: "192.168.1.100",
		Port:    8080,
		Role:    node.NodeRoleClient,
		Status:  node.NodeStatusOnline,
		Version: "1.0.0",
		Tags:    []string{"test", "client"},
		Metadata: map[string]string{
			"region": "us-west-1",
		},
		Capabilities: &node.NodeCapabilities{
			GPU:            true,
			GPUCount:       2,
			GPUNames:       []string{"RTX 3080", "RTX 3090"},
			CPUCount:       8,
			Memory:         16 * 1024 * 1024 * 1024, // 16GB
			SupportsLlama:  true,
			SupportsPython: true,
			CondaEnvs:      []string{"base", "pytorch"},
		},
		Resources: &node.NodeResources{
			CPUUsed:     500,
			CPUTotal:    1000,
			MemoryUsed:  8 * 1024 * 1024 * 1024,  // 8GB
			MemoryTotal: 16 * 1024 * 1024 * 1024, // 16GB
			GPUInfo: []node.GPUInfo{
				{
					Index:         0,
					Name:          "RTX 3080",
					TotalMemory:   10 * 1024 * 1024 * 1024, // 10GB
					UsedMemory:    5 * 1024 * 1024 * 1024,  // 5GB
					Temperature:   65.0,
					Utilization:   80.0,
					PowerUsage:    250.0,
					DriverVersion: "470.82.01",
				},
			},
			LoadAverage: []float64{1.0, 0.8, 0.6},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		LastSeen:  time.Now(),
	}

	// 注册节点
	err := nm.RegisterNode(original)
	require.NoError(t, err)

	// 获取副本
	copy, err := nm.GetNode("test-node-1")
	require.NoError(t, err)

	// 验证内容一致
	assert.Equal(t, original.ID, copy.ID)
	assert.Equal(t, original.Name, copy.Name)
	assert.Equal(t, original.Address, copy.Address)
	assert.Equal(t, original.Port, copy.Port)
	assert.Equal(t, original.Role, copy.Role)
	assert.Equal(t, original.Status, copy.Status)
	assert.Equal(t, original.Version, copy.Version)
	assert.Equal(t, original.Tags, copy.Tags)
	assert.Equal(t, original.Metadata, copy.Metadata)
	assert.Equal(t, original.Capabilities.GPU, copy.Capabilities.GPU)
	assert.Equal(t, original.Capabilities.GPUCount, copy.Capabilities.GPUCount)
	assert.Equal(t, original.Capabilities.GPUNames, copy.Capabilities.GPUNames)
	assert.Equal(t, original.Resources.CPUUsed, copy.Resources.CPUUsed)
	assert.Equal(t, len(original.Resources.GPUInfo), len(copy.Resources.GPUInfo))
	assert.Equal(t, original.Resources.LoadAverage, copy.Resources.LoadAverage)

	// 修改副本不应该影响原值
	copy.Name = "Modified Name"
	copy.Tags[0] = "modified"
	copy.Metadata["region"] = "modified"
	copy.Capabilities.GPU = false
	copy.Capabilities.GPUNames[0] = "modified"
	copy.Resources.CPUUsed = 999
	copy.Resources.GPUInfo[0].Name = "modified"
	copy.Resources.LoadAverage[0] = 9.9

	// 验证原值未改变
	originalAgain, _ := nm.GetNode("test-node-1")
	assert.Equal(t, "Test Node 1", originalAgain.Name)
	assert.Equal(t, "test", originalAgain.Tags[0])
	assert.Equal(t, "us-west-1", originalAgain.Metadata["region"])
	assert.True(t, originalAgain.Capabilities.GPU)
	assert.Equal(t, "RTX 3080", originalAgain.Capabilities.GPUNames[0])
	assert.Equal(t, int64(500), originalAgain.Resources.CPUUsed)
	assert.Equal(t, "RTX 3080", originalAgain.Resources.GPUInfo[0].Name)
	assert.Equal(t, 1.0, originalAgain.Resources.LoadAverage[0])
}
