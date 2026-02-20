#!/bin/bash
# Shepherd Linux/macOS 编译脚本
# 用法: ./scripts/build.sh [version]

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 项目信息
PROJECT_NAME="shepherd"
BUILD_DIR="build"
CMD_DIR="cmd/shepherd"
VERSION=${1:-"dev"}

# 构建信息
BUILD_TIME=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
GIT_COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")
GO_VERSION=$(go version | awk '{print $3}')

echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}  Shepherd Build Script (Linux/macOS)${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""
echo "版本: ${VERSION}"
echo "Go 版本: ${GO_VERSION}"
echo "Git Commit: ${GIT_COMMIT}"
echo ""

# 清理旧的构建文件
echo -e "${YELLOW}清理旧的构建文件...${NC}"
rm -rf "${BUILD_DIR}"
mkdir -p "${BUILD_DIR}"

# 设置编译参数
LDFLAGS="-X main.Version=${VERSION}"
LDFLAGS="${LDFLAGS} -X main.BuildTime=${BUILD_TIME}"
LDFLAGS="${LDFLAGS} -X main.GitCommit=${GIT_COMMIT}"
LDFLAGS="${LDFLAGS} -s -w"  # 去除调试信息，减小二进制大小

# 检测操作系统
OS=$(uname -s)
case "${OS}" in
    Linux*)
        TARGET_OS="linux"
        ;;
    Darwin*)
        TARGET_OS="darwin"
        ;;
    *)
        echo -e "${RED}未知操作系统: ${OS}${NC}"
        exit 1
        ;;
esac

# 检测架构
ARCH=$(uname -m)
case "${ARCH}" in
    x86_64|amd64)
        TARGET_ARCH="amd64"
        ;;
    aarch64|arm64)
        TARGET_ARCH="arm64"
        ;;
    *)
        echo -e "${RED}未知架构: ${ARCH}${NC}"
        exit 1
        ;;
esac

BINARY_NAME="${PROJECT_NAME}-${TARGET_OS}-${TARGET_ARCH}"
if [ "${TARGET_OS}" = "linux" ] && [ "${TARGET_ARCH}" = "amd64" ]; then
    BINARY_NAME="${PROJECT_NAME}"
fi

echo -e "${YELLOW}编译目标: ${TARGET_OS}/${TARGET_ARCH}${NC}"
echo -e "${YELLOW}输出文件: ${BUILD_DIR}/${BINARY_NAME}${NC}"
echo ""

# 编译
echo -e "${GREEN}开始编译...${NC}"
if [ "${GOPROXY}" = "" ]; then
    GOPROXY="https://goproxy.cn,direct"
fi
export GOPROXY

# 检查 Go 是否安装
if ! command -v go &> /dev/null; then
    echo -e "${RED}错误: Go 未安装或不在 PATH 中${NC}"
    exit 1
fi

go build \
    -ldflags "${LDFLAGS}" \
    -o "${BUILD_DIR}/${BINARY_NAME}" \
    ./${CMD_DIR}

if [ $? -eq 0 ]; then
    echo -e "${GREEN}✓ 编译成功!${NC}"
else
    echo -e "${RED}✗ 编译失败${NC}"
    exit 1
fi

# 获取二进制大小
BINARY_SIZE=$(du -h "${BUILD_DIR}/${BINARY_NAME}" | cut -f1)
echo ""
echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}  构建完成${NC}"
echo -e "${GREEN}========================================${NC}"
echo "二进制: ${BUILD_DIR}/${BINARY_NAME}"
echo "大小: ${BINARY_SIZE}"
echo ""

# 创建符号链接（如果需要）
if [ "${BINARY_NAME}" != "${PROJECT_NAME}" ]; then
    ln -sf "${BINARY_NAME}" "${BUILD_DIR}/${PROJECT_NAME}"
    echo "链接: ${BUILD_DIR}/${PROJECT_NAME} -> ${BINARY_NAME}"
    echo ""
fi

# 运行测试（可选）
if [ "${RUN_TESTS}" = "true" ]; then
    echo -e "${YELLOW}运行测试...${NC}"
    go test ./... -v
fi

# 使用提示
echo -e "${YELLOW}使用方法:${NC}"
echo "  ./${BUILD_DIR}/${BINARY_NAME}                    # 单机模式"
echo "  ./${BUILD_DIR}/${BINARY_NAME} --mode=master      # Master 模式"
echo "  ./${BUILD_DIR}/${BINARY_NAME} --mode=client      # Client 模式"
echo ""
