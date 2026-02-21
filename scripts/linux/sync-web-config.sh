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

# 配置查找优先级: node > example
SOURCE_CONFIG="$PROJECT_ROOT/config/node/web.config.yaml"
if [ ! -f "$SOURCE_CONFIG" ]; then
    SOURCE_CONFIG="$PROJECT_ROOT/config/example/web.config.yaml"
fi

TARGET_CONFIG="$PROJECT_ROOT/web/public/config.yaml"

echo "📋 同步前端配置文件..."
echo "   源: $SOURCE_CONFIG"
echo "   目标: $TARGET_CONFIG"

# 检查源文件是否存在
if [ ! -f "$SOURCE_CONFIG" ]; then
    echo "❌ 错误: 未找到 web.config.yaml"
    echo "   请在以下位置之一创建配置文件:"
    echo "   - config/node/web.config.yaml (推荐，用于实际运行)"
    echo "   - config/example/web.config.yaml (示例配置)"
    exit 1
fi

# 创建目标目录（如果不存在）
mkdir -p "$(dirname "$TARGET_CONFIG")"

# 复制配置文件
cp "$SOURCE_CONFIG" "$TARGET_CONFIG"

echo "✅ 配置文件同步完成"
