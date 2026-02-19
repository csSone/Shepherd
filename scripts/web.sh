#!/bin/bash
# Shepherd Web 前端运行脚本

set -e

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 获取项目根目录
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"
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
🐏 Shepherd Web 前端

用法: $0 [命令] [选项]

命令:
    dev         启动开发服务器 (默认)
    build       构建生产版本
    preview     预览生产构建
    install     安装依赖
    clean       清理构建文件

选项:
    -h, --help     显示此帮助信息
    -p, --port PORT    指定端口 (开发模式默认: 3000)

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
        exit 1
    fi

    if ! command -v npm &> /dev/null; then
        print_error "npm 未安装"
        exit 1
    fi

    if [ ! -d "${WEB_DIR}/node_modules" ]; then
        print_warning "依赖未安装，正在安装..."
        install_dependencies
    fi
}

# 安装依赖
install_dependencies() {
    print_info "安装 Web 前端依赖..."
    cd "$WEB_DIR"
    npm install
    print_success "依赖安装完成"
}

# 清理构建文件
clean_build() {
    print_info "清理 Web 构建文件..."
    cd "$WEB_DIR"
    rm -rf dist node_modules/.vite
    print_success "清理完成"
}

# 启动开发服务器
run_dev() {
    local port=${1:-3000}
    print_info "启动 Web 开发服务器 (端口: $port)..."
    cd "$WEB_DIR"
    exec npm run dev -- --port "$port"
}

# 构建生产版本
run_build() {
    print_info "构建 Web 生产版本..."
    cd "$WEB_DIR"
    npm run build
    print_success "构建完成，输出目录: web/dist/"
}

# 预览生产构建
run_preview() {
    print_info "预览 Web 生产构建..."
    cd "$WEB_DIR"
    exec npm run preview
}

# 主函数
main() {
    local command=""
    local port=""

    # 解析参数
    while [[ $# -gt 0 ]]; do
        case $1 in
            dev|build|preview|install|clean)
                command="$1"
                shift
                ;;
            -h|--help)
                show_help
                exit 0
                ;;
            -p|--port)
                port="$2"
                shift 2
                ;;
            *)
                print_error "未知参数: $1"
                show_help
                exit 1
                ;;
        esac
    done

    # 默认命令
    if [ -z "$command" ]; then
        command="dev"
    fi

    # 检查依赖
    check_dependencies

    # 执行命令
    case "$command" in
        dev)
            run_dev "$port"
            ;;
        build)
            run_build
            ;;
        preview)
            run_preview
            ;;
        install)
            install_dependencies
            ;;
        clean)
            clean_build
            ;;
    esac
}

# 运行主函数
main "$@"
