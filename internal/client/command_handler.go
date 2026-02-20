// Package client provides client node functionality for the distributed architecture.
// CommandHandler 处理 Master 下发的各种命令
package client

import (
	"context"
	"fmt"
	"os/exec"
	"time"

	"github.com/shepherd-project/shepherd/Shepherd/internal/logger"
	"github.com/shepherd-project/shepherd/Shepherd/internal/model"
	"github.com/shepherd-project/shepherd/Shepherd/internal/node"
	"github.com/shepherd-project/shepherd/Shepherd/internal/process"
)

// CommandHandler 处理从 Master 接收到的命令
type CommandHandler struct {
	// executor 用于执行基础命令
	executor *node.CommandExecutor
	// modelManager 管理模型加载和卸载
	modelManager *model.Manager
	// processManager 管理进程生命周期
	processManager *process.Manager
	// logger 日志记录器
	logger *logger.Logger
	// nodeID 当前节点的 ID
	nodeID string
}

// CommandHandlerConfig 包含 CommandHandler 的配置
type CommandHandlerConfig struct {
	// Executor 基础命令执行器
	Executor *node.CommandExecutor
	// ModelManager 模型管理器
	ModelManager *model.Manager
	// ProcessManager 进程管理器
	ProcessManager *process.Manager
	// Logger 日志记录器
	Logger *logger.Logger
	// NodeID 当前节点 ID
	NodeID string
}

// NewCommandHandler 创建一个新的命令处理器
func NewCommandHandler(config *CommandHandlerConfig) *CommandHandler {
	if config == nil {
		config = &CommandHandlerConfig{}
	}

	return &CommandHandler{
		executor:       config.Executor,
		modelManager:   config.ModelManager,
		processManager: config.ProcessManager,
		logger:         config.Logger,
		nodeID:         config.NodeID,
	}
}

// Handle 处理单个命令并返回结果
// 这是 CommandHandler 的主要入口方法
func (ch *CommandHandler) Handle(command *node.Command) (*node.CommandResult, error) {
	if command == nil {
		return nil, fmt.Errorf("命令不能为空")
	}

	if ch.logger != nil {
		ch.logger.Info(fmt.Sprintf("处理命令: %s (类型: %s)", command.ID, command.Type))
	}

	startTime := time.Now()

	// 创建结果对象
	result := &node.CommandResult{
		CommandID:   command.ID,
		FromNodeID:  ch.nodeID,
		ToNodeID:    command.FromNodeID,
		CompletedAt: time.Now(),
	}

	// 根据命令类型分发处理
	var err error
	switch command.Type {
	case node.CommandTypeLoadModel:
		err = ch.handleLoadModel(command, result)
	case node.CommandTypeUnloadModel:
		err = ch.handleUnloadModel(command, result)
	case node.CommandTypeRunLlamacpp:
		err = ch.handleRunLlamacpp(command, result)
	case node.CommandTypeStopProcess:
		err = ch.handleStopProcess(command, result)
	case node.CommandTypeScanModels:
		err = ch.handleScanModels(command, result)
	default:
		result.Success = false
		result.Error = fmt.Sprintf("未知的命令类型: %s", command.Type)
		err = fmt.Errorf("未知的命令类型: %s", command.Type)
	}

	// 计算执行时长
	result.Duration = time.Since(startTime).Milliseconds()
	result.CompletedAt = time.Now()

	if err != nil && result.Error == "" {
		result.Error = err.Error()
	}

	if ch.logger != nil {
		if result.Success {
			ch.logger.Info(fmt.Sprintf("命令执行成功: %s (耗时: %dms)", command.ID, result.Duration))
		} else {
			ch.logger.Error(fmt.Sprintf("命令执行失败: %s - %s", command.ID, result.Error))
		}
	}

	return result, err
}

