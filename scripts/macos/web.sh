#!/bin/bash
# Shepherd Web 前端运行脚本 (macOS)

set -e

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 获取项目根目录
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$(dirname "$SCRIPT_DIR")")"
WEB_DIR="${PROJECT_DIR}/web"

# 打印带颜色的消息
print_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 显示帮助信息
show_help() {
    cat << EOF
🐏 Shepherd Web 前端 (macOS)

用法: $0 [命令]

命令:
    dev         启动开发服务器 (默认)
    build       构建生产版本
    preview     预览生产构建
    install     安装/重新安装依赖
    clean       清理构建文件
    fix         修复依赖问题
    check       检查依赖状态

选项:
    -h, --help     显示此帮助信息
    -p, --port     指定开发服务器端口 (默认: 3000)

示例:
    $0 dev                 # 启动开发服务器
    $0 dev -p 4000         # 在端口 4000 启动
    $0 build              # 构建生产版本
    $0 preview            # 预览构建结果

EOF
}

# 检查依赖
check_dependencies() {
    if ! command -v node &> /dev/null; then
        print_error "Node.js 未安装"
        print_info "推荐使用 Homebrew 安装: brew install node"
        exit 1
    fi

    if ! command -v npm &> /dev/null; then
        print_error "npm 未安装"
        exit 1
    fi

    # 显示版本
    NODE_VERSION=$(node --version)
    NPM_VERSION=$(npm --version)
    print_info "Node.js: ${NODE_VERSION}"
    print_info "npm: ${NPM_VERSION}"
}

# 切换到 web 目录
cd_web() {
    if [ ! -d "${WEB_DIR}" ]; then
        print_error "Web 目录不存在: ${WEB_DIR}"
        exit 1
    fi
    cd "${WEB_DIR}"
}

# 安装依赖
install_deps() {
    print_info "安装依赖..."
    cd_web

    if [ -f "package-lock.json" ]; then
        npm ci
    else
        npm install
    fi

    print_success "依赖安装完成"
}

# 启动开发服务器
dev_server() {
    local PORT=${1:-3000}
    print_info "启动开发服务器 (端口: ${PORT})..."
    cd_web

    # 同步配置
    print_info "同步配置..."
    # 尝试从 linux 目录同步配置（如果在 macOS 上也可以使用）
    SYNC_SCRIPT="${PROJECT_DIR}/scripts/linux/sync-web-config.sh"
    if [ -f "$SYNC_SCRIPT" ]; then
        bash "$SYNC_SCRIPT"
    else
        print_warning "配置同步脚本不存在，跳过同步"
    fi

    # 启动开发服务器
    npm run dev -- --port "${PORT}"
}

# 构建生产版本
build_prod() {
    print_info "构建生产版本..."
    cd_web

    npm run build
    print_success "构建完成"
}

# 预览生产构建
preview_prod() {
    print_info "预览生产构建..."
    cd_web

    if [ ! -d "dist" ]; then
        print_error "构建目录不存在，请先运行: $0 build"
        exit 1
    fi

    # 使用简单的 HTTP 服务器预览
    if ! command -v npx &> /dev/null; then
        print_error "npx 不可用"
        exit 1
    fi

    print_info "启动预览服务器..."
    npx --yes serve dist -l 4173
}

# 清理构建文件
clean_build() {
    print_info "清理构建文件..."
    cd_web

    rm -rf dist node_modules/.vite
    print_success "清理完成"
}

# 修复依赖
fix_deps() {
    print_info "修复依赖问题..."
    cd_web

    # 清理并重新安装
    rm -rf node_modules package-lock.json
    npm install

    print_success "修复完成"
}

# 检查依赖状态
check_status() {
    print_info "检查依赖状态..."
    cd_web

    if [ ! -d "node_modules" ]; then
        print_warning "依赖未安装"
        return 1
    fi

    # 检查关键依赖
    print_info "已安装的包:"
    npm list --depth=0
}

# 主函数
main() {
    local COMMAND=""
    local PORT="3000"

    # 解析参数
    while [[ $# -gt 0 ]]; do
        case $1 in
            dev|build|preview|install|clean|fix|check)
                COMMAND="$1"
                shift
                ;;
            -p|--port)
                PORT="$2"
                shift 2
                ;;
            -h|--help)
                show_help
                exit 0
                ;;
            *)
                print_error "未知参数: $1"
                show_help
                exit 1
                ;;
        esac
    done

    # 默认命令
    if [ -z "$COMMAND" ]; then
        COMMAND="dev"
    fi

    # 检查依赖
    check_dependencies

    # 执行命令
    case "$COMMAND" in
        dev)
            dev_server "$PORT"
            ;;
        build)
            build_prod
            ;;
        preview)
            preview_prod
            ;;
        install)
            install_deps
            ;;
        clean)
            clean_build
            ;;
        fix)
            fix_deps
            ;;
        check)
            check_status
            ;;
        *)
            print_error "未知命令: $COMMAND"
            show_help
            exit 1
            ;;
    esac
}

# 运行主函数
main "$@"
