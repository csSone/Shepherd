#!/bin/bash
# Shepherd macOS 运行脚本
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
PROJECT_DIR="$(dirname "$(dirname "$SCRIPT_DIR")")"
BUILD_DIR="${PROJECT_DIR}/build"
BINARY_NAME="shepherd-darwin-$(uname -m)"

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
🐏 Shepherd 运行脚本 (macOS)

用法: $0 [模式] [选项]

模式:
    standalone     单机模式 (默认)
    master         Master 模式 - 管理多个 Client 节点
    client         Client 模式 - 作为工作节点

通用选项:
    -h, --help     显示此帮助信息
    -b, --build    运行前先编译
    -v, --version  显示版本信息
    --config PATH  指定配置文件路径

Client 模式选项:
    --master URL   Master 地址 (可选，也可从配置文件读取)
    --name NAME    Client 名称 (可选)
    --tags TAGS    Client 标签，逗号分隔 (可选)

macOS 特定选项:
    --no-gatekeeper   跳过 Gatekeeper 验证（解决隔离问题）

示例:
    # 单机模式
    $0 standalone

    # Master 模式
    $0 master

    # Client 模式（从命令行指定 Master 地址）
    $0 client --master http://192.168.1.100:9190 --name client-1

    # Client 模式（从配置文件读取 Master 地址）
    $0 client

    # 运行前先编译
    $0 standalone -b

    # 使用自定义配置文件
    $0 standalone --config config/node/standalone.config.yaml

    # 跳过 Gatekeeper 验证
    $0 standalone --no-gatekeeper

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

# 修复 Gatekeeper 隔离问题
fix_gatekeeper() {
    if [ -f "${BUILD_DIR}/${BINARY_NAME}" ]; then
        print_info "修复 Gatekeeper 隔离..."
        xattr -cr "${BUILD_DIR}/${BINARY_NAME}"
        print_success "修复完成"
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
    local FIX_GATEKEEPER=false
    local CONFIG_PATH=""

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
            --config)
                CONFIG_PATH="$2"
                shift 2
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
            --no-gatekeeper)
                FIX_GATEKEEPER=true
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

    # 修复 Gatekeeper（如果需要）
    if [ "$FIX_GATEKEEPER" = true ]; then
        fix_gatekeeper
    fi

    # 自动检测 config/node 下的配置文件
    if [ -z "$CONFIG_PATH" ]; then
        local NODE_CONFIG="${PROJECT_DIR}/config/node/${MODE}.config.yaml"
        if [ -f "$NODE_CONFIG" ]; then
            CONFIG_PATH="$NODE_CONFIG"
            print_info "使用配置文件: ${CONFIG_PATH}"
        else
            print_info "未找到 node 配置文件，使用默认配置"
        fi
    else
        # 验证自定义配置文件是否存在
        if [ ! -f "$CONFIG_PATH" ]; then
            # 尝试相对于项目目录的路径
            if [ -f "${PROJECT_DIR}/${CONFIG_PATH}" ]; then
                CONFIG_PATH="${PROJECT_DIR}/${CONFIG_PATH}"
            else
                print_error "配置文件不存在: ${CONFIG_PATH}"
                exit 1
            fi
        fi
        print_info "使用自定义配置文件: ${CONFIG_PATH}"
    fi

    case "$MODE" in
        master)
            print_info "启动 Master 模式..."
            ;;
        client)
            print_info "启动 Client 模式..."

            if [ -n "$MASTER_ADDR" ]; then
                print_info "Master 地址: ${MASTER_ADDR}"
            else
                print_info "将从配置文件读取 Master 地址"
            fi

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

    # 构建命令参数
    local ARGS=()
    ARGS+=("${MODE}")

    # 添加配置文件参数
    if [ -n "${CONFIG_PATH}" ]; then
        ARGS+=("--config" "${CONFIG_PATH}")
    fi

    # 显示启动信息
    echo ""
    echo "=========================================="
    echo "  🐏 Shepherd"
    echo "=========================================="
    echo "  模式: ${MODE}"
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
