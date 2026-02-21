#!/bin/bash
# Shepherd 连接诊断脚本

echo "========================================="
echo "Shepherd Master-Client 连接诊断"
echo "========================================="
echo ""

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${YELLOW}[1/5] 检查网络接口...${NC}"
echo "当前 IP 地址:"
ip addr show | grep "inet " | grep -v "127.0.0.1" | awk '{print $2}' | cut -d/ -f1
echo ""

echo -e "${YELLOW}[2/5] 检查 Master 进程...${NC}"
MASTER_PID=$(pgrep -f "shepherd.*master\|shepherd.*standalone" || echo "")
if [ -n "$MASTER_PID" ]; then
    echo -e "${GREEN}✓ Master 进程运行中 (PID: $MASTER_PID)${NC}"
    ps aux | grep -E "PID|$MASTER_PID" | grep -v grep
else
    echo -e "${RED}✗ Master 进程未运行${NC}"
fi
echo ""

echo -e "${YELLOW}[3/5] 检查端口监听...${NC}"
echo "检查 9190 端口:"
PORT_9190=$(sudo netstat -tlnp 2>/dev/null | grep 9190 || sudo ss -tlnp 2>/dev/null | grep 9190 || echo "")
if [ -n "$PORT_9190" ]; then
    echo -e "${GREEN}✓ 端口 9190 正在监听:${NC}"
    echo "$PORT_9190"
else
    echo -e "${RED}✗ 端口 9190 未监听${NC}"
fi
echo ""

echo -e "${YELLOW}[4/5] 测试 Master API (127.0.0.1)...${NC}"
echo "测试: curl -s http://127.0.0.1:9190/api/master/nodes"
RESPONSE=$(curl -s -w "\n%{http_code}" http://127.0.0.1:9190/api/master/nodes 2>&1)
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
BODY=$(echo "$RESPONSE" | sed '$d')

if [ "$HTTP_CODE" = "200" ]; then
    echo -e "${GREEN}✓ Master API 正常响应 (HTTP 200)${NC}"
    echo "响应内容: $BODY"
elif [ "$HTTP_CODE" = "000" ]; then
    echo -e "${RED}✗ 连接失败 (Connection refused)${NC}"
    echo "Master 可能未启动或未绑定到 127.0.0.1"
else
    echo -e "${RED}✗ 异常响应 (HTTP $HTTP_CODE)${NC}"
    echo "响应内容: $BODY"
fi
echo ""

echo -e "${YELLOW}[5/5] 测试 Master API (192.168.10.1)...${NC}"
echo "测试: curl -s http://192.168.10.1:9190/api/master/nodes"
RESPONSE=$(curl -s -w "\n%{http_code}" http://192.168.10.1:9190/api/master/nodes 2>&1)
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
BODY=$(echo "$RESPONSE" | sed '$d')

if [ "$HTTP_CODE" = "200" ]; then
    echo -e "${GREEN}✓ Master API 正常响应 (HTTP 200)${NC}"
    echo "响应内容: $BODY"
    echo ""
    echo -e "${GREEN}=========================================${NC}"
    echo -e "${GREEN}诊断结果: Master 运行正常${NC}"
    echo -e "${GREEN}请检查 Client 配置是否正确指向上述地址${NC}"
    echo -e "${GREEN}=========================================${NC}"
elif [ "$HTTP_CODE" = "000" ]; then
    echo -e "${RED}✗ 连接失败 (Connection refused)${NC}"
    echo ""
    echo -e "${RED}=========================================${NC}"
    echo -e "${RED}诊断结果: Master 未运行或未绑定到 192.168.10.1${NC}"
    echo -e "${RED}=========================================${NC}"
    echo ""
    echo "可能的原因:"
    echo "1. Master 进程未启动"
    echo "2. Master 配置中 server.host 设置为 127.0.0.1 (只监听本地)"
    echo ""
    echo "建议修复:"
    echo "  方法1: 启动 Master: ./build/shepherd master"
    echo "  方法2: 修改 config/node/master.config.yaml:"
    echo "         server:"
    echo "           host: 0.0.0.0  # 确保是 0.0.0.0 而不是 127.0.0.1"
else
    echo -e "${RED}✗ 异常响应 (HTTP $HTTP_CODE)${NC}"
    echo "响应内容: $BODY"
fi
echo ""

echo -e "${YELLOW}[附加] 检查 Client 配置...${NC}"
if [ -f "config/node/client.config.yaml" ]; then
    MASTER_ADDR=$(grep -A2 "client_role:" config/node/client.config.yaml | grep "master_address" | awk '{print $2}')
    echo "Client 配置中的 master_address: $MASTER_ADDR"
else
    echo -e "${RED}config/node/client.config.yaml 不存在${NC}"
fi
echo ""

echo "诊断完成。"