// handleLoadModel 处理加载模型命令
// Payload 期望包含:
//   - model_id: 模型 ID (必需)
//   - ctx_size: 上下文大小 (可选)
//   - gpu_layers: GPU 层数 (可选)
//   - threads: 线程数 (可选)
//   - batch_size: 批次大小 (可选)
//   - temperature: 温度参数 (可选)
//   - top_p: top_p 参数 (可选)
//   - top_k: top_k 参数 (可选)
func (ch *CommandHandler) handleLoadModel(command *node.Command, result *node.CommandResult) error {
	modelID, ok := command.Payload["model_id"].(string)
	if !ok || modelID == "" {
		result.Success = false
		result.Error = "缺少必需的参数: model_id"
		return fmt.Errorf("缺少必需的参数: model_id")
	}

	if ch.modelManager == nil {
		result.Success = false
		result.Error = "模型管理器未初始化"
		return fmt.Errorf("模型管理器未初始化")
	}

	// 构建加载请求
	req := &model.LoadRequest{
		ModelID: modelID,
	}

	// 解析可选参数
	if ctxSize, ok := command.Payload["ctx_size"].(float64); ok {
		req.CtxSize = int(ctxSize)
	}
	if gpuLayers, ok := command.Payload["gpu_layers"].(float64); ok {
		req.GPULayers = int(gpuLayers)
	}
	if threads, ok := command.Payload["threads"].(float64); ok {
		req.Threads = int(threads)
	}
	if batchSize, ok := command.Payload["batch_size"].(float64); ok {
		req.BatchSize = int(batchSize)
	}
	if temperature, ok := command.Payload["temperature"].(float64); ok {
		req.Temperature = temperature
	}
	if topP, ok := command.Payload["top_p"].(float64); ok {
		req.TopP = topP
	}
	if topK, ok := command.Payload["top_k"].(float64); ok {
		req.TopK = int(topK)
	}
	if repeatPenalty, ok := command.Payload["repeat_penalty"].(float64); ok {
		req.RepeatPenalty = repeatPenalty
	}
	if nPredict, ok := command.Payload["n_predict"].(float64); ok {
		req.NPredict = int(nPredict)
	}

	// 执行加载
	loadResult, err := ch.modelManager.Load(req)
	if err != nil {
		result.Success = false
		result.Error = fmt.Sprintf("加载模型失败: %v", err)
		return err
	}

	result.Success = loadResult.Success
	if loadResult.Success {
		result.Result = map[string]interface{}{
			"model_id": loadResult.ModelID,
			"port":     loadResult.Port,
			"ctx_size": loadResult.CtxSize,
			"duration": loadResult.Duration.Milliseconds(),
		}
	} else {
		result.Error = loadResult.Error.Error()
	}

	return nil
}

// handleUnloadModel 处理卸载模型命令
// Payload 期望包含:
//   - model_id: 模型 ID (必需)
func (ch *CommandHandler) handleUnloadModel(command *node.Command, result *node.CommandResult) error {
	modelID, ok := command.Payload["model_id"].(string)
	if !ok || modelID == "" {
		result.Success = false
		result.Error = "缺少必需的参数: model_id"
		return fmt.Errorf("缺少必需的参数: model_id")
	}

	if ch.modelManager == nil {
		result.Success = false
		result.Error = "模型管理器未初始化"
		return fmt.Errorf("模型管理器未初始化")
	}

	// 执行卸载
	err := ch.modelManager.Unload(modelID)
	if err != nil {
		result.Success = false
		result.Error = fmt.Sprintf("卸载模型失败: %v", err)
		return err
	}

	result.Success = true
	result.Result = map[string]interface{}{
		"model_id": modelID,
		"unloaded": true,
	}

	return nil
}

// handleRunLlamacpp 直接运行 llama.cpp 命令
// Payload 期望包含:
//   - binary_path: llama.cpp 可执行文件路径 (必需)
//   - model_path: 模型文件路径 (必需)
//   - args: 额外参数列表 (可选)
//   - timeout: 超时时间(秒) (可选)
func (ch *CommandHandler) handleRunLlamacpp(command *node.Command, result *node.CommandResult) error {
	binaryPath, ok := command.Payload["binary_path"].(string)
	if !ok || binaryPath == "" {
		result.Success = false
		result.Error = "缺少必需的参数: binary_path"
		return fmt.Errorf("缺少必需的参数: binary_path")
	}

	modelPath, ok := command.Payload["model_path"].(string)
	if !ok || modelPath == "" {
		result.Success = false
		result.Error = "缺少必需的参数: model_path"
		return fmt.Errorf("缺少必需的参数: model_path")
	}

	// 构建命令参数
	args := []string{"-m", modelPath}

	// 添加额外参数
	if extraArgs, ok := command.Payload["args"].([]interface{}); ok {
		for _, arg := range extraArgs {
			if strArg, ok := arg.(string); ok {
				args = append(args, strArg)
			}
		}
	}

	// 设置超时
	timeout := 300 * time.Second
	if timeoutSec, ok := command.Payload["timeout"].(float64); ok && timeoutSec > 0 {
		timeout = time.Duration(timeoutSec) * time.Second
	}
	if command.Timeout != nil {
		timeout = *command.Timeout
	}

	// 创建上下文
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// 执行命令
	cmd := exec.CommandContext(ctx, binaryPath, args...)
	output, err := cmd.CombinedOutput()

	if ctx.Err() == context.DeadlineExceeded {
		result.Success = false
		result.Error = "命令执行超时"
		result.Result = map[string]interface{}{
			"output": string(output),
		}
		return fmt.Errorf("命令执行超时")
	}

	if err != nil {
		result.Success = false
		result.Error = fmt.Sprintf("命令执行失败: %v", err)
		result.Result = map[string]interface{}{
			"output":    string(output),
			"exit_code": cmd.ProcessState.ExitCode(),
		}
		return err
	}

	result.Success = true
	result.Result = map[string]interface{}{
		"output":    string(output),
		"exit_code": cmd.ProcessState.ExitCode(),
	}

	return nil
}

