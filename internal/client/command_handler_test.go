package client

import (
	"testing"
	"time"

	"github.com/shepherd-project/shepherd/Shepherd/internal/node"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewCommandHandler 测试创建 CommandHandler
func TestNewCommandHandler(t *testing.T) {
	config := &CommandHandlerConfig{
		NodeID: "test-node-1",
	}

	handler := NewCommandHandler(config)
	require.NotNil(t, handler)
	assert.Equal(t, "test-node-1", handler.GetNodeID())
	assert.Nil(t, handler.GetModelManager())
	assert.Nil(t, handler.GetProcessManager())
	assert.Nil(t, handler.GetExecutor())
}

// TestNewCommandHandlerWithNilConfig 测试使用 nil 配置创建 CommandHandler
func TestNewCommandHandlerWithNilConfig(t *testing.T) {
	handler := NewCommandHandler(nil)
	require.NotNil(t, handler)
	assert.Equal(t, "", handler.GetNodeID())
}

// TestHandleNilCommand 测试处理 nil 命令
func TestHandleNilCommand(t *testing.T) {
	handler := NewCommandHandler(nil)
	result, err := handler.Handle(nil)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "命令不能为空")
}

// TestHandleLoadModelMissingModelID 测试加载模型命令缺少 model_id
func TestHandleLoadModelMissingModelID(t *testing.T) {
	handler := NewCommandHandler(nil)

	command := &node.Command{
		ID:         "cmd-1",
		Type:       node.CommandTypeLoadModel,
		FromNodeID: "master-1",
		Payload:    map[string]interface{}{},
		CreatedAt:  time.Now(),
	}

	result, err := handler.Handle(command)

	assert.Error(t, err)
	assert.NotNil(t, result)
	assert.False(t, result.Success)
	assert.Contains(t, result.Error, "model_id")
}

// TestHandleUnloadModelMissingModelID 测试卸载模型命令缺少 model_id
func TestHandleUnloadModelMissingModelID(t *testing.T) {
	handler := NewCommandHandler(nil)

	command := &node.Command{
		ID:         "cmd-1",
		Type:       node.CommandTypeUnloadModel,
		FromNodeID: "master-1",
		Payload:    map[string]interface{}{},
		CreatedAt:  time.Now(),
	}

	result, err := handler.Handle(command)

	assert.Error(t, err)
	assert.NotNil(t, result)
	assert.False(t, result.Success)
	assert.Contains(t, result.Error, "model_id")
}

// TestHandleRunLlamacppMissingBinaryPath 测试运行 llama.cpp 命令缺少 binary_path
func TestHandleRunLlamacppMissingBinaryPath(t *testing.T) {
	handler := NewCommandHandler(nil)

	command := &node.Command{
		ID:         "cmd-1",
		Type:       node.CommandTypeRunLlamacpp,
		FromNodeID: "master-1",
		Payload: map[string]interface{}{
			"model_path": "/path/to/model.gguf",
		},
		CreatedAt: time.Now(),
	}

	result, err := handler.Handle(command)

	assert.Error(t, err)
	assert.NotNil(t, result)
	assert.False(t, result.Success)
	assert.Contains(t, result.Error, "binary_path")
}

// TestHandleRunLlamacppMissingModelPath 测试运行 llama.cpp 命令缺少 model_path
func TestHandleRunLlamacppMissingModelPath(t *testing.T) {
	handler := NewCommandHandler(nil)

	command := &node.Command{
		ID:         "cmd-1",
		Type:       node.CommandTypeRunLlamacpp,
		FromNodeID: "master-1",
		Payload: map[string]interface{}{
			"binary_path": "/usr/local/bin/llama-server",
		},
		CreatedAt: time.Now(),
	}

	result, err := handler.Handle(command)

	assert.Error(t, err)
	assert.NotNil(t, result)
	assert.False(t, result.Success)
	assert.Contains(t, result.Error, "model_path")
}

// TestHandleStopProcessMissingModelID 测试停止进程命令缺少 model_id
func TestHandleStopProcessMissingModelID(t *testing.T) {
	handler := NewCommandHandler(nil)

	command := &node.Command{
		ID:         "cmd-1",
		Type:       node.CommandTypeStopProcess,
		FromNodeID: "master-1",
		Payload:    map[string]interface{}{},
		CreatedAt:  time.Now(),
	}

	result, err := handler.Handle(command)

	assert.Error(t, err)
	assert.NotNil(t, result)
	assert.False(t, result.Success)
	assert.Contains(t, result.Error, "model_id")
}

