// Package api provides Node API adapter for backward compatibility
// 这个包提供了 API 适配层，确保向后兼容性
package api

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shepherd-project/shepherd/Shepherd/internal/logger"
	"github.com/shepherd-project/shepherd/Shepherd/internal/node"
)

// NodeAdapter 将 API 调用适配到 Node
type NodeAdapter struct {
	node *node.Node
	log  *logger.Logger
}

// NewNodeAdapter 创建一个新的 Node API 适配器
func NewNodeAdapter(n *node.Node, log *logger.Logger) *NodeAdapter {
	return &NodeAdapter{
		node: n,
		log:  log,
	}
}

// ==================== 节点管理 API ====================

// RegisterNode 适配节点注册 API
// POST /api/master/nodes/register
func (a *NodeAdapter) RegisterNode(c *gin.Context) {
	var nodeInfo node.NodeInfo
	if err := c.ShouldBindJSON(&nodeInfo); err != nil {
		a.log.Errorf("解析节点注册请求失败: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "无效的请求格式",
			"message": err.Error(),
		})
		return
	}

	// 验证必需字段
	if nodeInfo.ID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "缺少必需字段",
			"message": "节点ID不能为空",
		})
		return
	}

	if nodeInfo.Address == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "缺少必需字段",
			"message": "节点地址不能为空",
		})
		return
	}

	if nodeInfo.Port <= 0 || nodeInfo.Port > 65535 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "无效的端口",
			"message": "端口必须在1-65535范围内",
		})
		return
	}

	// 调用 Node 的注册方法
	if err := a.node.RegisterClient(&nodeInfo); err != nil {
		a.log.Errorf("注册客户端失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "注册失败",
			"message": err.Error(),
		})
		return
	}

	a.log.Infof("节点注册成功: %s (%s:%d)", nodeInfo.ID, nodeInfo.Address, nodeInfo.Port)
	c.JSON(http.StatusOK, gin.H{
		"message": "节点注册成功",
		"node":    nodeInfo,
	})
}

// ListNodes 返回所有已注册的节点列表
// GET /api/master/nodes
func (a *NodeAdapter) ListNodes(c *gin.Context) {
	clients := a.node.ListClients()
	total, online, offline, busy := a.node.GetClientCount()

	// 转换为切片格式
	nodes := make([]node.NodeInfo, len(clients))
	for i, client := range clients {
		nodes[i] = *client
	}

	c.JSON(http.StatusOK, gin.H{
		"nodes": nodes,
		"stats": gin.H{
			"total":   total,
			"online":  online,
			"offline": offline,
			"busy":    busy,
		},
	})
}

// GetNode 获取指定节点的详细信息
// GET /api/master/nodes/:id
func (a *NodeAdapter) GetNode(c *gin.Context) {
	nodeID := c.Param("id")
	if nodeID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "节点ID不能为空",
		})
		return
	}

	client, err := a.node.GetClient(nodeID)
	if err != nil {
		a.log.Errorf("获取节点失败: %v", err)
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "节点未找到",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"node": client,
	})
}

// UnregisterNode 注销节点
// DELETE /api/master/nodes/:id
func (a *NodeAdapter) UnregisterNode(c *gin.Context) {
	nodeID := c.Param("id")
	if nodeID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "节点ID不能为空",
		})
		return
	}

	if err := a.node.UnregisterClient(nodeID); err != nil {
		a.log.Errorf("注销节点失败: %v", err)
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "注销失败",
			"message": err.Error(),
		})
		return
	}

	a.log.Infof("节点注销成功: %s", nodeID)
	c.JSON(http.StatusOK, gin.H{
		"message": "节点注销成功",
		"nodeID":  nodeID,
	})
}

// ==================== 心跳管理 API ====================

// HandleHeartbeat 适配心跳 API
// POST /api/master/heartbeat
func (a *NodeAdapter) HandleHeartbeat(c *gin.Context) {
	var heartbeat node.HeartbeatMessage
	if err := c.ShouldBindJSON(&heartbeat); err != nil {
		a.log.Errorf("解析心跳请求失败: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "无效的请求格式",
			"message": err.Error(),
		})
		return
	}

	// 验证必需字段
	if heartbeat.NodeID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "缺少必需字段",
			"message": "节点ID不能为空",
		})
		return
	}

	// 调用 Node 的心跳处理
	if err := a.node.HandleHeartbeat(heartbeat.NodeID, &heartbeat); err != nil {
		a.log.Errorf("处理心跳失败: %v", err)
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "心跳处理失败",
			"message": err.Error(),
		})
		return
	}

	a.log.Debugf("心跳处理成功: 节点=%s, 时间=%v", heartbeat.NodeID, heartbeat.Timestamp.Unix())
	c.JSON(http.StatusOK, gin.H{
		"message":   "心跳处理成功",
		"timestamp": time.Now().Unix(),
	})
}

// ==================== 命令管理 API ====================

