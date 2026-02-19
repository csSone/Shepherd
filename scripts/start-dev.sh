#!/bin/bash
#
# start-dev.sh - 启动前端开发服务器
#
# 此脚本会：
# 1. 同步最新配置
# 2. 检查端口占用
# 3. 启动 Vite 开发服务器

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
WEB_DIR="$PROJECT_ROOT/web"

echo "🚀 启动 Shepherd 前端开发服务器"
echo "========================================"

# 1. 同步配置
echo "📋 同步配置文件..."
"$SCRIPT_DIR/sync-web-config.sh"

# 2. 读取配置的端口
echo ""
echo "🔍 读取配置..."
PORT=$(grep -oP 'port:\s*\K\d+' "$WEB_DIR/public/config.yaml" || echo "3000")
echo "   配置端口: $PORT"

# 3. 检查端口占用
if lsof -Pi :$PORT -sTCP:LISTEN -t >/dev/null 2>&1; then
    echo ""
    echo "⚠️  警告: 端口 $PORT 已被占用"
    echo "   正在尝试停止占用进程..."

    # 尝试停止占用端口的进程
    lsof -ti :$PORT -sTCP:LISTEN | xargs -r kill -9 2>/dev/null || true
    sleep 1

    if lsof -Pi :$PORT -sTCP:LISTEN -t >/dev/null 2>&1; then
        echo "❌ 无法停止占用端口的进程，请手动停止:"
        lsof -Pi :$PORT -sTCP:LISTEN -t
        echo ""
        echo "然后重新运行此脚本"
        exit 1
    else
        echo "✅ 已停止占用端口的进程"
    fi
fi

# 4. 进入 web 目录
cd "$WEB_DIR"

# 5. 启动开发服务器
echo ""
echo "🚀 启动 Vite 开发服务器 (端口: $PORT)..."
echo "   访问地址:"
echo "   - http://localhost:$PORT"
echo "   - http://10.0.0.193:$PORT"
echo ""
echo "按 Ctrl+C 停止服务器"
echo "========================================"

# 启动 Vite（端口由 vite.config.ts 从配置文件读取）
npm run dev
