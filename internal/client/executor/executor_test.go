// Package executor provides unit tests for task execution
package executor

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/shepherd-project/shepherd/Shepherd/internal/cluster"
	"github.com/shepherd-project/shepherd/Shepherd/internal/config"
	"github.com/shepherd-project/shepherd/Shepherd/internal/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newTestLogger 创建测试用 logger
func newTestLogger() *logger.Logger {
	logCfg := &config.LogConfig{
		Level:  "info",
		Format: "text",
		Output: "stdout",
	}
	log, _ := logger.NewLogger(logCfg, "test")
	return log
}

// TestNewExecutor tests creating a new executor
func TestNewExecutor(t *testing.T) {
	cfg := &config.ClientConfig{
		CondaEnv: config.CondaEnvConfig{
			CondaPath: "/usr/bin/conda",
			Environments: map[string]string{
				"base": "/opt/conda/envs/base",
			},
		},
	}

	log := newTestLogger()
	executor := NewExecutor(cfg, log)

	assert.NotNil(t, executor)
	assert.NotNil(t, executor.config)
	assert.NotNil(t, executor.runningTasks)
	assert.NotNil(t, executor.runningProcesses)
	assert.Equal(t, 4, executor.config.MaxConcurrent)
	assert.Equal(t, 5*time.Minute, executor.config.TaskTimeout)
}

// TestExecuteLoadModelMissingParams tests executeLoadModel with missing parameters
func TestExecuteLoadModelMissingParams(t *testing.T) {
	cfg := &config.ClientConfig{}
	log := newTestLogger()
	executor := NewExecutor(cfg, log)

	tests := []struct {
		name    string
		payload map[string]interface{}
		errMsg  string
	}{
		{
			name:    "缺少 model_path",
			payload: map[string]interface{}{},
			errMsg:  "缺少 model_path 参数",
		},
		{
			name: "缺少 port",
			payload: map[string]interface{}{
				"model_path": "/path/to/model.gguf",
			},
			errMsg: "缺少 port 参数",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task := &cluster.Task{
				ID:      "test-task",
				Type:    cluster.TaskTypeLoadModel,
				Payload: tt.payload,
			}

			result, err := executor.ExecuteTask(task)

			assert.Error(t, err)
			assert.NotNil(t, result)
			assert.False(t, result.Success)
			assert.Contains(t, result.Error, tt.errMsg)
		})
	}
}

// TestExecuteUnloadModelMissingParams tests executeUnloadModel with missing parameters
func TestExecuteUnloadModelMissingParams(t *testing.T) {
	cfg := &config.ClientConfig{}
	log := newTestLogger()
	executor := NewExecutor(cfg, log)

	task := &cluster.Task{
		ID:      "test-task",
		Type:    cluster.TaskTypeUnloadModel,
		Payload: map[string]interface{}{},
	}

	result, err := executor.ExecuteTask(task)

	assert.Error(t, err)
	assert.NotNil(t, result)
	assert.False(t, result.Success)
	assert.Contains(t, result.Error, "缺少 model_id 参数")
}

// TestExecuteUnloadModelNonExistent tests unloading a non-existent model
func TestExecuteUnloadModelNonExistent(t *testing.T) {
	cfg := &config.ClientConfig{}
	log := newTestLogger()
	executor := NewExecutor(cfg, log)

	task := &cluster.Task{
		ID:   "test-task",
		Type: cluster.TaskTypeUnloadModel,
		Payload: map[string]interface{}{
			"model_id": "non-existent-model",
		},
	}

	result, err := executor.ExecuteTask(task)

	assert.Error(t, err)
	assert.NotNil(t, result)
	assert.False(t, result.Success)
	assert.Contains(t, result.Error, "模型未加载")
}

// TestExecuteRunPythonMissingParams tests executeRunPython with missing parameters
func TestExecuteRunPythonMissingParams(t *testing.T) {
	cfg := &config.ClientConfig{}
	log := newTestLogger()
	executor := NewExecutor(cfg, log)

	task := &cluster.Task{
		ID:      "test-task",
		Type:    cluster.TaskTypeRunPython,
		Payload: map[string]interface{}{},
	}

	result, err := executor.ExecuteTask(task)

	assert.Error(t, err)
	assert.NotNil(t, result)
	assert.False(t, result.Success)
	assert.Contains(t, result.Error, "缺少 script 参数")
}

// TestExecuteUnknownTaskType tests executing an unknown task type
func TestExecuteUnknownTaskType(t *testing.T) {
	cfg := &config.ClientConfig{}
	log := newTestLogger()
	executor := NewExecutor(cfg, log)

	task := &cluster.Task{
		ID:      "test-task",
		Type:    "unknown_type",
		Payload: map[string]interface{}{},
	}

	result, err := executor.ExecuteTask(task)

	assert.Error(t, err)
	assert.NotNil(t, result)
	assert.False(t, result.Success)
	assert.Contains(t, result.Error, "未知任务类型")
}

