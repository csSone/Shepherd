// Package executor provides task execution functionality for client nodes.
package executor

import (
	"context"
	"fmt"
	"os/exec"
	"sync"
	"time"

	"github.com/shepherd-project/shepherd/Shepherd/internal/client"
	"github.com/shepherd-project/shepherd/Shepherd/internal/cluster"
	"github.com/shepherd-project/shepherd/Shepherd/internal/config"
	"github.com/shepherd-project/shepherd/Shepherd/internal/logger"
)

// Executor executes tasks on the client node
type Executor struct {
	config *client.ExecutorConfig
	mu     sync.RWMutex
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
	log    *logger.Logger

	// Running tasks tracking
	runningTasks map[string]*runningTask
}

// runningTask represents a currently running task
type runningTask struct {
	context context.Context
	cancel  context.CancelFunc
	started time.Time
}

// NewExecutor creates a new task executor
func NewExecutor(cfg *config.ClientConfig, log *logger.Logger) *Executor {
	ctx, cancel := context.WithCancel(context.Background())

	execConfig := &client.ExecutorConfig{
		CondaPath:     cfg.CondaEnv.CondaPath,
		CondaEnvs:     cfg.CondaEnv.Environments,
		LlamacppPaths: []string{}, // TODO: Get from config
		TaskTimeout:   5 * time.Minute,
		MaxConcurrent: 4,
	}

	return &Executor{
		config:       execConfig,
		ctx:          ctx,
		cancel:       cancel,
		wg:           sync.WaitGroup{},
		log:          log,
		runningTasks: make(map[string]*runningTask),
	}
}

// Start starts the executor
func (e *Executor) Start() {
	// No background tasks needed for now
}

// Stop stops the executor and cancels all running tasks
func (e *Executor) Stop() {
	e.cancel()

	e.mu.Lock()
	defer e.mu.Unlock()

	// Cancel all running tasks
	for _, task := range e.runningTasks {
		task.cancel()
	}

	e.wg.Wait()
}

// ExecuteTask executes a task
func (e *Executor) ExecuteTask(task *cluster.Task) (*client.TaskResult, error) {
	startTime := time.Now()

	e.log.Info(fmt.Sprintf("开始执行任务: %s (类型: %s)", task.ID, task.Type))

	// Check task limit
	e.mu.Lock()
	if len(e.runningTasks) >= e.config.MaxConcurrent {
		e.mu.Unlock()
		return &client.TaskResult{
			TaskID:      task.ID,
			Success:     false,
			Error:       "任务队列已满",
			CompletedAt: time.Now(),
			Duration:    time.Since(startTime).Milliseconds(),
		}, fmt.Errorf("任务队列已满")
	}

	// Create task context
	taskCtx, taskCancel := context.WithTimeout(e.ctx, e.config.TaskTimeout)

	// Track running task
	e.runningTasks[task.ID] = &runningTask{
		context: taskCtx,
		cancel:  taskCancel,
		started: startTime,
	}
	e.mu.Unlock()

	// Execute task based on type
	var result *client.TaskResult
	var err error

	switch task.Type {
	case cluster.TaskTypeLoadModel:
		result, err = e.executeLoadModel(taskCtx, task)
	case cluster.TaskTypeUnloadModel:
		result, err = e.executeUnloadModel(taskCtx, task)
	case cluster.TaskTypeRunPython:
		result, err = e.executeRunPython(taskCtx, task)
	case cluster.TaskTypeRunLlamacpp:
		result, err = e.executeRunLlamacpp(taskCtx, task)
	default:
		result = &client.TaskResult{
			TaskID:      task.ID,
			Success:     false,
			Error:       fmt.Sprintf("未知任务类型: %s", task.Type),
			CompletedAt: time.Now(),
			Duration:    time.Since(startTime).Milliseconds(),
		}
		err = fmt.Errorf("未知任务类型: %s", task.Type)
	}

	// Remove from running tasks
	e.mu.Lock()
	delete(e.runningTasks, task.ID)
	e.mu.Unlock()

	result.CompletedAt = time.Now()
	result.Duration = time.Since(startTime).Milliseconds()

	if err != nil {
		e.log.Error(fmt.Sprintf("任务执行失败: %s - %v", task.ID, err))
	} else {
		e.log.Info(fmt.Sprintf("任务执行成功: %s", task.ID))
	}

	return result, err
}