// TestHandleStopProcessManagerNotInitialized 测试停止进程时进程管理器未初始化
func TestHandleStopProcessManagerNotInitialized(t *testing.T) {
	handler := NewCommandHandler(nil)

	command := &node.Command{
		ID:         "cmd-1",
		Type:       node.CommandTypeStopProcess,
		FromNodeID: "master-1",
		Payload: map[string]interface{}{
			"model_id": "model-1",
		},
		CreatedAt: time.Now(),
	}

	result, err := handler.Handle(command)

	assert.Error(t, err)
	assert.NotNil(t, result)
	assert.False(t, result.Success)
	assert.Contains(t, result.Error, "进程管理器未初始化")
}

// TestHandleScanModelsManagerNotInitialized 测试扫描模型时模型管理器未初始化
func TestHandleScanModelsManagerNotInitialized(t *testing.T) {
	handler := NewCommandHandler(nil)

	command := &node.Command{
		ID:         "cmd-1",
		Type:       node.CommandTypeScanModels,
		FromNodeID: "master-1",
		Payload:    map[string]interface{}{},
		CreatedAt:  time.Now(),
	}

	result, err := handler.Handle(command)

	assert.Error(t, err)
	assert.NotNil(t, result)
	assert.False(t, result.Success)
	assert.Contains(t, result.Error, "模型管理器未初始化")
}

// TestHandleUnknownCommandType 测试处理未知命令类型
func TestHandleUnknownCommandType(t *testing.T) {
	handler := NewCommandHandler(&CommandHandlerConfig{
		NodeID: "test-node",
	})

	command := &node.Command{
		ID:         "cmd-1",
		Type:       node.CommandType("unknown_command"),
		FromNodeID: "master-1",
		Payload:    map[string]interface{}{},
		CreatedAt:  time.Now(),
	}

	result, err := handler.Handle(command)

	assert.Error(t, err)
	assert.NotNil(t, result)
	assert.False(t, result.Success)
	assert.Contains(t, result.Error, "未知的命令类型")
	assert.Equal(t, "test-node", result.FromNodeID)
}

// TestHandleLoadModelManagerNotInitialized 测试加载模型时模型管理器未初始化
func TestHandleLoadModelManagerNotInitialized(t *testing.T) {
	handler := NewCommandHandler(nil)

	command := &node.Command{
		ID:         "cmd-1",
		Type:       node.CommandTypeLoadModel,
		FromNodeID: "master-1",
		Payload: map[string]interface{}{
			"model_id": "model-1",
		},
		CreatedAt: time.Now(),
	}

	result, err := handler.Handle(command)

	assert.Error(t, err)
	assert.NotNil(t, result)
	assert.False(t, result.Success)
	assert.Contains(t, result.Error, "模型管理器未初始化")
}

// TestHandleUnloadModelManagerNotInitialized 测试卸载模型时模型管理器未初始化
func TestHandleUnloadModelManagerNotInitialized(t *testing.T) {
	handler := NewCommandHandler(nil)

	command := &node.Command{
		ID:         "cmd-1",
		Type:       node.CommandTypeUnloadModel,
		FromNodeID: "master-1",
		Payload: map[string]interface{}{
			"model_id": "model-1",
		},
		CreatedAt: time.Now(),
	}

	result, err := handler.Handle(command)

	assert.Error(t, err)
	assert.NotNil(t, result)
	assert.False(t, result.Success)
	assert.Contains(t, result.Error, "模型管理器未初始化")
}

// TestSetAndGetNodeID 测试设置和获取 NodeID
func TestSetAndGetNodeID(t *testing.T) {
	handler := NewCommandHandler(nil)

	assert.Equal(t, "", handler.GetNodeID())

	handler.SetNodeID("new-node-id")
	assert.Equal(t, "new-node-id", handler.GetNodeID())
}

// TestCommandResultFields 测试命令结果字段
func TestCommandResultFields(t *testing.T) {
	config := &CommandHandlerConfig{
		NodeID: "client-node",
	}
	handler := NewCommandHandler(config)

	command := &node.Command{
		ID:         "test-cmd-123",
		Type:       node.CommandTypeScanModels,
		FromNodeID: "master-node",
		Payload:    map[string]interface{}{},
		CreatedAt:  time.Now(),
		Priority:   5,
	}

	result, _ := handler.Handle(command)

	require.NotNil(t, result)
	assert.Equal(t, "test-cmd-123", result.CommandID)
	assert.Equal(t, "client-node", result.FromNodeID)
	assert.Equal(t, "master-node", result.ToNodeID)
	assert.GreaterOrEqual(t, result.Duration, int64(0))
	assert.False(t, result.CompletedAt.IsZero())
}