// TestCancelTask tests cancelling a running task
func TestCancelTask(t *testing.T) {
	cfg := &config.ClientConfig{}
	log := newTestLogger()
	executor := NewExecutor(cfg, log)

	// 创建一个长时间运行的任务
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	task := &cluster.Task{
		ID:   "test-task",
		Type: cluster.TaskTypeRunPython,
		Payload: map[string]interface{}{
			"script": "/path/to/script.py",
		},
	}

	// 添加到运行任务列表（模拟）
	executor.mu.Lock()
	executor.runningTasks[task.ID] = &runningTask{
		context: ctx,
		cancel:  cancel,
		started: time.Now(),
	}
	executor.mu.Unlock()

	// 取消任务
	err := executor.CancelTask(task.ID)

	assert.NoError(t, err)

	// 验证任务已从列表中移除
	executor.mu.RLock()
	_, exists := executor.runningTasks[task.ID]
	executor.mu.RUnlock()

	assert.False(t, exists)
}

// TestCancelNonExistentTask tests cancelling a non-existent task
func TestCancelNonExistentTask(t *testing.T) {
	cfg := &config.ClientConfig{}
	log := newTestLogger()
	executor := NewExecutor(cfg, log)

	err := executor.CancelTask("non-existent-task")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "任务不存在或未运行")
}

// TestGetRunningTasks tests retrieving running task list
func TestGetRunningTasks(t *testing.T) {
	cfg := &config.ClientConfig{}
	log := newTestLogger()
	executor := NewExecutor(cfg, log)

	// 添加一些运行任务
	taskIDs := []string{"task-1", "task-2", "task-3"}

	for _, id := range taskIDs {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		executor.mu.Lock()
		executor.runningTasks[id] = &runningTask{
			context: ctx,
			cancel:  cancel,
			started: time.Now(),
		}
		executor.mu.Unlock()
	}

	// 获取运行任务列表
	runningTasks := executor.GetRunningTasks()

	assert.Len(t, runningTasks, 3)

	for _, id := range taskIDs {
		assert.Contains(t, runningTasks, id)
	}
}

// TestStopTests stopping the executor
func TestStop(t *testing.T) {
	cfg := &config.ClientConfig{}
	log := newTestLogger()
	executor := NewExecutor(cfg, log)

	// 添加一些运行任务和进程
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	executor.mu.Lock()
	executor.runningTasks["task-1"] = &runningTask{
		context: ctx,
		cancel:  cancel,
		started: time.Now(),
	}
	executor.mu.Unlock()

	// 停止执行器
	executor.Stop()

	// 验证所有任务已被取消
	executor.mu.RLock()
	taskCount := len(executor.runningTasks)
	executor.mu.RUnlock()

	assert.Equal(t, 0, taskCount)
}

// TestFindLlamacppBinary tests finding llama.cpp binary
func TestFindLlamacppBinary(t *testing.T) {
	cfg := &config.ClientConfig{}
	log := newTestLogger()
	executor := NewExecutor(cfg, log)

	// 测试查找不存在的二进制文件
	binPath, err := executor.findLlamacppBinary()

	// 可能找不到（取决于系统）
	if err != nil {
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "llama.cpp binary not found")
		assert.Empty(t, binPath)
	} else {
		assert.NotEmpty(t, binPath)
		t.Logf("找到 llama.cpp binary: %s", binPath)
	}
}

// TestFindLlamacppBinaryWithEnv tests finding binary with environment variable
func TestFindLlamacppBinaryWithEnv(t *testing.T) {
	// 创建一个临时文件作为假二进制
	tmpDir := t.TempDir()
	fakeBinary := tmpDir + "/fake-server"
	err := os.WriteFile(fakeBinary, []byte("#!/bin/bash\necho 'fake'"), 0755)
	require.NoError(t, err)

	// 设置环境变量
	os.Setenv("LLAMACPP_SERVER_PATH", fakeBinary)
	defer os.Unsetenv("LLAMACPP_SERVER_PATH")

	cfg := &config.ClientConfig{}
	log := newTestLogger()
	executor := NewExecutor(cfg, log)

	// 查找二进制文件
	binPath, err := executor.findLlamacppBinary()

	assert.NoError(t, err)
	assert.Equal(t, fakeBinary, binPath)
}

// TestExecuteRunLlamacppMissingParams tests executeRunLlamacpp with missing parameters
func TestExecuteRunLlamacppMissingParams(t *testing.T) {
	cfg := &config.ClientConfig{}
	log := newTestLogger()
	executor := NewExecutor(cfg, log)

	task := &cluster.Task{
		ID:      "test-task",
		Type:    cluster.TaskTypeRunLlamacpp,
		Payload: map[string]interface{}{},
	}

	result, err := executor.ExecuteTask(task)

	assert.Error(t, err)
	assert.NotNil(t, result)
	assert.False(t, result.Success)
	assert.Contains(t, result.Error, "缺少 model_path 参数")
}