// executeLoadModel executes a load model task
func (e *Executor) executeLoadModel(ctx context.Context, task *cluster.Task) (*client.TaskResult, error) {
	// Extract parameters
	modelPath, ok := task.Payload["model_path"].(string)
	if !ok {
		return &client.TaskResult{
			TaskID:  task.ID,
			Success: false,
			Error:   "缺少 model_path 参数",
		}, fmt.Errorf("缺少 model_path 参数")
	}

	// This is a placeholder - actual implementation would use the process manager
	// to start llama.cpp with the model
	e.log.Info(fmt.Sprintf("加载模型: %s", modelPath))

	return &client.TaskResult{
		TaskID:  task.ID,
		Success: true,
		Result: map[string]interface{}{
			"model_path": modelPath,
			"loaded":     true,
		},
		Output: "模型加载成功",
	}, nil
}

// executeUnloadModel executes an unload model task
func (e *Executor) executeUnloadModel(ctx context.Context, task *cluster.Task) (*client.TaskResult, error) {
	// Extract parameters
	modelID, ok := task.Payload["model_id"].(string)
	if !ok {
		return &client.TaskResult{
			TaskID:  task.ID,
			Success: false,
			Error:   "缺少 model_id 参数",
		}, fmt.Errorf("缺少 model_id 参数")
	}

	e.log.Info(fmt.Sprintf("卸载模型: %s", modelID))

	return &client.TaskResult{
		TaskID:  task.ID,
		Success: true,
		Result: map[string]interface{}{
			"model_id": modelID,
			"unloaded": true,
		},
		Output: "模型卸载成功",
	}, nil
}

// executeRunPython executes a Python script task
func (e *Executor) executeRunPython(ctx context.Context, task *cluster.Task) (*client.TaskResult, error) {
	// Extract parameters
	script, ok := task.Payload["script"].(string)
	if !ok {
		return &client.TaskResult{
			TaskID:  task.ID,
			Success: false,
			Error:   "缺少 script 参数",
		}, fmt.Errorf("缺少 script 参数")
	}

	condaEnv := ""
	if env, ok := task.Payload["conda_env"].(string); ok {
		condaEnv = env
	}

	e.log.Info(fmt.Sprintf("运行Python脚本: %s (环境: %s)", script, condaEnv))

	// Build command
	var cmd *exec.Cmd
	if condaEnv != "" && e.config.CondaPath != "" {
		envPath, exists := e.config.CondaEnvs[condaEnv]
		if !exists {
			return &client.TaskResult{
				TaskID:  task.ID,
				Success: false,
				Error:   fmt.Sprintf("Conda环境不存在: %s", condaEnv),
			}, fmt.Errorf("conda环境不存在: %s", condaEnv)
		}
		cmd = exec.CommandContext(ctx, envPath+"/bin/python", script)
	} else {
		cmd = exec.CommandContext(ctx, "python", script)
	}

	// Run command
	output, err := cmd.CombinedOutput()
	if err != nil {
		return &client.TaskResult{
			TaskID:  task.ID,
			Success: false,
			Error:   err.Error(),
			Output:  string(output),
		}, err
	}

	return &client.TaskResult{
		TaskID:  task.ID,
		Success: true,
		Result: map[string]interface{}{
			"exit_code": 0,
		},
		Output: string(output),
	}, nil
}

// executeRunLlamacpp executes a llama.cpp task
func (e *Executor) executeRunLlamacpp(ctx context.Context, task *cluster.Task) (*client.TaskResult, error) {
	// Extract parameters
	modelPath, ok := task.Payload["model_path"].(string)
	if !ok {
		return &client.TaskResult{
			TaskID:  task.ID,
			Success: false,
			Error:   "缺少 model_path 参数",
		}, fmt.Errorf("缺少 model_path 参数")
	}

	prompt := ""
	if p, ok := task.Payload["prompt"].(string); ok {
		prompt = p
	}

	e.log.Info(fmt.Sprintf("运行llama.cpp: %s", modelPath))

	// This is a placeholder - actual implementation would:
	// 1. Find llama.cpp binary
	// 2. Start with model and parameters
	// 3. Capture output

	return &client.TaskResult{
		TaskID:  task.ID,
		Success: true,
		Result: map[string]interface{}{
			"model_path": modelPath,
			"prompt":     prompt,
		},
		Output: "llama.cpp执行完成",
	}, nil
}

// CancelTask cancels a running task
func (e *Executor) CancelTask(taskID string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	task, exists := e.runningTasks[taskID]
	if !exists {
		return fmt.Errorf("任务不存在或未运行: %s", taskID)
	}

	task.cancel()
	delete(e.runningTasks, taskID)

	e.log.Info(fmt.Sprintf("取消任务: %s", taskID))

	return nil
}

// GetRunningTasks returns the list of running task IDs
func (e *Executor) GetRunningTasks() []string {
	e.mu.RLock()
	defer e.mu.RUnlock()

	tasks := make([]string, 0, len(e.runningTasks))
	for id := range e.runningTasks {
		tasks = append(tasks, id)
	}
	return tasks
}