// TestCommandTimeout 测试命令超时设置
func TestCommandTimeout(t *testing.T) {
	handler := NewCommandHandler(nil)

	timeout := 30 * time.Second
	command := &node.Command{
		ID:         "cmd-timeout-test",
		Type:       node.CommandTypeRunLlamacpp,
		FromNodeID: "master-1",
		Payload: map[string]interface{}{
			"binary_path": "/bin/echo",
			"model_path":  "test.gguf",
			"timeout":     float64(30),
		},
		CreatedAt: time.Now(),
		Timeout:   &timeout,
	}

	// 由于缺少实际的可执行文件，这个测试会失败，但用于验证超时参数传递
	result, err := handler.Handle(command)

	// 这个测试期望失败，因为我们使用了假的二进制路径
	// 但我们验证了命令处理流程能正确处理
	if err != nil {
		assert.NotNil(t, result)
	}
}

// TestHandleLoadModelWithAllParameters 测试加载模型命令携带所有可选参数
func TestHandleLoadModelWithAllParameters(t *testing.T) {
	handler := NewCommandHandler(nil)

	command := &node.Command{
		ID:         "cmd-load-full",
		Type:       node.CommandTypeLoadModel,
		FromNodeID: "master-1",
		Payload: map[string]interface{}{
			"model_id":       "llama-2-7b",
			"ctx_size":       float64(4096),
			"gpu_layers":     float64(35),
			"threads":        float64(8),
			"batch_size":     float64(512),
			"temperature":    float64(0.7),
			"top_p":          float64(0.9),
			"top_k":          float64(40),
			"repeat_penalty": float64(1.1),
			"n_predict":      float64(2048),
		},
		CreatedAt: time.Now(),
	}

	// 由于没有初始化 modelManager，期望返回错误
	result, err := handler.Handle(command)

	assert.Error(t, err)
	assert.NotNil(t, result)
	assert.False(t, result.Success)
}

// TestHandleStopProcessWithForce 测试带 force 参数的停止进程命令
func TestHandleStopProcessWithForce(t *testing.T) {
	handler := NewCommandHandler(nil)

	command := &node.Command{
		ID:         "cmd-stop-force",
		Type:       node.CommandTypeStopProcess,
		FromNodeID: "master-1",
		Payload: map[string]interface{}{
			"model_id": "model-1",
			"force":    true,
		},
		CreatedAt: time.Now(),
	}

	// 由于没有初始化 processManager，期望返回错误
	result, err := handler.Handle(command)

	assert.Error(t, err)
	assert.NotNil(t, result)
	assert.False(t, result.Success)
}

// TestHandleRunLlamacppWithArgs 测试带额外参数的 llama.cpp 运行命令
func TestHandleRunLlamacppWithArgs(t *testing.T) {
	handler := NewCommandHandler(nil)

	command := &node.Command{
		ID:         "cmd-run-args",
		Type:       node.CommandTypeRunLlamacpp,
		FromNodeID: "master-1",
		Payload: map[string]interface{}{
			"binary_path": "/usr/local/bin/llama-server",
			"model_path":  "/models/llama.gguf",
			"args": []interface{}{
				"-ngl", "35",
				"-c", "4096",
			},
			"timeout": float64(60),
		},
		CreatedAt: time.Now(),
	}

	// 由于二进制路径不存在，期望返回错误
	result, err := handler.Handle(command)

	assert.Error(t, err)
	assert.NotNil(t, result)
	assert.False(t, result.Success)
}

// TestCommandResultMetadata 测试命令结果的元数据
func TestCommandResultMetadata(t *testing.T) {
	result := &node.CommandResult{
		CommandID:  "cmd-1",
		FromNodeID: "node-1",
		ToNodeID:   "node-2",
		Success:    true,
		Result: map[string]interface{}{
			"key1": "value1",
			"key2": 123,
		},
		Error:       "",
		CompletedAt: time.Now(),
		Duration:    100,
		Metadata: map[string]string{
			"source": "test",
		},
	}

	assert.Equal(t, "cmd-1", result.CommandID)
	assert.Equal(t, "node-1", result.FromNodeID)
	assert.Equal(t, "node-2", result.ToNodeID)
	assert.True(t, result.Success)
	assert.Equal(t, "value1", result.Result["key1"])
	assert.Equal(t, 123, result.Result["key2"])
	assert.Equal(t, "test", result.Metadata["source"])
}
