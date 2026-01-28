# Cqtai 渠道配置指南

本指南将帮助你快速添加和配置 Cqtai 渠道。

## 📋 前置要求

- 已部署 new-api 服务
- 拥有 Cqtai API Key
- 数据库文件位于项目根目录 (`one-api.db`)

## 🎯 快速开始

### 方式1：使用自动化脚本（推荐）

#### 步骤1：修改API Key

编辑 `scripts/add_cqtai_channel.go` 文件的第14行：

```bash
# 使用你喜欢的编辑器打开
nano scripts/add_cqtai_channel.go
# 或
vim scripts/add_cqtai_channel.go
```

将以下内容：
```go
apiKey := "your-cqtai-api-key-here"
```

修改为：
```go
apiKey := "你的实际Cqtai-API-Key"
```

#### 步骤2：执行添加脚本

```bash
# 方式A：使用便捷脚本
./scripts/add_cqtai_channel.sh

# 方式B：直接运行Go程序
go run scripts/add_cqtai_channel.go
```

### 方式2：使用SQL脚本

如果你更熟悉SQL操作：

#### 步骤1：编辑SQL文件

```bash
nano add_cqtai_channel.sql
```

修改第31行的API Key：
```sql
'your-cqtai-api-key-here',  -- 改为你的实际API Key
```

#### 步骤2：安装sqlite3（如未安装）

```bash
sudo apt-get update && sudo apt-get install sqlite3
```

#### 步骤3：执行SQL

```bash
sqlite3 one-api.db < add_cqtai_channel.sql
```

## ✅ 验证安装

执行以下命令验证渠道是否添加成功：

```bash
# 查看Cqtai渠道
sqlite3 one-api.db "SELECT id, name, type, base_url, models, status FROM channels WHERE type=58;"

# 查看模型能力
sqlite3 one-api.db "SELECT a.id, a.model, a.enabled, c.name FROM abilities a JOIN channels c ON a.channel_id = c.id WHERE c.type=58;"
```

预期输出：
```
58|Cqtai|58|https://api.cqtai.com|suno_music,suno_lyrics|1

1|suno_music|1|Cqtai
2|suno_lyrics|1|Cqtai
```

## 🔄 重启服务

添加渠道后，需要重启 new-api 服务：

```bash
# 如果使用systemd
sudo systemctl restart new-api

# 如果直接运行
# 停止当前进程（Ctrl+C），然后重新启动
./new-api

# 如果使用Docker
docker restart new-api
```

## 🧪 测试渠道

使用调试脚本测试渠道是否正常工作：

```bash
# 设置环境变量
export NEW_API_URL="http://localhost:3001"
export NEW_API_TOKEN="sk-your-token"

# 运行测试脚本
./debug_cqtai.sh
```

或手动测试：

```bash
curl -X POST http://localhost:3001/api/cqt/generator/suno \
  -H "Authorization: Bearer sk-your-token" \
  -H "Content-Type: application/json" \
  -d '{
   "task": "lyrics",
   "prompt": "A happy song about a girl and a boy."
   }'
```

## 📊 管理后台配置

1. 登录管理后台：`http://your-domain:3001`
2. 进入 **渠道管理** 页面
3. 找到 **Cqtai** 渠道，检查：
   - ✅ 状态为"启用"
   - ✅ Base URL 为 `https://api.cqtai.com`
   - ✅ 模型列表包含 `suno_music,suno_lyrics`
   - ✅ API Key 已正确配置

4. 进入 **令牌管理** 页面，确保：
   - ✅ 你的Token启用了 `suno_music` 和 `suno_lyrics` 模型

## 🔧 Cqtai渠道配置说明

### 支持的模型

| 模型名称 | 功能 | 说明 |
|---------|------|------|
| `suno_music` | 音乐生成 | 生成完整的音乐作品 |
| `suno_lyrics` | 歌词生成 | 生成歌词内容 |

### API端点

Cqtai渠道使用直接代理模式，支持以下端点：

| 端点 | 方法 | 说明 |
|------|------|------|
| `/api/cqt/generator/suno` | POST | 提交音乐生成任务 |
| `/api/cqt/v2/sunoinfo` | GET | 查询任务状态 |

### 请求示例

#### 提交音乐生成任务

```bash
POST /api/cqt/generator/suno
Content-Type: application/json
Authorization: Bearer {your-token}

{
  "task": "lyrics",
  "prompt": "A happy song about a girl and a boy."
}
```

#### 查询任务状态

```bash
GET /api/cqt/v2/sunoinfo?id={task_id}
Authorization: Bearer {your-token}
```

## 🐛 故障排除

### 问题1：渠道未显示

**原因**：服务未重启或数据未刷新

**解决**：
```bash
# 重启服务
sudo systemctl restart new-api

# 清理缓存（如果启用了缓存）
redis-cli FLUSHDB
```

### 问题2：Token无法使用Cqtai模型

**原因**：Token未启用对应模型

**解决**：
1. 进入管理后台 → 令牌管理
2. 编辑对应Token
3. 在模型列表中勾选 `suno_music` 和 `suno_lyrics`
4. 保存

### 问题3：401 Unauthorized

**原因**：API Key配置错误

**解决**：
1. 确认Cqtai API Key是否正确
2. 在管理后台 → 渠道管理 → 编辑Cqtai渠道
3. 更新正确的API Key
4. 测试渠道

### 问题4：404 Not Found

**原因**：路由未正确注册

**解决**：
1. 检查代码是否已正确合并
2. 确认 `router/relay-router.go` 包含Cqtai路由
3. 检查 `middleware/distributor.go` 是否正确处理路径
4. 重新编译并重启服务

## 📚 相关文档

- [调试指南](./DEBUG_CQTAI.md) - 详细的调试步骤和故障排查
- [调试脚本](./debug_cqtai.sh) - 自动化测试脚本

## 🆘 获取帮助

如果遇到问题：

1. 查看应用日志：
   ```bash
   tail -f logs/chat.log
   ```

2. 查看调试文档：
   ```bash
   cat DEBUG_CQTAI.md
   ```

3. 检查数据库记录：
   ```bash
   sqlite3 one-api.db "SELECT * FROM channels WHERE type=58;"
   ```

4. 运行完整测试：
   ```bash
   ./debug_cqtai.sh
   ```

---

**提示**：完成配置后，建议先通过调试脚本验证渠道功能正常，再投入生产使用。
