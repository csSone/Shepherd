#!/bin/bash
#
# watch-sync-config.sh - 监视配置文件变化并自动同步
#
# 使用 inotifywait 监视 config/web.config.yaml 的变化
# 当文件发生变化时，自动同步到 web/public/config.yaml

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# 从 linux/ 子目录向上两级到达项目根目录
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

SOURCE_CONFIG="$PROJECT_ROOT/config/web.config.yaml"
TARGET_CONFIG="$PROJECT_ROOT/web/public/config.yaml"

echo "👀 监视配置文件变化..."
echo "   源: $SOURCE_CONFIG"
echo "   目标: $TARGET_CONFIG"
echo ""
echo "按 Ctrl+C 停止监视"
echo "----------------------------------------"

# 检查是否安装了 inotifywait
if ! command -v inotifywait &> /dev/null; then
    echo "❌ 错误: inotifywait 未安装"
    echo ""
    echo "请安装 inotify-tools:"
    echo "  Ubuntu/Debian: sudo apt-get install inotify-tools"
    echo "  CentOS/RHEL:   sudo yum install inotify-tools"
    echo "  macOS:         brew install fswatch"
    exit 1
fi

# 初始同步
echo "📋 执行初始同步..."
"$SCRIPT_DIR/sync-web-config.sh"
echo "✅ 初始同步完成"
echo ""

# 监视文件变化
inotifywait -m -e modify,create,move "$SOURCE_CONFIG" 2>/dev/null | while read -r directory events filename; do
    echo ""
    echo "📝 检测到配置文件变化: $events $filename"
    echo "⏰ $(date '+%Y-%m-%d %H:%M:%S')"

    # 等待文件写入完成（有些编辑器会触发多次事件）
    sleep 0.5

    # 同步配置
    if "$SCRIPT_DIR/sync-web-config.sh"; then
        echo "✅ 配置已自动同步"
        echo ""
        echo "💡 提示: 前端开发服务器会自动重新加载配置"
        echo "   如果没有自动生效，请刷新浏览器（Ctrl+Shift+R 强制刷新）"
    else
        echo "❌ 同步失败"
    fi

    echo "----------------------------------------"
done