// SendCommand 向指定节点发送命令
// POST /api/master/nodes/:id/command
func (a *NodeAdapter) SendCommand(c *gin.Context) {
	nodeID := c.Param("id")
	if nodeID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "节点ID不能为空",
		})
		return
	}

	var cmd node.Command
	if err := c.ShouldBindJSON(&cmd); err != nil {
		a.log.Errorf("解析命令请求失败: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "无效的请求格式",
			"message": err.Error(),
		})
		return
	}

	// 设置命令目标节点
	cmd.ToNodeID = nodeID
	if cmd.CreatedAt.IsZero() {
		cmd.CreatedAt = time.Now()
	}

	// 生成命令 ID（如果未设置）
	if cmd.ID == "" {
		cmd.ID = fmt.Sprintf("cmd-%s-%d", nodeID, time.Now().UnixNano())
	}

	// 将命令加入队列
	if err := a.node.QueueCommand(nodeID, &cmd); err != nil {
		a.log.Errorf("命令发送失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "命令发送失败",
			"message": err.Error(),
		})
		return
	}

	a.log.Infof("命令发送成功: 节点=%s, 类型=%s, ID=%s", nodeID, cmd.Type, cmd.ID)
	c.JSON(http.StatusOK, gin.H{
		"message":  "命令发送成功",
		"nodeID":   nodeID,
		"command":  cmd.Type,
		"commandID": cmd.ID,
	})
}

// GetCommands 获取指定节点的待执行命令
// GET /api/master/nodes/:id/commands
func (a *NodeAdapter) GetCommands(c *gin.Context) {
	nodeID := c.Param("id")
	if nodeID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "节点ID不能为空",
		})
		return
	}

	// 从 Node 获取待执行命令
	commands := a.node.GetPendingCommands(nodeID)

	a.log.Debugf("返回待执行命令: 节点=%s, 数量=%d", nodeID, len(commands))
	c.JSON(http.StatusOK, gin.H{
		"success":  true,
		"commands": commands,
	})
}

// ReportCommandResult 上报命令执行结果
// POST /api/master/command/result
func (a *NodeAdapter) ReportCommandResult(c *gin.Context) {
	var req struct {
		NodeID    string      `json:"node_id"`
		CommandID string      `json:"command_id"`
		Success   bool        `json:"success"`
		Output    string      `json:"output"`
		Error     string      `json:"error,omitempty"`
		Metadata  interface{} `json:"metadata,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		a.log.Errorf("解析命令结果请求失败: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "无效的请求格式",
			"message": err.Error(),
		})
		return
	}

	// 验证必需字段
	if req.NodeID == "" || req.CommandID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "缺少必需字段",
			"message": "节点ID和命令ID不能为空",
		})
		return
	}

	// 构建命令结果对象
	result := &node.CommandResult{
		CommandID:   req.CommandID,
		FromNodeID:  req.NodeID,
		ToNodeID:    a.node.GetID(),
		Success:     req.Success,
		CompletedAt: time.Now(),
	}

	// 添加结果数据
	if req.Output != "" {
		if result.Result == nil {
			result.Result = make(map[string]interface{})
		}
		result.Result["output"] = req.Output
	}

	if req.Error != "" {
		result.Error = req.Error
	}

	if req.Metadata != nil {
		// 转换 metadata 为字符串映射
		if metadataMap, ok := req.Metadata.(map[string]interface{}); ok {
			result.Metadata = make(map[string]string)
			for k, v := range metadataMap {
				result.Metadata[k] = fmt.Sprintf("%v", v)
			}
		}
	}

	// 计算执行时长（毫秒）
	startTime := time.Now()

	// 存储命令结果到持久化存储
	if err := a.node.StoreCommandResult(result); err != nil {
		a.log.Errorf("存储命令结果失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "存储命令结果失败",
			"message": err.Error(),
		})
		return
	}

	result.Duration = time.Since(startTime).Milliseconds()

	a.log.Infof("命令结果已存储: 节点=%s, 命令=%s, 成功=%v, 耗时=%dms",
		req.NodeID, req.CommandID, req.Success, result.Duration)

	if req.Error != "" {
		a.log.Errorf("命令执行失败: %s", req.Error)
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "命令结果已记录",
	})
}

// ==================== 路由注册 ====================

// RegisterRoutes 注册所有 API 路由
// 这个方法保持了与 MasterHandler 相同的路由结构
func (a *NodeAdapter) RegisterRoutes(router *gin.RouterGroup) {
	master := router.Group("/master")
	{
		// 节点管理
		nodes := master.Group("/nodes")
		{
			nodes.POST("/register", a.RegisterNode)
			nodes.GET("", a.ListNodes)
			nodes.GET("/:id", a.GetNode)
			nodes.DELETE("/:id", a.UnregisterNode)
			nodes.POST("/:id/command", a.SendCommand)
			nodes.GET("/:id/commands", a.GetCommands)
		}

		// 心跳
		master.POST("/heartbeat", a.HandleHeartbeat)

		// 命令结果上报
		master.POST("/command/result", a.ReportCommandResult)
	}

	a.log.Infof("Node API 适配器路由已注册")
}

// ==================== 辅助方法 ====================

// GetNodeInstance 返回底层 Node 实例
// 供需要直接访问 Node 的场景使用
func (a *NodeAdapter) GetNodeInstance() *node.Node {
	return a.node
}

// SetNodeInstance 设置底层 Node 实例
func (a *NodeAdapter) SetNodeInstance(n *node.Node) {
	a.node = n
}
