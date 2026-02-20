# 模型模块完整修复计划

## 📋 **当前问题诊断**

基于代码审查，发现以下关键问题：

### 🔴 **严重问题**

#### 1. **API 响应未正确序列化模型数据**
- `handleListModels()` 直接返回 `modelMgr.ListModels()`
- Model 结构体包含 `*gguf.Metadata` 指针类型
- Gin 的 JSON 序列化可能无法正确处理嵌套指针

#### 2. **配置路径未正确加载到 Model Manager**
- `NewManager()` 调用 `loadModels()` 加载已保存模型
- 但配置中的 `PathConfigs` 可能未正确传递给扫描器
- `getScanPaths()` 函数可能返回空列表

#### 3. **模型数据存储和加载不一致**
- `saveModels()` 保存到配置文件
- `loadModels()` 从配置文件加载
- 但扫描后的新模型可能没有正确保存

### 🟡 **中等问题**

#### 4. **缺少调试日志**
- 模型扫描流程没有日志输出
- 无法判断扫描是否执行、找到多少模型

#### 5. **前端刷新问题**
- 扫描后前端列表未自动刷新
- 需要手动刷新页面才能看到新模型

---

## 🔧 **修复计划**

### **Phase 1: 添加调试日志** (30分钟) ✅ 已完成

**目标**: 在关键位置添加日志，追踪数据流

**修改文件**: `internal/model/manager.go`

**任务**:
1. 在 `Scan()` 函数中添加日志：
   - 扫描开始时的路径列表
   - 每个路径扫描完成后的模型数量
   - 扫描完成后的总模型数
   - 保存模型时的数量

2. 在 `loadModels()` 中添加日志：
   - 加载配置时的模型数量
   - 配置路径列表

3. 在 `NewManager()` 中添加日志：
   - 初始化时的配置路径

**代码示例**:
```go
func (m *Manager) Scan(ctx context.Context) (*ScanResult, error) {
    // ... 原有代码 ...
    
    scanPaths := m.getScanPaths()
    logger.Infof("开始扫描模型路径: %v", scanPaths)
    
    for _, scanPath := range scanPaths {
        logger.Infof("正在扫描路径: %s", scanPath)
        pathModels, pathErrors := m.scanPath(ctx, scanPath)
        logger.Infof("路径 %s 扫描完成: 找到 %d 个模型, %d 个错误", 
            scanPath, len(pathModels), len(pathErrors))
        // ...
    }
    
    logger.Infof("模型扫描完成: 总共 %d 个模型", len(result.Models))
    return result, nil
}
```

**验证方式**:
```bash
./scripts/run.sh standalone 2>&1 | grep -i model
```

---

### **Phase 2: 修复 API 响应序列化** (45分钟)

**目标**: 确保模型数据能正确序列化为 JSON

**修改文件**: `internal/server/server.go`

**任务**:
1. 创建 API 响应用的 Model DTO（Data Transfer Object）
2. 修改 `handleListModels()` 使用 DTO
3. 处理 `gguf.Metadata` 的序列化

**代码修改**:

创建 DTO 结构体:
```go
// ModelDTO 用于 API 响应的模型数据结构
type ModelDTO struct {
    ID          string                 `json:"id"`
    Name        string                 `json:"name"`
    DisplayName string                 `json:"displayName"`
    Alias       string                 `json:"alias"`
    Path        string                 `json:"path"`
    PathPrefix  string                 `json:"pathPrefix"`
    Size        int64                  `json:"size"`
    Favourite   bool                   `json:"favourite"`
    Metadata    map[string]interface{} `json:"metadata"`
    Status      string                 `json:"status"`
    IsLoaded    bool                   `json:"isLoaded"`
}

func convertModelToDTO(m *model.Model, status *model.ModelStatus) ModelDTO {
    dto := ModelDTO{
        ID:          m.ID,
        Name:        m.Name,
        DisplayName: m.DisplayName,
        Alias:       m.Alias,
        Path:        m.Path,
        PathPrefix:  m.PathPrefix,
        Size:        m.Size,
        Favourite:   m.Favourite,
        Status:      "stopped",
        IsLoaded:    false,
    }
    
    // 转换 metadata
    if m.Metadata != nil {
        dto.Metadata = convertMetadataToMap(m.Metadata)
    }
    
    // 添加状态信息
    if status != nil {
        dto.Status = status.State.String()
        dto.IsLoaded = status.State == model.StateLoaded
    }
    
    return dto
}

func convertMetadataToMap(meta *gguf.Metadata) map[string]interface{} {
    return map[string]interface{}{
        "name":             meta.Name,
        "architecture":     meta.Architecture,
        "quantization":     meta.Quantization,
        "contextLength":    meta.ContextLength,
        "embeddingLength":  meta.EmbeddingLength,
        "blockSize":        meta.BlockSize,
        "layerCount":       meta.LayerCount,
        "attentionHeads":   meta.AttentionHeadCount,
    }
}
```

修改 handler:
```go
func (s *Server) handleListModels(c *gin.Context) {
    models := s.modelMgr.ListModels()
    statuses := s.modelMgr.ListStatus()
    
    var dtos []ModelDTO
    for _, m := range models {
        status, _ := statuses[m.ID]
        dtos = append(dtos, convertModelToDTO(m, status))
    }
    
    c.JSON(http.StatusOK, gin.H{
        "models": dtos, 
        "total": len(dtos),
    })
}
```

