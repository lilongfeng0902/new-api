# Cqtai渠道调试指南

## 🚀 快速开始

### 1. 设置环境变量

```bash
export NEW_API_URL="http://localhost:3001"
export NEW_API_TOKEN="sk-xxxxxx"  # 在管理后台创建的token
```

### 2. 运行调试脚本

```bash
./debug_cqtai.sh
```

---

## 🔍 详细调试步骤

### 步骤1: 检查渠道配置

```bash
# 查看数据库中的Cqtai渠道
sqlite3 one-api.db "SELECT id, name, type, base_url, status FROM channels WHERE type=58;"

# 或者使用管理后台 → 渠道管理 → 查看Cqtai渠道
```

**检查项**：
- ✅ type = 58 (ChannelTypeCqtai)
- ✅ base_url = https://api.cqtai.com
- ✅ status = 1 (启用)
- ✅ API Key配置正确

---

### 步骤2: 测试提交任务

#### 方法A: 使用curl

```bash
curl -X POST http://localhost:3001/api/cqt/generator/suno \
  -H "Authorization: Bearer $NEW_API_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
      "task": "lyrics",
      "prompt": "A happy song about a girl and a boy."
    }' \
  -v 2>&1 | tee submit_test.log
```

#### 方法B: 使用Postman/Apifox

```
POST http://localhost:3001/api/cqt/generator/suno
Headers:
  Authorization: Bearer sk-xxxxxx
  Content-Type: application/json
Body (raw JSON):
{
  "task": "lyrics",
  "prompt": "A happy song about a girl and a boy."
}
```

**期望结果**：
```json
{
  "code": "success",
  "message": "ok",
  "data": "task_id_xxx"
}
```

**常见问题**：

| 错误 | 原因 | 解决方案 |
|------|------|----------|
| `401 Unauthorized` | Token无效 | 检查token是否正确 |
| `404 Not Found` | 路由未注册 | 检查router配置 |
| `500 Internal Error` | 代码错误 | 查看日志 |
| `channel not found` | 渠道未配置 | 添加Cqtai渠道 |

---

### 步骤3: 查看日志

#### A. 应用日志

```bash
# 查看最新日志
tail -f logs/chat.log

# 过滤Cqtai相关日志
tail -f logs/chat.log | grep -i cqtai

# 查看错误日志
tail -f logs/chat.log | grep -i error
```

#### B. 添加调试日志

在关键位置添加日志（临时调试）：

```go
// relay/channel/task/cqtai/adaptor.go

func (a *TaskAdaptor) BuildRequestURL(info *relaycommon.RelayInfo) (string, error) {
    baseURL := info.ChannelBaseUrl
    requestPath := info.RequestURLPath
    fullRequestURL := fmt.Sprintf("%s%s", baseURL, requestPath)

    // 添加调试日志
    common.SysLog(fmt.Sprintf("[Cqtai Debug] BuildRequestURL: %s", fullRequestURL))

    return fullRequestURL, nil
}
```

---

### 步骤4: 测试查询任务

```bash
# 查询单个任务
curl -X GET "http://localhost:3001/api/cqt/v2/sunoinfo?id=task_id_xxx" \
  -H "Authorization: Bearer $NEW_API_TOKEN" \
  -v

# 查询多个任务
curl -X GET "http://localhost:3001/api/cqt/v2/sunoinfo?id=task1&id=task2" \
  -H "Authorization: Bearer $NEW_API_TOKEN" \
  -v
```

---

### 步骤5: 检查数据库记录

```bash
# 查看任务表
sqlite3 one-api.db "SELECT id, task_id, platform, action, status, channel_id FROM tasks WHERE platform='cqtai' ORDER BY id DESC LIMIT 10;"

# 查看特定任务
sqlite3 one-api.db "SELECT * FROM tasks WHERE task_id='task_id_xxx';"
```

---

## 🐛 常见问题排查

### 问题1: 请求未到达adaptor

**症状**: 日志中没有看到Cqtai相关请求

**排查**:
```bash
# 1. 检查路由是否注册
grep -n "api/cqt" router/relay-router.go

# 2. 检查middleware是否正确匹配
grep -n "api/cqt" middleware/distributor.go

# 3. 测试路由可达性
curl -X POST http://localhost:3001/api/cqt/generator/suno \
  -H "Authorization: Bearer $NEW_API_TOKEN" \
  -v
```

---

### 问题2: 转发到上游失败

**症状**: 500错误，日志显示连接上游失败

**排查**:
```bash
# 1. 测试网络连通性
curl -X POST https://api.cqtai.com/api/cqt/generator/suno \
  -H "Authorization: your-cqtai-key" \
  -H "Content-Type: application/json" \
  -d '{"prompt": "test"}' \
  -v

# 2. 检查渠道BaseURL
sqlite3 one-api.db "SELECT base_url FROM channels WHERE type=58;"

# 3. 检查代理设置
# 如果使用了代理，确保代理配置正确
```

---

### 问题3: 参数格式错误

**症状**: 上游返回400错误

**排查**:
```bash
# 1. 查看实际发送的请求体
# 在adaptor.go的BuildRequestBody中添加日志：
common.SysLog(fmt.Sprintf("[Cqtai Debug] Request Body: %s", string(data)))

# 2. 对比Cqtai官方文档的参数格式
# 确保参数名称、类型、必填项都正确
```

