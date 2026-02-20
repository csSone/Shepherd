package client

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/shepherd-project/shepherd/Shepherd/internal/config"
	"github.com/shepherd-project/shepherd/Shepherd/internal/logger"
	"github.com/shepherd-project/shepherd/Shepherd/internal/node"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockResourceMonitor 模拟资源监控器
type mockResourceMonitor struct{}

func (m *mockResourceMonitor) GetSnapshot() *node.NodeResources {
	return &node.NodeResources{
		CPUUsed:     100,
		CPUTotal:    4000,
		MemoryUsed:  1024 * 1024 * 1024,
		MemoryTotal: 8 * 1024 * 1024 * 1024,
	}
}

// newTestLogger 创建测试用日志记录器
func newTestLogger(t *testing.T) *logger.Logger {
	cfg := &config.LogConfig{
		Level:      "info",
		Format:     "text",
		Output:     "stdout",
		Directory:  "logs",
		MaxSize:    100,
		MaxBackups: 7,
		MaxAge:     7,
	}
	log, err := logger.NewLogger(cfg, "test")
	require.NoError(t, err)
	return log
}

// TestNewMasterConnector 测试创建 MasterConnector
func TestNewMasterConnector(t *testing.T) {
	log := newTestLogger(t)
	hbMgr := node.NewHeartbeatManager(&node.HeartbeatConfig{
		NodeID:     "test-node",
		MasterAddr: "http://localhost:9190",
		Logger:     log,
	})
	executor := node.NewCommandExecutor(&node.CommandExecutorConfig{
		Logger: log,
	})

	tests := []struct {
		name    string
		config  *MasterConnectorConfig
		wantErr bool
		errMsg  string
	}{
		{
			name: "正常创建",
			config: &MasterConnectorConfig{
				NodeID:       "test-node",
				MasterAddr:   "http://localhost:9190",
				HeartbeatMgr: hbMgr,
				Executor:     executor,
				Logger:       log,
			},
			wantErr: false,
		},
		{
			name:    "配置为空",
			config:  nil,
			wantErr: true,
			errMsg:  "配置不能为空",
		},
		{
			name: "节点ID为空",
			config: &MasterConnectorConfig{
				NodeID:       "",
				MasterAddr:   "http://localhost:9190",
				HeartbeatMgr: hbMgr,
				Executor:     executor,
			},
			wantErr: true,
			errMsg:  "节点 ID 不能为空",
		},
		{
			name: "Master地址为空",
			config: &MasterConnectorConfig{
				NodeID:       "test-node",
				MasterAddr:   "",
				HeartbeatMgr: hbMgr,
				Executor:     executor,
			},
			wantErr: true,
			errMsg:  "Master 地址不能为空",
		},
		{
			name: "HeartbeatManager为空",
			config: &MasterConnectorConfig{
				NodeID:     "test-node",
				MasterAddr: "http://localhost:9190",
				Executor:   executor,
			},
			wantErr: true,
			errMsg:  "HeartbeatManager 不能为空",
		},
		{
			name: "CommandExecutor为空",
			config: &MasterConnectorConfig{
				NodeID:       "test-node",
				MasterAddr:   "http://localhost:9190",
				HeartbeatMgr: hbMgr,
			},
			wantErr: true,
			errMsg:  "CommandExecutor 不能为空",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conn, err := NewMasterConnector(tt.config)
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				return
			}
			require.NoError(t, err)
			assert.NotNil(t, conn)
			assert.Equal(t, "test-node", conn.GetNodeID())
			assert.Equal(t, "http://localhost:9190", conn.GetMasterAddr())
		})
	}
}

