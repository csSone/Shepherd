#!/bin/bash
# 同步配置文件到 public 目录

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"
WEB_DIR="${PROJECT_DIR}/web"
CONFIG_FILE="${WEB_DIR}/config.yaml"
PUBLIC_DIR="${WEB_DIR}/public"
TARGET_FILE="${PUBLIC_DIR}/config.yaml"

# 检查源配置文件是否存在
if [ ! -f "$CONFIG_FILE" ]; then
    echo "错误: 配置文件不存在: $CONFIG_FILE"
    exit 1
fi

# 确保目标目录存在
mkdir -p "$PUBLIC_DIR"

# 复制配置文件
cp "$CONFIG_FILE" "$TARGET_FILE"

echo "配置文件已同步到: $TARGET_FILE"