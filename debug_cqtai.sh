#!/bin/bash
# Cqtai调试脚本

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 配置
NEW_API_URL="${NEW_API_URL:-http://localhost:3000}"
NEW_API_TOKEN="${NEW_API_TOKEN:-}"

if [ -z "$NEW_API_TOKEN" ]; then
    echo -e "${RED}错误: 请设置环境变量 NEW_API_TOKEN${NC}"
    echo "export NEW_API_TOKEN='sk-xxxxxx'"
    exit 1
fi

echo -e "${GREEN}=== Cqtai API 调试工具 ===${NC}\n"

# 1. 测试提交任务
echo -e "${YELLOW}[1] 测试提交音乐生成任务${NC}"
RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "${NEW_API_URL}/api/cqt/generator/suno" \
  -H "Authorization: Bearer ${NEW_API_TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{
    "prompt": "A cheerful pop song about debugging",
    "tags": "pop, happy",
    "mv": "chirp-v3-0"
  }')

HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
BODY=$(echo "$RESPONSE" | sed '$d')

echo "HTTP状态码: $HTTP_CODE"
echo "响应内容: $BODY" | jq '.' 2>/dev/null || echo "$BODY"

if [ "$HTTP_CODE" = "200" ]; then
    echo -e "${GREEN}✓ 提交成功${NC}\n"
    TASK_ID=$(echo "$BODY" | jq -r '.data' 2>/dev/null)
    if [ -n "$TASK_ID" ] && [ "$TASK_ID" != "null" ]; then
        echo "Task ID: $TASK_ID"

        # 2. 测试查询任务
        echo -e "\n${YELLOW}[2] 测试查询任务状态${NC}"
        sleep 2

        QUERY_RESPONSE=$(curl -s -w "\n%{http_code}" -X GET "${NEW_API_URL}/api/cqt/v2/sunoinfo?id=${TASK_ID}" \
          -H "Authorization: Bearer ${NEW_API_TOKEN}")

        QUERY_HTTP_CODE=$(echo "$QUERY_RESPONSE" | tail -n1)
        QUERY_BODY=$(echo "$QUERY_RESPONSE" | sed '$d')

        echo "HTTP状态码: $QUERY_HTTP_CODE"
        echo "响应内容: $QUERY_BODY" | jq '.' 2>/dev/null || echo "$QUERY_BODY"

        if [ "$QUERY_HTTP_CODE" = "200" ]; then
            echo -e "${GREEN}✓ 查询成功${NC}\n"
        else
            echo -e "${RED}✗ 查询失败${NC}\n"
        fi
    fi
else
    echo -e "${RED}✗ 提交失败${NC}\n"
fi

# 3. 测试渠道健康检查
echo -e "${YELLOW}[3] 检查Cqtai渠道配置${NC}"
curl -s "${NEW_API_URL}/api/channel" \
  -H "Authorization: Bearer ${NEW_API_TOKEN}" | \
  jq '.data[] | select(.type == 58) | {id, name, status, base_url}' 2>/dev/null || \
  echo "无法获取渠道信息（需要管理员权限）"

echo -e "\n${GREEN}=== 调试完成 ===${NC}"