// TestMasterConnector_Connect 测试连接功能
func TestMasterConnector_Connect(t *testing.T) {
	log := newTestLogger(t)

	t.Run("连接成功", func(t *testing.T) {
		// 创建模拟服务器
		registered := false
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Path {
			case "/api/master/nodes/register":
				registered = true
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
			case "/api/master/heartbeat":
				w.WriteHeader(http.StatusOK)
			case "/api/master/nodes/test-node/commands":
				w.WriteHeader(http.StatusNoContent)
			default:
				w.WriteHeader(http.StatusNotFound)
			}
		}))
		defer server.Close()

		hbMgr := node.NewHeartbeatManager(&node.HeartbeatConfig{
			NodeID:     "test-node",
			MasterAddr: server.URL,
			Interval:   1 * time.Second,
			Logger:     log,
		})
		executor := node.NewCommandExecutor(&node.CommandExecutorConfig{
			Logger: log,
		})

		conn, err := NewMasterConnector(&MasterConnectorConfig{
			NodeID:              "test-node",
			MasterAddr:          server.URL,
			HeartbeatMgr:        hbMgr,
			Executor:            executor,
			CommandPollInterval: 500 * time.Millisecond,
			Logger:              log,
		})
		require.NoError(t, err)

		err = conn.Connect()
		require.NoError(t, err)
		assert.True(t, registered)
		assert.True(t, conn.IsRegistered())

		// 等待心跳发送
		time.Sleep(1500 * time.Millisecond)
		conn.Disconnect()
	})

	t.Run("注册失败重试", func(t *testing.T) {
		attemptCount := 0
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/api/master/nodes/register" {
				attemptCount++
				if attemptCount < 3 {
					w.WriteHeader(http.StatusServiceUnavailable)
					return
				}
				w.WriteHeader(http.StatusOK)
			}
		}))
		defer server.Close()

		hbMgr := node.NewHeartbeatManager(&node.HeartbeatConfig{
			NodeID:     "test-node",
			MasterAddr: server.URL,
			Logger:     log,
		})
		executor := node.NewCommandExecutor(&node.CommandExecutorConfig{
			Logger: log,
		})

		conn, err := NewMasterConnector(&MasterConnectorConfig{
			NodeID:               "test-node",
			MasterAddr:           server.URL,
			HeartbeatMgr:         hbMgr,
			Executor:             executor,
			MaxReconnectAttempts: 5,
			ReconnectBackoff:     100 * time.Millisecond,
			Logger:               log,
		})
		require.NoError(t, err)

		err = conn.Connect()
		require.NoError(t, err)
		assert.True(t, conn.IsRegistered())
		assert.GreaterOrEqual(t, attemptCount, 3)

		conn.Disconnect()
	})

	t.Run("注册最终失败", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusServiceUnavailable)
		}))
		defer server.Close()

		hbMgr := node.NewHeartbeatManager(&node.HeartbeatConfig{
			NodeID:     "test-node",
			MasterAddr: server.URL,
			Logger:     log,
		})
		executor := node.NewCommandExecutor(&node.CommandExecutorConfig{
			Logger: log,
		})

		conn, err := NewMasterConnector(&MasterConnectorConfig{
			NodeID:               "test-node",
			MasterAddr:           server.URL,
			HeartbeatMgr:         hbMgr,
			Executor:             executor,
			MaxReconnectAttempts: 2,
			ReconnectBackoff:     50 * time.Millisecond,
			Logger:               log,
		})
		require.NoError(t, err)

		err = conn.Connect()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "注册到 Master 失败")
		assert.False(t, conn.IsRegistered())
	})

	t.Run("重复连接", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		hbMgr := node.NewHeartbeatManager(&node.HeartbeatConfig{
			NodeID:     "test-node",
			MasterAddr: server.URL,
			Logger:     log,
		})
		executor := node.NewCommandExecutor(&node.CommandExecutorConfig{
			Logger: log,
		})

		conn, err := NewMasterConnector(&MasterConnectorConfig{
			NodeID:       "test-node",
			MasterAddr:   server.URL,
			HeartbeatMgr: hbMgr,
			Executor:     executor,
			Logger:       log,
		})
		require.NoError(t, err)

		err = conn.Connect()
		require.NoError(t, err)
		defer conn.Disconnect()

		// 尝试再次连接
		err = conn.Connect()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "节点已注册")
	})
}

