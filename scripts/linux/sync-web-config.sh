#!/bin/bash
#
# sync-web-config.sh - 同步前端配置文件
#
# 将 config/web.config.yaml 同步到 web/public/config.yaml
# 这样前端可以读取统一的配置文件

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# 从 linux/ 子目录向上两级到达项目根目录
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

SOURCE_CONFIG="$PROJECT_ROOT/config/node/web.config.yaml"

TARGET_CONFIG="$PROJECT_ROOT/web/public/config.yaml"

echo "📋 同步前端配置文件..."
echo "   源: $SOURCE_CONFIG"
echo "   目标: $TARGET_CONFIG"

# 检查源文件是否存在
if [ ! -f "$SOURCE_CONFIG" ]; then
    echo "❌ 错误: 源配置文件不存在: $SOURCE_CONFIG"
    exit 1
fi

# 创建目标目录（如果不存在）
mkdir -p "$(dirname "$TARGET_CONFIG")"

# 复制配置文件
cp "$SOURCE_CONFIG" "$TARGET_CONFIG"

echo "✅ 配置文件同步完成"
