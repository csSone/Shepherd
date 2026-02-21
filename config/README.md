# Shepherd 配置目录说明

本目录包含 Shepherd 服务器的配置文件。

## 目录结构

```
config/
├── example/                    # 标准示例配置（可直接复制使用）
│   ├── standalone.config.yaml  # 单机模式示例
│   ├── master.config.yaml      # Master 模式示例
│   └── client.config.yaml      # Client 模式示例
│
├── node/                       # 实际运行的节点配置
│   ├── standalone.config.yaml  # 单机模式配置
│   ├── master.config.yaml      # Master 模式配置
│   └── client.config.yaml      # Client 模式配置
│
├── server.config.yaml          # 当前使用的配置（向后兼容）
├── master.config.yaml          # Master 配置（向后兼容）
├── client.config.yaml          # Client 配置（向后兼容）
├── models.json                 # 模型配置文件
└── web.config.yaml             # Web 前端配置

```

## 使用说明

### 1. 快速开始

选择适合你需求的配置模板：

- **单机使用** → 使用 `example/standalone.config.yaml`
- **部署 Master** → 使用 `example/master.config.yaml`
- **部署 Client** → 使用 `example/client.config.yaml`

### 2. 配置文件加载规则

Shepherd 按以下优先级加载配置：

1. 环境变量 `SHEPHERD_CONFIG_DIR` 指定的目录
2. 默认目录 `./config`
3. 根据运行模式加载对应的配置文件：
   - `standalone` 模式 → `server.config.yaml`
   - `master` 模式 → `master.config.yaml`
   - `client` 模式 → `client.config.yaml`

### 3. 配置文件迁移

新的统一配置格式支持在一个文件中配置所有模式。旧的 `master.config.yaml` 和 `client.config.yaml` 仍可继续使用，系统会自动迁移到新格式。

### 4. 多节点部署

在 `node/` 目录中可以存放多个节点的配置：

```bash
# 启动不同模式的节点
./shepherd standalone --config config/node/standalone.config.yaml
./shepherd master --config config/node/master.config.yaml
./shepherd client --config config/node/client.config.yaml
```

## 配置项说明

### 关键配置项

- `mode`: 运行模式（standalone/master/client）
- `node.role`: 节点角色（standalone/master/client/hybrid）
- `server.web_port`: Web UI/API 端口
- `model.paths`: 模型扫描路径
- `llamacpp.paths`: llama.cpp 后端路径

详细配置说明请参考各配置文件中的注释。

## 安全建议

1. **生产环境**请设置 `security.api_key`
2. **限制 CORS** 来源，不要设置为 `*`
3. **启用 SSL** 配置证书和密钥
4. **定期备份** `models.json` 和数据库文件

## 更多信息

- 项目文档: [README.md](../README.md)
- API 文档: [doc/api.md](../doc/api.md)
- 部署指南: [doc/deployment.md](../doc/deployment.md)