// TestMasterConnector_Disconnect 测试断开连接
func TestMasterConnector_Disconnect(t *testing.T) {
	log := newTestLogger(t)

	t.Run("正常断开", func(t *testing.T) {
		unregistered := false
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Path {
			case "/api/master/nodes/register":
				w.WriteHeader(http.StatusOK)
			case "/api/master/nodes/test-node/unregister":
				unregistered = true
				w.WriteHeader(http.StatusOK)
			case "/api/master/heartbeat":
				w.WriteHeader(http.StatusOK)
			case "/api/master/nodes/test-node/commands":
				w.WriteHeader(http.StatusNoContent)
			}
		}))
		defer server.Close()

		hbMgr := node.NewHeartbeatManager(&node.HeartbeatConfig{
			NodeID:     "test-node",
			MasterAddr: server.URL,
			Logger:     log,
		})
		executor := node.NewCommandExecutor(&node.CommandExecutorConfig{
			Logger: log,
		})

		conn, err := NewMasterConnector(&MasterConnectorConfig{
			NodeID:              "test-node",
			MasterAddr:          server.URL,
			HeartbeatMgr:        hbMgr,
			Executor:            executor,
			CommandPollInterval: 500 * time.Millisecond,
			Logger:              log,
		})
		require.NoError(t, err)

		err = conn.Connect()
		require.NoError(t, err)
		assert.True(t, conn.IsRegistered())

		err = conn.Disconnect()
		require.NoError(t, err)
		assert.False(t, conn.IsRegistered())
		assert.True(t, unregistered)
	})

	t.Run("重复断开", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		hbMgr := node.NewHeartbeatManager(&node.HeartbeatConfig{
			NodeID:     "test-node",
			MasterAddr: server.URL,
			Logger:     log,
		})
		executor := node.NewCommandExecutor(&node.CommandExecutorConfig{
			Logger: log,
		})

		conn, err := NewMasterConnector(&MasterConnectorConfig{
			NodeID:       "test-node",
			MasterAddr:   server.URL,
			HeartbeatMgr: hbMgr,
			Executor:     executor,
			Logger:       log,
		})
		require.NoError(t, err)

		// 第一次断开（未连接状态）
		err = conn.Disconnect()
		require.NoError(t, err)

		// 第二次断开
		err = conn.Disconnect()
		require.NoError(t, err)
	})
}

