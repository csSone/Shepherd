# MasterConnector 实现学习笔记

## 实现概述

实现了 `internal/client/connector.go` 和 `internal/client/connector_test.go`，用于 Client 连接 Master 并维持心跳。

## 关键设计决策

### 1. 架构设计

```
┌─────────────────────────────────────────────────────────────┐
│                    MasterConnector                          │
├─────────────────────────────────────────────────────────────┤
│  属性:                                                       │
│  - nodeID: 节点唯一标识                                      │
│  - masterAddr: Master 服务地址                               │
│  - heartbeatMgr: HeartbeatManager (来自 node 包)            │
│  - executor: CommandExecutor (来自 node 包)                 │
├─────────────────────────────────────────────────────────────┤
│  功能:                                                       │
│  - Connect(): 向 Master 注册节点                            │
│  - Disconnect(): 优雅断开连接                               │
│  - 自动心跳维持                                              │
│  - 命令轮询和执行                                           │
│  - 断线重连机制                                             │
└─────────────────────────────────────────────────────────────┘
```

### 2. 核心方法

| 方法 | 功能 |
|------|------|
| `Connect()` | 向 Master 注册，带重试机制 |
| `Disconnect()` | 优雅断开，注销节点，停止心跳 |
| `IsConnected()` | 检查连接状态（通过 HeartbeatManager） |
| `IsRegistered()` | 检查是否已注册 |
| `UpdateNodeInfo()` | 更新节点信息 |
| `SetCommandHandler()` | 设置自定义命令处理器 |

### 3. 重连策略

- **指数退避**: 延迟 = `backoff * 2^(attempt-1)`，最大 60s
- **抖动**: ±25% 随机抖动避免惊群效应
- **重试次数**: 可配置，默认 10 次

### 4. 命令处理流程

```
┌──────────────┐     ┌──────────────┐     ┌──────────────┐
│  Master API  │────▶│   Connector  │────▶│   Executor   │
│  GET /cmds   │     │  poll loop   │     │   Execute    │
└──────────────┘     └──────────────┘     └──────────────┘
        ▲                                           │
        │              ┌──────────────┐             │
        └──────────────│ POST /result │◀────────────┘
                       └──────────────┘
```

## 集成点

### 与 node 包的集成

```go
// HeartbeatManager - 发送定期心跳
heartbeatMgr *node.HeartbeatManager

// CommandExecutor - 执行 Master 下发的命令
executor *node.CommandExecutor

// NodeInfo - 节点信息
nodeInfo *node.NodeInfo
```

### Master API 端点

| 端点 | 方法 | 用途 |
|------|------|------|
| `/api/master/nodes/register` | POST | 节点注册 |
| `/api/master/nodes/:id/unregister` | POST | 节点注销 |
| `/api/master/heartbeat` | POST | 心跳发送 |
| `/api/master/nodes/:id/commands` | GET | 获取待执行命令 |
| `/api/master/command/result` | POST | 上报执行结果 |

## 测试覆盖

创建了全面的单元测试：

1. **TestNewMasterConnector**: 测试配置验证
2. **TestMasterConnector_Connect**: 测试连接、重试、失败场景
3. **TestMasterConnector_Disconnect**: 测试优雅断开
4. **TestMasterConnector_CommandExecution**: 测试命令执行和结果上报
5. **TestMasterConnector_SetCommandHandler**: 测试自定义处理器
6. **TestMasterConnector_UpdateNodeInfo**: 测试节点信息更新
7. **TestMasterConnector_Getters**: 测试 getter 方法
8. **TestCalculateBackoff**: 测试退避算法
9. **BenchmarkMasterConnector_Connect**: 性能基准测试

## 编码规范遵循

- ✅ 中文注释（符合项目规范）
- ✅ 错误包装使用 `%w`
- ✅ 结构体标签使用反引号
- ✅ 上下文控制使用 `context.Context`
- ✅ 并发安全使用 `sync.RWMutex`
- ✅ 使用 testify 进行测试断言

## 注意事项

1. **不修改 node 包**: 只使用已有接口
2. **HTTP 客户端**: 使用 `net/http` 作为客户端（不实现服务器）
3. **回调机制**: 通过回调而非直接操作 Node 状态
4. **资源清理**: Disconnect() 确保所有 goroutine 和资源被清理

## 后续优化建议

1. 考虑使用 WebSocket 替代 HTTP 轮询命令
2. 添加指标收集（连接成功率、命令执行时间等）
3. 支持 TLS 加密通信
4. 添加连接质量监控（延迟、丢包率）