---

### 问题4: 后台同步不工作

**症状**: 任务一直处于pending状态

**排查**:
```bash
# 1. 检查UpdateCqtaiTaskAll是否被调用
grep -n "UpdateCqtaiTaskAll" controller/task.go

# 2. 检查FetchTask实现
# 确认GET请求和query参数构建正确

# 3. 手动测试同步
# 在日志中查看是否有"渠道 #X 未完成的任务有: N"
```

---

## 🔬 高级调试技巧

### 1. 使用Go调试器

```bash
# 安装delve
go install github.com/go-delve/delve/cmd/dlv@latest

# 调试运行
dlv debug . -- --port 3001

# 设置断点
(dlv) break relay/channel/task/cqtai/adaptor.go:BuildRequestURL
(dlv) continue
```

### 2. 抓包分析

```bash
# 使用mitmproxy抓包
pip install mitmproxy
mitmproxy -p 8080

# 设置代理
export HTTP_PROXY=http://localhost:8080
export HTTPS_PROXY=http://localhost:8080
```

### 3. 单元测试

创建测试文件 `relay/channel/task/cqtai/adaptor_test.go`:

```go
package cqtai

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestBuildRequestURL(t *testing.T) {
	adaptor := &TaskAdaptor{}
	info := &relaycommon.RelayInfo{
		ChannelBaseUrl: "https://api.cqtai.com",
		RequestURLPath: "/api/cqt/generator/suno",
	}

	url, err := adaptor.BuildRequestURL(info)
	assert.NoError(t, err)
	assert.Equal(t, "https://api.cqtai.com/api/cqt/generator/suno", url)
}
```

运行测试:
```bash
go test -v ./relay/channel/task/cqtai/
```

---

## 📊 监控指标

### 关键指标

1. **任务提交成功率**
   ```sql
   SELECT
     COUNT(*) as total,
     SUM(CASE WHEN status='SUCCESS' THEN 1 ELSE 0 END) as success,
     SUM(CASE WHEN status='FAILURE' THEN 1 ELSE 0 END) as failure
   FROM tasks
   WHERE platform='cqtai'
   AND created_at > datetime('now', '-1 hour');
   ```

2. **平均响应时间**
   - 查看日志中的请求耗时

3. **渠道可用性**
   ```sql
   SELECT id, name, status, test_time
   FROM channels
   WHERE type=58;
   ```

---

## 🎯 完整测试案例

```bash
#!/bin/bash
# 完整测试流程

echo "=== 开始Cqtai完整测试 ==="

# 1. 提交任务
echo "[1/4] 提交音乐生成任务..."
SUBMIT_RESP=$(curl -s -X POST http://localhost:3001/api/cqt/generator/suno \
  -H "Authorization: Bearer $NEW_API_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "prompt": "A beautiful piano melody",
    "tags": "piano, classical",
    "mv": "chirp-v3-0"
  }')

echo "$SUBMIT_RESP" | jq '.'
TASK_ID=$(echo "$SUBMIT_RESP" | jq -r '.data')

if [ "$TASK_ID" = "null" ] || [ -z "$TASK_ID" ]; then
    echo "❌ 任务提交失败"
    exit 1
fi

echo "✅ 任务ID: $TASK_ID"

# 2. 立即查询（应该是pending状态）
echo -e "\n[2/4] 查询任务状态（应该是pending）..."
sleep 2
curl -s -X GET "http://localhost:3001/api/cqt/v2/sunoinfo?id=$TASK_ID" \
  -H "Authorization: Bearer $NEW_API_TOKEN" | jq '.'

# 3. 等待15秒后查询（后台同步后）
echo -e "\n[3/4] 等待后台同步..."
sleep 15

echo -e "\n[4/4] 再次查询任务状态..."
curl -s -X GET "http://localhost:3001/api/cqt/v2/sunoinfo?id=$TASK_ID" \
  -H "Authorization: Bearer $NEW_API_TOKEN" | jq '.'

echo -e "\n=== 测试完成 ==="
```

---

## 📝 调试检查清单

- [ ] 渠道已在管理后台配置（type=58, status=1）
- [ ] API Key正确填写
- [ ] Base URL = https://api.cqtai.com
- [ ] Token已创建并包含cqtai_music模型
- [ ] 路由配置正确 (/api/cqt/*)
- [ ] Middleware正确匹配路径
- [ ] Adaptor能正确构建URL
- [ ] 请求体透传无误
- [ ] GET请求的query参数正确构建
- [ ] 后台同步任务正常运行
- [ ] 数据库中能看到任务记录
- [ ] 日志中无error信息

---

## 🆘 获取帮助

如果以上方法都无法解决问题：

1. **查看完整日志**
   ```bash
   tail -n 1000 logs/chat.log > cqtai_debug.log
   ```

2. **导出数据库记录**
   ```bash
   sqlite3 one-api.db <<EOF
   .mode csv
   .output tasks.csv
   SELECT * FROM tasks WHERE platform='cqtai';
   .quit
   EOF
   ```

3. **提供以下信息**
   - Go版本: `go version`
   - 完整错误日志
   - 数据库任务记录
   - curl完整请求和响应