// TestMasterConnector_CommandExecution 测试命令执行
func TestMasterConnector_CommandExecution(t *testing.T) {
	log := newTestLogger(t)

	t.Run("执行命令并上报结果", func(t *testing.T) {
		commandReceived := false
		resultReceived := false

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Path {
			case "/api/master/nodes/register":
				w.WriteHeader(http.StatusOK)
			case "/api/master/heartbeat":
				w.WriteHeader(http.StatusOK)
			case "/api/master/nodes/test-node/commands":
				if !commandReceived {
					commandReceived = true
					cmds := []*node.Command{
						{
							ID:         "cmd-1",
							Type:       node.CommandTypeLoadModel,
							FromNodeID: "master",
							Payload:    map[string]interface{}{"model": "test.gguf"},
							CreatedAt:  time.Now(),
						},
					}
					w.WriteHeader(http.StatusOK)
					json.NewEncoder(w).Encode(cmds)
				} else {
					w.WriteHeader(http.StatusNoContent)
				}
			case "/api/master/command/result":
				resultReceived = true
				body, _ := io.ReadAll(r.Body)
				var result node.CommandResult
				json.Unmarshal(body, &result)
				assert.Equal(t, "cmd-1", result.CommandID)
				assert.True(t, result.Success)
				w.WriteHeader(http.StatusOK)
			case "/api/master/nodes/test-node/unregister":
				w.WriteHeader(http.StatusOK)
			}
		}))
		defer server.Close()

		hbMgr := node.NewHeartbeatManager(&node.HeartbeatConfig{
			NodeID:     "test-node",
			MasterAddr: server.URL,
			Logger:     log,
		})
		executor := node.NewCommandExecutor(&node.CommandExecutorConfig{
			Logger: log,
		})

		// 使用自定义命令处理器来确保命令成功
		customHandler := func(cmd *node.Command) (*node.CommandResult, error) {
			return &node.CommandResult{
				CommandID:   cmd.ID,
				FromNodeID:  "test-node",
				ToNodeID:    cmd.FromNodeID,
				Success:     true,
				Result:      map[string]interface{}{"model": "loaded"},
				CompletedAt: time.Now(),
			}, nil
		}

		conn, err := NewMasterConnector(&MasterConnectorConfig{
			NodeID:              "test-node",
			MasterAddr:          server.URL,
			HeartbeatMgr:        hbMgr,
			Executor:            executor,
			CommandPollInterval: 200 * time.Millisecond,
			CommandHandler:      customHandler,
			Logger:              log,
		})
		require.NoError(t, err)

		err = conn.Connect()
		require.NoError(t, err)

		// 等待命令执行和结果上报
		time.Sleep(500 * time.Millisecond)

		conn.Disconnect()

		assert.True(t, commandReceived, "应该收到命令请求")
		assert.True(t, resultReceived, "应该收到结果上报")
	})
}

// TestMasterConnector_SetCommandHandler 测试设置命令处理器
func TestMasterConnector_SetCommandHandler(t *testing.T) {
	log := newTestLogger(t)
	hbMgr := node.NewHeartbeatManager(&node.HeartbeatConfig{
		NodeID:     "test-node",
		MasterAddr: "http://localhost:9190",
		Logger:     log,
	})
	executor := node.NewCommandExecutor(&node.CommandExecutorConfig{
		Logger: log,
	})

	conn, err := NewMasterConnector(&MasterConnectorConfig{
		NodeID:       "test-node",
		MasterAddr:   "http://localhost:9190",
		HeartbeatMgr: hbMgr,
		Executor:     executor,
		Logger:       log,
	})
	require.NoError(t, err)

	// 设置自定义处理器
	customHandler := func(cmd *node.Command) (*node.CommandResult, error) {
		return &node.CommandResult{
			CommandID:   cmd.ID,
			Success:     true,
			Result:      map[string]interface{}{"custom": true},
			CompletedAt: time.Now(),
		}, nil
	}

	conn.SetCommandHandler(customHandler)
	assert.NotNil(t, conn.commandHandler)
}

// TestMasterConnector_UpdateNodeInfo 测试更新节点信息
func TestMasterConnector_UpdateNodeInfo(t *testing.T) {
	log := newTestLogger(t)
	hbMgr := node.NewHeartbeatManager(&node.HeartbeatConfig{
		NodeID:     "test-node",
		MasterAddr: "http://localhost:9190",
		Logger:     log,
	})
	executor := node.NewCommandExecutor(&node.CommandExecutorConfig{
		Logger: log,
	})

	initialInfo := &node.NodeInfo{
		ID:   "test-node",
		Name: "Initial Name",
	}

	conn, err := NewMasterConnector(&MasterConnectorConfig{
		NodeID:       "test-node",
		MasterAddr:   "http://localhost:9190",
		NodeInfo:     initialInfo,
		HeartbeatMgr: hbMgr,
		Executor:     executor,
		Logger:       log,
	})
	require.NoError(t, err)

	assert.Equal(t, "Initial Name", conn.nodeInfo.Name)

	// 更新节点信息
	newInfo := &node.NodeInfo{
		ID:   "test-node",
		Name: "Updated Name",
	}
	conn.UpdateNodeInfo(newInfo)

	assert.Equal(t, "Updated Name", conn.nodeInfo.Name)
}