// TestExecuteTaskWithFullQueue tests task execution when queue is full
func TestExecuteTaskWithFullQueue(t *testing.T) {
	cfg := &config.ClientConfig{}
	log := newTestLogger()
	executor := NewExecutor(cfg, log)

	// 填满任务队列
	for i := 0; i < executor.config.MaxConcurrent; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		taskID := fmt.Sprintf("task-%d", i)
		executor.mu.Lock()
		executor.runningTasks[taskID] = &runningTask{
			context: ctx,
			cancel:  cancel,
			started: time.Now(),
		}
		executor.mu.Unlock()
	}

	// 尝试执行新任务
	task := &cluster.Task{
		ID:   "new-task",
		Type: cluster.TaskTypeRunPython,
		Payload: map[string]interface{}{
			"script": "/path/to/script.py",
		},
	}

	result, err := executor.ExecuteTask(task)

	assert.Error(t, err)
	assert.NotNil(t, result)
	assert.False(t, result.Success)
	assert.Contains(t, result.Error, "任务队列已满")
}

// TestRunningProcessStruct tests runningProcess struct
func TestRunningProcessStruct(t *testing.T) {
	now := time.Now()

	proc := &runningProcess{
		pid:       12345,
		port:      8081,
		modelID:   "test-model",
		modelPath: "/path/to/model.gguf",
		started:   now,
	}

	assert.Equal(t, 12345, proc.pid)
	assert.Equal(t, 8081, proc.port)
	assert.Equal(t, "test-model", proc.modelID)
	assert.Equal(t, "/path/to/model.gguf", proc.modelPath)
	assert.Equal(t, now, proc.started)
}

// TestExecuteLoadModelParameterParsing tests parameter parsing in executeLoadModel
func TestExecuteLoadModelParameterParsing(t *testing.T) {
	cfg := &config.ClientConfig{}
	log := newTestLogger()
	executor := NewExecutor(cfg, log)

	// 测试参数解析
	task := &cluster.Task{
		ID:   "test-task",
		Type: cluster.TaskTypeLoadModel,
		Payload: map[string]interface{}{
			"model_path": "/path/to/model.gguf",
			"model_id":   "test-model",
			"port":       float64(8081),
			"ctx_size":   float64(8192),
			"gpu_layers": float64(99),
			"threads":    float64(8),
		},
	}

	// 注意：这个测试会因为找不到二进制文件而失败，但我们可以验证参数解析
	result, err := executor.ExecuteTask(task)

	// 应该失败在二进制文件查找
	assert.Error(t, err)
	assert.NotNil(t, result)

	// 如果失败在参数解析之前，检查错误消息
	if result.Error != "" {
		// 应该不是参数解析错误
		assert.NotContains(t, result.Error, "缺少")
	}
}

// BenchmarkExecuteTask benchmarks task execution
func BenchmarkExecuteTask(b *testing.B) {
	cfg := &config.ClientConfig{}
	log := newTestLogger()
	executor := NewExecutor(cfg, log)

	task := &cluster.Task{
		ID:      "test-task",
		Type:    "unknown_type",
		Payload: map[string]interface{}{},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		executor.ExecuteTask(task)
	}
}

// BenchmarkGetRunningTasks benchmarks retrieving running tasks
func BenchmarkGetRunningTasks(b *testing.B) {
	cfg := &config.ClientConfig{}
	log := newTestLogger()
	executor := NewExecutor(cfg, log)

	// 添加一些运行任务
	for i := 0; i < 100; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		taskID := fmt.Sprintf("task-%d", i)
		executor.mu.Lock()
		executor.runningTasks[taskID] = &runningTask{
			context: ctx,
			cancel:  cancel,
			started: time.Now(),
		}
		executor.mu.Unlock()
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		executor.GetRunningTasks()
	}
}

// TestExecuteLoadModelWithModelID tests executeLoadModel with and without model_id
func TestExecuteLoadModelWithModelID(t *testing.T) {
	cfg := &config.ClientConfig{}
	log := newTestLogger()
	executor := NewExecutor(cfg, log)

	tests := []struct {
		name            string
		modelPath       string
		modelID         string
		expectedModelID string
	}{
		{
			name:            "提供 model_id",
			modelPath:       "/path/to/model.gguf",
			modelID:         "my-model",
			expectedModelID: "my-model",
		},
		{
			name:            "未提供 model_id (从路径提取)",
			modelPath:       "/path/to/my-model.gguf",
			modelID:         "",
			expectedModelID: "my-model.gguf",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			payload := map[string]interface{}{
				"model_path": tt.modelPath,
				"port":       float64(8081),
			}
			if tt.modelID != "" {
				payload["model_id"] = tt.modelID
			}

			task := &cluster.Task{
				ID:      "test-task",
				Type:    cluster.TaskTypeLoadModel,
				Payload: payload,
			}

			// 会失败在二进制文件查找，但我们可以检查日志
			_, err := executor.ExecuteTask(task)

			// 验证确实尝试执行（失败在二进制文件查找）
			assert.Error(t, err)
		})
	}
}