**验证方式**:
```bash
curl -s http://10.0.0.193:9190/api/models | jq .
```

---

### **Phase 3: 修复配置路径加载** (30分钟)

**目标**: 确保配置路径正确传递给 Model Manager

**修改文件**: `internal/model/manager.go`, `cmd/shepherd/main.go`

**任务**:
1. 在 `NewManager()` 中添加配置路径检查
2. 确保 `getScanPaths()` 正确工作
3. 在初始化时如果没有路径，记录警告

**代码修改**:

```go
func NewManager(cfg *config.Config, cfgMgr *config.Manager, procMgr *process.Manager) *Manager {
    // ... 原有代码 ...
    
    // 检查配置路径
    paths := m.getScanPaths()
    if len(paths) == 0 {
        logger.Warn("模型管理器初始化: 未配置模型扫描路径")
    } else {
        logger.Infof("模型管理器初始化: 配置路径 %v", paths)
    }
    
    return m
}
```

在 main.go 中添加配置检查:
```go
// 在初始化 modelMgr 之后
if len(cfg.Model.PathConfigs) == 0 && len(cfg.Model.Paths) == 0 {
    logger.Warn("警告: 未配置模型路径，请在设置中配置")
}
```

---

### **Phase 4: 修复前端刷新** (20分钟)

**目标**: 扫描后自动刷新模型列表

**修改文件**: `web/src/features/models/hooks.ts`, `web/src/pages/models/index.tsx`

**任务**:
1. 修改 `useScanModels()` hook，在成功后刷新列表
2. 添加扫描进度显示

**代码修改**:

```typescript
export function useScanModels() {
  const queryClient = useQueryClient();
  const [progress, setProgress] = useState(0);

  return useMutation({
    mutationFn: async () => {
      const response = await apiClient.post<ScanResult>('/scan');
      return response;
    },
    onSuccess: () => {
      // 扫描成功后刷新模型列表
      queryClient.invalidateQueries({ queryKey: ['models'] });
      // 显示成功消息
      toast.success(`扫描完成，找到 ${data.models_found} 个模型`);
    },
    onError: (error) => {
      toast.error(`扫描失败: ${error.message}`);
    },
  });
}
```

---

### **Phase 5: 修复模型存储** (30分钟)

**目标**: 确保扫描后的模型正确保存

**修改文件**: `internal/model/manager.go`

**任务**:
1. 检查 `saveModels()` 是否正确调用
2. 确保配置管理器正确保存

**代码修改**:

在 `Scan()` 函数中增强保存逻辑:
```go
// Update models map
m.mu.Lock()
for _, model := range result.Models {
    m.models[model.ID] = model
    logger.Debugf("添加模型到缓存: %s (%s)", model.ID, model.Name)
}
m.mu.Unlock()

// Save to config
if err := m.saveModels(); err != nil {
    logger.Errorf("保存模型配置失败: %v", err)
} else {
    logger.Infof("已保存 %d 个模型到配置", len(result.Models))
}
```

---

## 🧪 **测试计划**

### **单元测试**

1. **测试模型扫描**:
```bash
cd /home/user/workspace/Shepherd
go test ./internal/model -run TestScan -v
```

2. **测试 API 响应**:
```bash
go test ./internal/server -run TestModel -v
```

### **集成测试**

1. **完整流程测试**:
```bash
# 1. 启动服务
./scripts/run.sh standalone

# 2. 检查初始状态
curl http://10.0.0.193:9190/api/models

# 3. 配置路径（通过 API 或配置文件）
curl -X POST http://10.0.0.193:9190/api/config/models/paths \
  -H "Content-Type: application/json" \
  -d '{"path": "/home/user/workspace/LlamacppServer/build/models"}'

# 4. 触发扫描
curl -X POST http://10.0.0.193:9190/api/scan

# 5. 检查扫描状态
curl http://10.0.0.193:9190/api/scan/status

# 6. 验证模型列表
curl http://10.0.0.193:9190/api/models | jq '.models | length'
```

---

## 📊 **时间表**

| Phase | 任务 | 预估时间 | 优先级 |
|-------|------|----------|--------|
| Phase 1 | 添加调试日志 | 30分钟 | 🔴 高 |
| Phase 2 | 修复 API 响应 | 45分钟 | 🔴 高 |
| Phase 3 | 修复配置路径 | 30分钟 | 🟡 中 |
| Phase 4 | 修复前端刷新 | 20分钟 | 🟡 中 |
| Phase 5 | 修复模型存储 | 30分钟 | 🟡 中 |
| **总计** | | **155分钟 (~2.5小时)** | |

---

## ✅ **验收标准**

1. **日志输出**: 启动和扫描时有清晰的日志
2. **API 测试**: curl 能正确返回模型列表
3. **前端显示**: 页面能显示模型卡片
4. **扫描功能**: 点击扫描按钮后能发现新模型
5. **数据持久化**: 重启服务后模型列表不丢失

---

## 🚨 **风险与应对**

| 风险 | 可能性 | 应对措施 |
|------|--------|----------|
| JSON 序列化仍有错误 | 中 | 使用 DTO 模式，避免指针类型 |
| 配置文件格式不兼容 | 低 | 检查版本兼容性，添加迁移代码 |
| 前端类型不匹配 | 中 | 同步更新 TypeScript 类型定义 |
| 并发问题 | 低 | 使用 sync.RWMutex 保护数据 |

---

**下一步**: 开始执行 Phase 1，添加调试日志以诊断具体问题。