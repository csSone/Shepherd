# 🎉 Shepherd v0.1.3 Release Notes

## ✨ 新功能

### 路径配置功能
- **Llama.cpp 路径配置**: 通过 Web UI 配置多个 llama.cpp 安装路径
  - 支持自定义名称和描述
  - 路径有效性自动验证
  - 适用于多 Llama.cpp 环境管理

- **模型路径配置**: 灵活管理模型扫描目录
  - 配置多个模型扫描路径
  - 自动扫描和发现 GGUF 模型
  - 便于组织和管理分散的模型文件

## 🎨 UI 优化

### 设置页面改进
- 整体布局更加紧凑，空间利用更高效
- 关于页面现已居中显示
- 减小了间距和内边距
- 优化了文字和图标大小
- 更好的视觉比例

### 暗色模式支持
- 修复了聊天页面在暗色模式下的显示问题
- 所有组件现已支持主题变量
- 统一的颜色系统

## 🔧 技术改进

### 后端 (Go)
- 新增路径配置 API 端点
  - `GET/POST/DELETE /api/config/llamacpp/paths`
  - `GET/POST/PUT/DELETE /api/config/models/paths`
  - `POST /api/config/llamacpp/test`
- 完善的路径验证和安全检查
- 配置持久化支持

### 前端 (TypeScript/React)
- 新增路径配置组件
  - `PathConfigPanel` - 路径配置面板
  - `PathEditDialog` - 路径编辑对话框
  - `PathItem` - 路径列表项
  - `Dialog` - 通用对话框组件
- 改进的 API 客户端
- 更好的类型安全

## 📝 文档更新
- 更新 README 添加路径配置功能说明
- 更新许可证为 Apache License 2.0
- 添加 API 端点参考

## 📦 版本信息
- **版本号**: v0.1.2 → v0.1.3
- **Go 版本**: 1.25+
- **React 版本**: 19.x
- **许可证**: Apache License 2.0

## 🚀 快速开始

### 从源码编译
```bash
git clone https://github.com/shepherd-project/shepherd.git
cd shepherd
make build
```

### 运行
```bash
# 单机模式
./build/shepherd standalone

# 访问 Web UI
open http://localhost:9190
```

## 🔗 相关链接

- **完整更新日志**: https://github.com/shepherd-project/shepherd/compare/v0.1.2...v0.1.3
- **文档**: https://github.com/shepherd-project/shepherd/tree/main/docs
- **问题反馈**: https://github.com/shepherd-project/shepherd/issues
- **功能建议**: https://github.com/shepherd-project/shepherd/discussions

---

**⭐ 如果这个项目对你有帮助，请点个 Star！**

**🐏 Shepherd - 高性能轻量级 llama.cpp 模型管理系统**
