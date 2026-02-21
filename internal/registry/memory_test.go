// Package registry provides client registry tests
package registry

import (
	"fmt"
	"testing"
	"time"

	"github.com/shepherd-project/shepherd/Shepherd/internal/config"
	"github.com/shepherd-project/shepherd/Shepherd/internal/logger"
	"github.com/shepherd-project/shepherd/Shepherd/internal/node"
)

// newTestLogger 创建一个用于测试的日志记录器
func newTestLogger(t *testing.T) *logger.Logger {
	log, err := logger.NewLogger(&config.LogConfig{
		Level:  "info",
		Output: "stdout",
	}, "test")
	if err != nil {
		t.Fatalf("创建日志记录器失败: %v", err)
	}
	return log
}

// TestMemoryClientRegistry_Register tests registering clients
func TestMemoryClientRegistry_Register(t *testing.T) {
	reg := NewMemoryClientRegistry(newTestLogger(t), 5*time.Minute, 1*time.Minute)

	// Test registering a new client
	info := &node.NodeInfo{
		ID:      "test-node-1",
		Name:    "Test Node 1",
		Address: "localhost:8080",
		Role:    node.NodeRoleClient,
	}

	err := reg.Register(info)
	if err != nil {
		t.Fatalf("注册客户端失败: %v", err)
	}

	// Test duplicate registration
	err = reg.Register(info)
	if err == nil {
		t.Error("重复注册应该返回错误")
	}

	// Test nil client
	err = reg.Register(nil)
	if err == nil {
		t.Error("注册 nil 客户端应该返回错误")
	}

	// Test empty ID
	info.ID = ""
	err = reg.Register(info)
	if err == nil {
		t.Error("注册空 ID 客户端应该返回错误")
	}
}

// TestMemoryClientRegistry_Unregister tests unregistering clients
func TestMemoryClientRegistry_Unregister(t *testing.T) {
	reg := NewMemoryClientRegistry(newTestLogger(t), 5*time.Minute, 1*time.Minute)

	// Register a client first
	info := &node.NodeInfo{
		ID:      "test-node-1",
		Name:    "Test Node 1",
		Address: "localhost:8080",
		Role:    node.NodeRoleClient,
	}
	reg.Register(info)

	// Test unregistering
	err := reg.Unregister("test-node-1")
	if err != nil {
		t.Fatalf("注销客户端失败: %v", err)
	}

	// Verify it's gone
	_, err = reg.Get("test-node-1")
	if err == nil {
		t.Error("客户端应该已被注销")
	}

	// Test unregistering non-existent client
	err = reg.Unregister("non-existent")
	if err == nil {
		t.Error("注销不存在的客户端应该返回错误")
	}
}

// TestMemoryClientRegistry_Get tests retrieving clients
func TestMemoryClientRegistry_Get(t *testing.T) {
	reg := NewMemoryClientRegistry(newTestLogger(t), 5*time.Minute, 1*time.Minute)

	// Test getting non-existent client
	_, err := reg.Get("non-existent")
	if err == nil {
		t.Error("获取不存在的客户端应该返回错误")
	}

	// Register and get
	info := &node.NodeInfo{
		ID:      "test-node-1",
		Name:    "Test Node 1",
		Address: "localhost:8080",
		Role:    node.NodeRoleClient,
	}
	reg.Register(info)

	retrieved, err := reg.Get("test-node-1")
	if err != nil {
		t.Fatalf("获取客户端失败: %v", err)
	}

	if retrieved.ID != "test-node-1" {
		t.Errorf("期望 ID=test-node-1, 实际=%s", retrieved.ID)
	}
}

// TestMemoryClientRegistry_List tests listing all clients
func TestMemoryClientRegistry_List(t *testing.T) {
	reg := NewMemoryClientRegistry(newTestLogger(t), 5*time.Minute, 1*time.Minute)

	// Empty list
	clients := reg.List()
	if len(clients) != 0 {
		t.Errorf("空注册表应该返回空列表，实际长度=%d", len(clients))
	}

	// Add some clients
	for i := 1; i <= 3; i++ {
		info := &node.NodeInfo{
			ID:      fmt.Sprintf("test-node-%d", i),
			Name:    fmt.Sprintf("Test Node %d", i),
			Address: "localhost:8080",
			Role:    node.NodeRoleClient,
		}
		reg.Register(info)
	}

	clients = reg.List()
	if len(clients) != 3 {
		t.Errorf("期望 3 个客户端，实际=%d", len(clients))
	}
}

// TestMemoryClientRegistry_GetStats tests getting statistics
func TestMemoryClientRegistry_GetStats(t *testing.T) {
	reg := NewMemoryClientRegistry(newTestLogger(t), 5*time.Minute, 1*time.Minute)

	// Register clients (Register will set status to Online)
	clients := []*node.NodeInfo{
		{ID: "node-1", Role: node.NodeRoleClient},
		{ID: "node-2", Role: node.NodeRoleClient},
		{ID: "node-3", Role: node.NodeRoleClient},
		{ID: "node-4", Role: node.NodeRoleClient},
	}

	for _, info := range clients {
		reg.Register(info)
	}

	// Update some clients to different statuses
	reg.UpdateStatus("node-3", node.NodeStatusOffline)
	reg.UpdateStatus("node-4", node.NodeStatusBusy)

	stats := reg.GetStats()
	if stats.TotalClients != 4 {
		t.Errorf("期望总数=4, 实际=%d", stats.TotalClients)
	}
	if stats.OnlineClients != 2 {
		t.Errorf("期望在线=2, 实际=%d", stats.OnlineClients)
	}
	if stats.OfflineClients != 1 {
		t.Errorf("期望离线=1, 实际=%d", stats.OfflineClients)
	}
	if stats.BusyClients != 1 {
		t.Errorf("期望忙碌=1, 实际=%d", stats.BusyClients)
	}
}

