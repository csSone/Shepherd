#!/bin/bash
# Shepherd 运行脚本
# 支持 standalone, master, client 三种模式

set -e

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 获取脚本所在目录
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"
BUILD_DIR="${PROJECT_DIR}/build"
BINARY_NAME="shepherd"

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
🐏 Shepherd 运行脚本

用法: $0 [模式] [选项]

模式:
    standalone     单机模式 (默认)
    master         Master 模式 - 管理多个 Client 节点
    client         Client 模式 - 作为工作节点

通用选项:
    -h, --help     显示此帮助信息
    -b, --build    运行前先编译
    -v, --version  显示版本信息

Master 模式选项:
    --port PORT    Web 服务器端口 (默认: 9190)
    --scan         启动时自动扫描网络

Client 模式选项:
    --master URL   Master 地址 (必需)
    --name NAME    Client 名称 (可选)
    --tags TAGS    Client 标签，逗号分隔 (可选)

示例:
    # 单机模式
    $0 standalone

    # Master 模式
    $0 master --port 9190 --scan

    # Client 模式
    $0 client --master http://192.168.1.100:9190 --name client-1

    # 运行前先编译
    $0 master -b

EOF
}

# 检查二进制文件是否存在
check_binary() {
    if [ ! -f "${BUILD_DIR}/${BINARY_NAME}" ]; then
        print_warning "二进制文件不存在: ${BUILD_DIR}/${BINARY_NAME}"
        read -p "是否现在编译? (y/N): " -n 1 -r
        echo
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            (cd "${SCRIPT_DIR}" && ./build.sh)
        else
            print_error "无法继续，请先编译项目"
            exit 1
        fi
    fi
}

# 显示版本信息
show_version() {
    if [ -f "${BUILD_DIR}/${BINARY_NAME}" ]; then
        "${BUILD_DIR}/${BINARY_NAME}" --version
    else
        print_error "二进制文件不存在，请先编译"
        exit 1
    fi
    exit 0
}

# 主函数
main() {
    local MODE=""
    local BUILD_FIRST=false
    local MASTER_ADDR=""
    local CLIENT_NAME=""
    local CLIENT_TAGS=""
    local WEB_PORT="9190"
    local AUTO_SCAN=false

    # 解析参数
    while [[ $# -gt 0 ]]; do
        case $1 in
            standalone|master|client)
                MODE="$1"
                shift
                ;;
            -h|--help)
                show_help
                exit 0
                ;;
            -b|--build)
                BUILD_FIRST=true
                shift
                ;;
            -v|--version)
                show_version
                ;;
            --master)
                MASTER_ADDR="$2"
                shift 2
                ;;
            --name)
                CLIENT_NAME="$2"
                shift 2
                ;;
            --tags)
                CLIENT_TAGS="$2"
                shift 2
                ;;
            --port)
                WEB_PORT="$2"
                shift 2
                ;;
            --scan)
                AUTO_SCAN=true
                shift
                ;;
            *)
                print_error "未知参数: $1"
                show_help
                exit 1
                ;;
        esac
    done

    # 默认模式
    if [ -z "$MODE" ]; then
        MODE="standalone"
    fi

    # 编译（如果需要）
    if [ "$BUILD_FIRST" = true ]; then
        print_info "编译项目..."
        (cd "${SCRIPT_DIR}" && ./build.sh)
        print_success "编译完成"
    fi

    # 检查二进制文件
    check_binary

    # 构建命令参数
    local ARGS=()
    ARGS+=("--mode=${MODE}")

    case "$MODE" in
        master)
            print_info "启动 Master 模式..."
            ARGS+=("--master-addr=0.0.0.0:${WEB_PORT}")

            if [ "$AUTO_SCAN" = true ]; then
                print_info "启用自动网络扫描"
            fi
            ;;
        client)
            if [ -z "$MASTER_ADDR" ]; then
                print_error "Client 模式需要指定 Master 地址 (--master)"
                print_info "示例: $0 client --master http://192.168.1.100:9190"
                exit 1
            fi
            print_info "启动 Client 模式..."
            print_info "Master 地址: ${MASTER_ADDR}"
            ARGS+=("--master-address=${MASTER_ADDR}")

            if [ -n "$CLIENT_NAME" ]; then
                print_info "Client 名称: ${CLIENT_NAME}"
                export SHEPHERD_CLIENT_NAME="$CLIENT_NAME"
            fi

            if [ -n "$CLIENT_TAGS" ]; then
                print_info "Client 标签: ${CLIENT_TAGS}"
                export SHEPHERD_CLIENT_TAGS="$CLIENT_TAGS"
            fi
            ;;
        standalone)
            print_info "启动单机模式..."
            ;;
    esac

    # 显示启动信息
    echo ""
    echo "=========================================="
    echo "  🐏 Shepherd v${MODE}"
    echo "=========================================="
    echo "  模式: ${MODE}"
    if [ "$MODE" = "master" ]; then
        echo "  端口: ${WEB_PORT}"
    fi
    if [ "$MODE" = "client" ]; then
        echo "  Master: ${MASTER_ADDR}"
    fi
    echo "=========================================="
    echo ""

    # 启动程序
    cd "${PROJECT_DIR}"
    exec "${BUILD_DIR}/${BINARY_NAME}" "${ARGS[@]}"
}

# 运行主函数
main "$@"