// handleStopProcess 处理停止进程命令
// Payload 期望包含:
//   - model_id: 模型 ID (必需)
//   - force: 是否强制停止 (可选，默认 false)
func (ch *CommandHandler) handleStopProcess(command *node.Command, result *node.CommandResult) error {
	modelID, ok := command.Payload["model_id"].(string)
	if !ok || modelID == "" {
		result.Success = false
		result.Error = "缺少必需的参数: model_id"
		return fmt.Errorf("缺少必需的参数: model_id")
	}

	if ch.processManager == nil {
		result.Success = false
		result.Error = "进程管理器未初始化"
		return fmt.Errorf("进程管理器未初始化")
	}

	// 检查是否强制停止
	force := false
	if f, ok := command.Payload["force"].(bool); ok {
		force = f
	}

	// 获取进程信息
	_, exists := ch.processManager.Get(modelID)
	if !exists {
		result.Success = false
		result.Error = fmt.Sprintf("进程不存在: %s", modelID)
		return fmt.Errorf("进程不存在: %s", modelID)
	}

	// 停止进程 - Process.Stop() 内部已实现优雅关闭和强制终止逻辑
	err := ch.processManager.Stop(modelID)
	if err != nil {
		result.Success = false
		result.Error = fmt.Sprintf("停止进程失败: %v", err)
		return err
	}

	result.Success = true
	result.Result = map[string]interface{}{
		"model_id": modelID,
		"stopped":  true,
		"force":    force,
	}

	return nil
}

// handleScanModels 处理扫描模型命令
// Payload 可选包含:
//   - paths: 要扫描的路径列表 (可选，默认使用配置的扫描路径)
func (ch *CommandHandler) handleScanModels(command *node.Command, result *node.CommandResult) error {
	if ch.modelManager == nil {
		result.Success = false
		result.Error = "模型管理器未初始化"
		return fmt.Errorf("模型管理器未初始化")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// 执行扫描
	scanResult, err := ch.modelManager.Scan(ctx)
	if err != nil {
		result.Success = false
		result.Error = fmt.Sprintf("扫描模型失败: %v", err)
		return err
	}

	// 构建模型信息列表
	models := make([]map[string]interface{}, 0, len(scanResult.Models))
	for _, m := range scanResult.Models {
		modelInfo := map[string]interface{}{
			"id":   m.ID,
			"name": m.Name,
			"path": m.Path,
			"size": m.Size,
		}
		if m.Metadata != nil {
			modelInfo["architecture"] = m.Metadata.Architecture
			modelInfo["context_length"] = m.Metadata.ContextLength
			modelInfo["embedding_length"] = m.Metadata.EmbeddingLength
		}
		models = append(models, modelInfo)
	}

	// 构建错误信息列表
	errors := make([]map[string]string, 0, len(scanResult.Errors))
	for _, e := range scanResult.Errors {
		errors = append(errors, map[string]string{
			"path":  e.Path,
			"error": e.Error,
		})
	}

	result.Success = true
	result.Result = map[string]interface{}{
		"models_found":  len(scanResult.Models),
		"models":        models,
		"errors":        errors,
		"error_count":   len(scanResult.Errors),
		"duration_ms":   scanResult.Duration.Milliseconds(),
		"total_files":   scanResult.TotalFiles,
		"matched_files": scanResult.MatchedFiles,
		"scanned_at":    scanResult.ScannedAt.Format(time.RFC3339),
	}

	return nil
}

// GetModelManager 返回模型管理器
func (ch *CommandHandler) GetModelManager() *model.Manager {
	return ch.modelManager
}

// GetProcessManager 返回进程管理器
func (ch *CommandHandler) GetProcessManager() *process.Manager {
	return ch.processManager
}

// GetExecutor 返回命令执行器
func (ch *CommandHandler) GetExecutor() *node.CommandExecutor {
	return ch.executor
}

// SetNodeID 设置节点 ID
func (ch *CommandHandler) SetNodeID(nodeID string) {
	ch.nodeID = nodeID
}

// GetNodeID 获取节点 ID
func (ch *CommandHandler) GetNodeID() string {
	return ch.nodeID
}