// TestMasterConnector_Getters 测试 getter 方法
func TestMasterConnector_Getters(t *testing.T) {
	log := newTestLogger(t)
	hbMgr := node.NewHeartbeatManager(&node.HeartbeatConfig{
		NodeID:     "test-node",
		MasterAddr: "http://localhost:9190",
		Logger:     log,
	})
	executor := node.NewCommandExecutor(&node.CommandExecutorConfig{
		Logger: log,
	})

	conn, err := NewMasterConnector(&MasterConnectorConfig{
		NodeID:       "test-node-123",
		MasterAddr:   "http://master:9190",
		HeartbeatMgr: hbMgr,
		Executor:     executor,
		Logger:       log,
	})
	require.NoError(t, err)

	assert.Equal(t, "test-node-123", conn.GetNodeID())
	assert.Equal(t, "http://master:9190", conn.GetMasterAddr())
	assert.False(t, conn.IsConnected())
	assert.False(t, conn.IsRegistered())
}

// TestCalculateBackoff 测试退避计算
func TestCalculateBackoff(t *testing.T) {
	log := newTestLogger(t)
	hbMgr := node.NewHeartbeatManager(&node.HeartbeatConfig{
		NodeID:     "test-node",
		MasterAddr: "http://localhost:9190",
		Logger:     log,
	})
	executor := node.NewCommandExecutor(&node.CommandExecutorConfig{
		Logger: log,
	})

	conn, err := NewMasterConnector(&MasterConnectorConfig{
		NodeID:           "test-node",
		MasterAddr:       "http://localhost:9190",
		HeartbeatMgr:     hbMgr,
		Executor:         executor,
		ReconnectBackoff: 1 * time.Second,
		Logger:           log,
	})
	require.NoError(t, err)

	// 测试不同尝试次数的退避时间
	tests := []struct {
		attempt  int
		maxDelay time.Duration
		minDelay time.Duration
	}{
		{1, 2 * time.Second, 1 * time.Second},
		{2, 3 * time.Second, 2 * time.Second},
		{3, 6 * time.Second, 4 * time.Second},
		{4, 12 * time.Second, 8 * time.Second},
		{10, 75 * time.Second, 60 * time.Second}, // 最大 60s
	}

	for _, tt := range tests {
		delay := conn.calculateBackoff(tt.attempt)
		assert.LessOrEqual(t, delay, tt.maxDelay, "尝试 %d 的延迟应该小于等于 %v", tt.attempt, tt.maxDelay)
		assert.GreaterOrEqual(t, delay, tt.minDelay, "尝试 %d 的延迟应该大于等于 %v", tt.attempt, tt.minDelay)
	}
}

// BenchmarkMasterConnector_Connect 基准测试连接性能
func BenchmarkMasterConnector_Connect(b *testing.B) {
	cfg := &config.LogConfig{
		Level:      "info",
		Format:     "text",
		Output:     "stdout",
		Directory:  "logs",
		MaxSize:    100,
		MaxBackups: 7,
		MaxAge:     7,
	}
	log, _ := logger.NewLogger(cfg, "test")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		hbMgr := node.NewHeartbeatManager(&node.HeartbeatConfig{
			NodeID:     fmt.Sprintf("node-%d", i),
			MasterAddr: server.URL,
			Logger:     log,
		})
		executor := node.NewCommandExecutor(&node.CommandExecutorConfig{
			Logger: log,
		})

		conn, _ := NewMasterConnector(&MasterConnectorConfig{
			NodeID:       fmt.Sprintf("node-%d", i),
			MasterAddr:   server.URL,
			HeartbeatMgr: hbMgr,
			Executor:     executor,
			Logger:       log,
		})

		conn.Connect()
		conn.Disconnect()
	}
}