// TestMemoryClientRegistry_Find tests finding clients by predicate
func TestMemoryClientRegistry_Find(t *testing.T) {
	reg := NewMemoryClientRegistry(newTestLogger(t), 5*time.Minute, 1*time.Minute)

	// Register clients (Register sets status to Online)
	clients := []*node.NodeInfo{
		{ID: "node-1", Role: node.NodeRoleClient},
		{ID: "node-2", Role: node.NodeRoleMaster},
		{ID: "node-3", Role: node.NodeRoleClient},
	}

	for _, info := range clients {
		reg.Register(info)
	}

	// Update one to offline
	reg.UpdateStatus("node-3", node.NodeStatusOffline)

	// Find all online clients
	onlineClients := reg.Find(func(info *node.NodeInfo) bool {
		return info.Status == node.NodeStatusOnline
	})

	if len(onlineClients) != 2 {
		t.Errorf("期望找到 2 个在线客户端，实际=%d", len(onlineClients))
	}

	// Find all client role nodes
	clientNodes := reg.Find(func(info *node.NodeInfo) bool {
		return info.Role == node.NodeRoleClient
	})

	if len(clientNodes) != 2 {
		t.Errorf("期望找到 2 个客户端节点，实际=%d", len(clientNodes))
	}
}

// TestMemoryClientRegistry_UpdateStatus tests updating client status
func TestMemoryClientRegistry_UpdateStatus(t *testing.T) {
	reg := NewMemoryClientRegistry(newTestLogger(t), 5*time.Minute, 1*time.Minute)

	info := &node.NodeInfo{
		ID:   "test-node-1",
		Role: node.NodeRoleClient,
		// Status will be set to Online by Register
	}
	reg.Register(info)

	// Update status
	err := reg.UpdateStatus("test-node-1", node.NodeStatusBusy)
	if err != nil {
		t.Fatalf("更新状态失败: %v", err)
	}

	// Verify
	client, _ := reg.Get("test-node-1")
	if client.Status != node.NodeStatusBusy {
		t.Errorf("期望状态=%s, 实际=%s", node.NodeStatusBusy, client.Status)
	}

	// Test non-existent client
	err = reg.UpdateStatus("non-existent", node.NodeStatusBusy)
	if err == nil {
		t.Error("更新不存在的客户端状态应该返回错误")
	}
}

// TestMemoryClientRegistry_UpdateResources tests updating client resources
func TestMemoryClientRegistry_UpdateResources(t *testing.T) {
	reg := NewMemoryClientRegistry(newTestLogger(t), 5*time.Minute, 1*time.Minute)

	info := &node.NodeInfo{
		ID:     "test-node-1",
		Role:   node.NodeRoleClient,
		Status: node.NodeStatusOnline,
	}
	reg.Register(info)

	resources := &node.NodeResources{
		CPUTotal:    4000,
		MemoryTotal: 8 * 1024 * 1024 * 1024,
	}

	err := reg.UpdateResources("test-node-1", resources)
	if err != nil {
		t.Fatalf("更新资源失败: %v", err)
	}

	// Verify
	client, _ := reg.Get("test-node-1")
	if client.Resources == nil {
		t.Error("资源应该已被更新")
	}
}

// TestMemoryClientRegistry_Cleanup tests cleaning up stale clients
func TestMemoryClientRegistry_Cleanup(t *testing.T) {
	reg := NewMemoryClientRegistry(newTestLogger(t), 5*time.Minute, 1*time.Minute)

	// Register an active client
	info1 := &node.NodeInfo{
		ID:   "active-node",
		Role: node.NodeRoleClient,
	}
	reg.Register(info1)

	// Register and simulate a stale client by directly modifying LastSeen
	info2 := &node.NodeInfo{
		ID:   "stale-node",
		Role: node.NodeRoleClient,
	}
	reg.Register(info2)

	// Manually set stale node's LastSeen to an old time
	reg.mu.Lock()
	if staleClient, exists := reg.clients["stale-node"]; exists {
		staleClient.LastSeen = time.Now().Add(-2 * time.Minute)
	}
	reg.mu.Unlock()

	// Cleanup with 1 minute timeout
	cleaned := reg.Cleanup(1 * time.Minute)

	if cleaned != 1 {
		t.Errorf("期望清理 1 个客户端，实际=%d", cleaned)
	}

	// Verify active node still exists
	if !reg.Exists("active-node") {
		t.Error("活跃客户端应该仍然存在")
	}

	// Verify stale node was removed
	if reg.Exists("stale-node") {
		t.Error("过期客户端应该已被移除")
	}
}

// TestMemoryClientRegistry_GetOnlineClients tests getting online clients
func TestMemoryClientRegistry_GetOnlineClients(t *testing.T) {
	reg := NewMemoryClientRegistry(newTestLogger(t), 5*time.Minute, 1*time.Minute)

	// Register clients (Register sets status to Online)
	clients := []*node.NodeInfo{
		{ID: "node-1", Role: node.NodeRoleClient},
		{ID: "node-2", Role: node.NodeRoleClient},
		{ID: "node-3", Role: node.NodeRoleClient},
	}

	for _, info := range clients {
		reg.Register(info)
	}

	// Set one to offline
	reg.UpdateStatus("node-2", node.NodeStatusOffline)

	online := reg.GetOnlineClients()
	if len(online) != 2 {
		t.Errorf("期望 2 个在线客户端，实际=%d", len(online))
	}
}
