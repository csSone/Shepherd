// Package master provides HTTP handlers for master node management.
//
// Deprecated: This package is part of the old distributed architecture.
// New code should use github.com/shepherd-project/shepherd/Shepherd/internal/api.NodeAdapter instead.
// This package will be removed in a future release.
package master

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shepherd-project/shepherd/Shepherd/internal/logger"
	"github.com/shepherd-project/shepherd/Shepherd/internal/node"
)

// MasterHandler provides HTTP handlers for master node operations
//
// Deprecated: Use api.NodeAdapter instead. This type will be removed in a future release.
type MasterHandler struct {
	nodeManager *NodeManager
	log         *logger.Logger
}

// NewMasterHandler creates a new master handler
func NewMasterHandler(nodeManager *NodeManager, log *logger.Logger) *MasterHandler {
	return &MasterHandler{
		nodeManager: nodeManager,
		log:         log,
	}
}

// RegisterRoutes registers all master API routes
func (h *MasterHandler) RegisterRoutes(router *gin.RouterGroup) {
	master := router.Group("/master")
	{
		// Node management
		nodes := master.Group("/nodes")
		{
			nodes.POST("/register", h.handleRegisterNode)
			nodes.GET("", h.handleListNodes)
			nodes.GET("/:id", h.handleGetNode)
			nodes.DELETE("/:id", h.handleUnregisterNode)
			nodes.POST("/:id/command", h.handleSendCommand)
		}

		// Heartbeat
		master.POST("/heartbeat", h.handleHeartbeat)
	}
}

// handleRegisterNode handles node registration
func (h *MasterHandler) handleRegisterNode(c *gin.Context) {
	var nodeInfo node.NodeInfo
	if err := c.ShouldBindJSON(&nodeInfo); err != nil {
		h.log.Errorf("解析节点注册请求失败: %v", err)
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

	// 注册节点
	if err := h.nodeManager.RegisterNode(&nodeInfo); err != nil {
		h.log.Errorf("注册节点失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "注册节点失败",
			"message": err.Error(),
		})
		return
	}

	h.log.Infof("节点注册成功: %s (%s:%d)", nodeInfo.ID, nodeInfo.Address, nodeInfo.Port)
	c.JSON(http.StatusOK, gin.H{
		"message": "节点注册成功",
		"node":    nodeInfo,
	})
}

// handleHeartbeat handles heartbeat messages from nodes
func (h *MasterHandler) handleHeartbeat(c *gin.Context) {
	var heartbeat node.HeartbeatMessage
	if err := c.ShouldBindJSON(&heartbeat); err != nil {
		h.log.Errorf("解析心跳请求失败: %v", err)
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

	// 处理心跳
	if err := h.nodeManager.HandleHeartbeat(heartbeat.NodeID, &heartbeat); err != nil {
		h.log.Errorf("处理心跳失败: %v", err)
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "处理心跳失败",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "心跳处理成功",
		"timestamp": time.Now().Unix(),
	})
}

// handleListNodes returns a list of all registered nodes
func (h *MasterHandler) handleListNodes(c *gin.Context) {
	nodes := h.nodeManager.ListNodes()

	// 获取统计信息
	total, online, offline, busy := h.nodeManager.GetNodeCount()

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

// handleGetNode returns information about a specific node
func (h *MasterHandler) handleGetNode(c *gin.Context) {
	nodeID := c.Param("id")
	if nodeID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "缺少必需参数",
			"message": "节点ID不能为空",
		})
		return
	}

	nodeInfo, err := h.nodeManager.GetNode(nodeID)
	if err != nil {
		h.log.Errorf("获取节点信息失败: %v", err)
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "获取节点信息失败",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"node": nodeInfo,
	})
}

// handleUnregisterNode removes a node from the registry
func (h *MasterHandler) handleUnregisterNode(c *gin.Context) {
	nodeID := c.Param("id")
	if nodeID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "缺少必需参数",
			"message": "节点ID不能为空",
		})
		return
	}

	if err := h.nodeManager.UpdateNodeStatus(nodeID, node.NodeStatusOffline); err != nil {
		h.log.Errorf("注销节点失败: %v", err)
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "注销节点失败",
			"message": err.Error(),
		})
		return
	}

	h.log.Infof("节点注销成功: %s", nodeID)
	c.JSON(http.StatusOK, gin.H{
		"message": "节点注销成功",
		"nodeId":  nodeID,
	})
}

// handleSendCommand sends a command to a specific node
func (h *MasterHandler) handleSendCommand(c *gin.Context) {
	nodeID := c.Param("id")
	if nodeID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "缺少必需参数",
			"message": "节点ID不能为空",
		})
		return
	}

	var cmd node.Command
	if err := c.ShouldBindJSON(&cmd); err != nil {
		h.log.Errorf("解析命令请求失败: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "无效的请求格式",
			"message": err.Error(),
		})
		return
	}

	// 验证必需字段
	if cmd.Type == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "缺少必需字段",
			"message": "命令类型不能为空",
		})
		return
	}

	// 设置目标节点ID和创建时间
	cmd.ToNodeID = nodeID
	cmd.CreatedAt = time.Now()

	// 生成命令ID（简单实现，生产环境可以使用UUID）
	cmd.ID = "cmd-" + strconv.FormatInt(time.Now().UnixNano(), 10)

	h.log.Infof("发送命令到节点 %s: %s", nodeID, cmd.Type)

	c.JSON(http.StatusOK, gin.H{
		"message": "命令发送成功",
		"command": cmd,
	})
}
