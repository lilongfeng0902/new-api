# NewAPI限流限额系统架构分析方案

## 概述

NewAPI网关实现了一个复杂的三层控制系统用于AI请求管理：

1. **限流层**：5种基于IP/用户的请求节流类型
2. **配额管理层**：带信任配额优化的预消费机制、多比率定价
3. **通道路由层**：优先级加权随机选择与智能故障转移

本分析为理解和扩展这些系统提供了完整的技术蓝图。

---

## 系统架构总览

```
用户请求
    ↓
路由层 (Gin)
    ↓
中间件链
  ├─ TokenAuth (身份认证)
  ├─ ModelRequestRateLimit (基于用户的限流)
  └─ Distribute (通道选择)
    ↓
控制器层 (Relay)
  ├─ PreConsumeQuota (配额验证与扣减)
  └─ 重试循环与通道选择
    ↓
服务层
  ├─ 通道选择 (加权随机 + 优先级)
  ├─ Token计数与定价
  └─ 格式转换
    ↓
转发层 (30+ AI提供商)
    ↓
后结算
  └─ 实际用量计算与配额调整
    ↓
数据库 & Redis (审计日志、缓存)
```

---

## 1. 限流系统

### 1.1 全局限流器（基于IP）

**相关文件：**
- [middleware/rate-limit.go](e:\sunzone\MyDocument\github\new-api\middleware\rate-limit.go)
- [common/rate-limit.go](e:\sunzone\MyDocument\github\new-api\common\rate-limit.go)
- [common/limiter/limiter.go](e:\sunzone\MyDocument\github\new-api\common\limiter\limiter.go)

**五种类型：**
1. **GlobalWebRateLimit** - Web界面限流（默认：60次/180秒）
2. **GlobalAPIRateLimit** - API端点限流（默认：180次/180秒）
3. **CriticalRateLimit** - 登录/注册限流（默认：20次/20分钟）
4. **DownloadRateLimit** - 文件下载限流（默认：10次/60秒）
5. **UploadRateLimit** - 文件上传限流（默认：10次/60秒）

**实现方式：**
- **Redis模式**：使用Redis List实现分布式限流
- **内存模式**：使用内存滑动窗口，适用于单实例部署
- **算法**：基于时间戳队列的滑动窗口

**Key命名模式：**
```
Key: "rateLimit:{mark}{clientIP}"
1. 检查列表长度 < maxRequestNum → 允许 + LPush时间戳
2. 如果长度 >= max → 检查最旧的时间戳
3. 如果 (now - oldest) < duration → 拒绝 (429)
4. 否则 → LPush + LTrim → 允许
```

### 1.2 模型级限流器（基于用户）

**相关文件：**
- [middleware/model-rate-limit.go](e:\sunzone\MyDocument\github\new-api\middleware\model-rate-limit.go)
- [setting/rate_limit.go](e:\sunzone\MyDocument\github\new-api\setting\rate_limit.go)

**功能特性：**
- 按用户限流（非基于IP）
- 双计数器：**总请求数** + **成功请求数**（状态码 < 400）
- 通过 `ModelRequestRateLimitGroup` map 实现分组特定覆盖
- 总请求数使用令牌桶算法（使用Lua脚本）

**配置示例：**
```go
ModelRequestRateLimitEnabled = true/false
ModelRequestRateLimitDurationMinutes = 1
ModelRequestRateLimitCount = 0           // 总请求数（0 = 禁用）
ModelRequestRateLimitSuccessCount = 1000 // 成功请求数
ModelRequestRateLimitGroup = {"premium": [5000, 3000], "standard": [1000, 800]}
```

**可扩展性：** 基于分组的配置允许实现多租户限流

---

## 2. 配额管理系统

### 2.1 预消费机制

**相关文件：**
- [service/pre_consume_quota.go](e:\sunzone\MyDocument\github\new-api\service\pre_consume_quota.go)
- [common/quota.go](e:\sunzone\MyDocument\github\new-api\common\quota.go)

**目的：** 防止无限LLM生成导致成本失控

**信任配额优化：**
```
trustQuota = 10 * QuotaPerUnit  // 默认：500万配额

如果 userQuota > trustQuota:
  如果 token.unlimited 或 tokenQuota > trustQuota:
    preConsumedQuota = 0  // 跳过预消费
    日志: "用户受信任，无需预消费"
```

**执行流程：**
1. 验证用户配额 > 0
2. 检查是否需要预消费（信任配额逻辑）
3. 先从Token配额扣减
4. 再从用户配额扣减
5. 存储到 `relayInfo.FinalPreConsumedQuota`

**优势：** 高余额用户跳过预消费以提升用户体验，同时仍在请求后跟踪实际使用量。

### 2.2 定价计算

**相关文件：**
- [relay/helper/price.go](e:\sunzone\MyDocument\github\new-api\relay\helper\price.go:48-140)
- [setting/ratio_setting/model_ratio.go](e:\sunzone\MyDocument\github\new-api\setting\ratio_setting\model_ratio.go)
- [model/pricing.go](e:\sunzone\MyDocument\github\new-api\model\pricing.go)

**多比率系统：**

**基于Token的定价（默认）：**
```
preConsumedTokens = max(promptTokens, PreConsumedQuota) + maxTokens

应用的比率：
- modelRatio: 模型基础成本倍率
- completionRatio: 输出token倍率（如Claude为3倍）
- groupRatio: 租户特定定价
- cacheRatio: 提示词缓存读取折扣（如0.1倍）
- cacheCreationRatio: 缓存写入成本（如1.25倍）
- cacheCreation1hRatio: 1小时缓存层级（Claude为6/3.75倍率）
- audioRatio: 音频输入token成本
- audioCompletionRatio: 音频输出token成本
- imageRatio: 图片token等价定价

quota = preConsumedTokens × modelRatio × groupRatio
```

**基于价格的定价：**
```
quota = modelPrice × QuotaPerUnit × groupRatio
```

**免费模型处理：**
- 如果 `modelRatio == 0` 或 `modelPrice == 0` → `freeModel = true`，跳过预消费
- 配置项：`setting.EnableFreeModelPreConsume`（默认：true）

### 2.3 后结算

**相关文件：**
- [service/quota.go](e:\sunzone\MyDocument\github\new-api\service\quota.go:238-533)
- [relay/compatible_handler.go](e:\sunzone\MyDocument\github\new-api\relay\compatible_handler.go:196-484)

**执行流程：**
1. 从响应中提取实际token（提示词、补全、缓存、音频、图片）
2. 应用所有比率计算实际配额
3. 与预消费对比：`quotaDelta = actualQuota - preConsumedQuota`
4. 如果 quotaDelta > 0：扣除额外配额
5. 如果 quotaDelta < 0：退还多余配额
6. 更新用户/Token已用配额计数器
7. 记录详细消费日志

**高级特性：**
- 使用 `shopspring/decimal` 进行精确配额计算
- 支持Claude特定的提示词缓存处理（1小时/5分钟缓存层级）
- 音频输入/输出token区分
- 混合文本+音频token计算

---

## 3. 通道路由系统

### 3.1 通道选择算法

**相关文件：**
- [service/channel_select.go](e:\sunzone\MyDocument\github\new-api\service\channel_select.go:83-162)
- [model/channel_cache.go](e:\sunzone\MyDocument\github\new-api\model\channel_cache.go:96-191)

**优先级加权随机选择：**

**数据结构：**
```
group2model2channels: map[string]map[string][]int
  分组 → 模型 → [通道ID1, 通道ID2, ...]
  （按优先级降序排列）

channelsIDM: map[int]*Channel
  通道ID → Channel对象
```

**算法流程：**
```
1. 获取 (group, model) 对应的通道列表
2. 提取唯一优先级：[100, 50, 10, 0]
3. 根据重试次数选择优先级层级
   targetPriority = priorities[retry]
4. 按 targetPriority 过滤通道
5. 加权随机选择：
   sumWeight = Σ(channel.weight)

   如果 sumWeight == 0:
     // 所有权重为0 → 等概率
     sumWeight = len(targetChannels) × 100
     smoothingAdjustment = 100
   否则如果 sumWeight/len < 10:
     smoothingFactor = 100  // 放大小权重

   randomWeight = rand(0, sumWeight × smoothingFactor)

   对于每个通道:
     randomWeight -= (channel.weight × smoothingFactor + smoothingAdjustment)
     如果 randomWeight < 0 → 返回该通道
```

**多密钥支持：**
- 模式1：随机 - 随机选择已启用的密钥
- 模式2：轮询 - 带锁保护的递增轮询
- 失败时自动禁用单个密钥

### 3.2 自动分组跨组重试

**相关文件：** [service/channel_select.go](e:\sunzone\MyDocument\github\new-api\service\channel_select.go:89-154)

**概念：** 当 `tokenGroup == "auto"` 时，智能跨多个分组搜索，每个分组内优先级逐级耗尽。

**示例流程：**
```
autoGroups = ["premium", "standard", "free"]

重试0: Group="premium", priority=100 → channel_123 → 失败 (503)
重试1: Group="premium", priority=50  → channel_456 → 失败 (429)
重试2: Group="premium", priority=10  → 无可用通道
       → 切换到 Group="standard", priority=100 → channel_789 → 成功 ✓
```

**状态管理：**
- `ContextKeyAutoGroupIndex`：当前分组索引
- `ContextKeyAutoGroupRetryIndex`：当前分组开始时的重试计数
- 允许无缝故障转移：Premium（P0→P1→P2）→ Standard（P0→P1）→ Free（P0）

### 3.3 健康监控与自动禁用

**相关文件：**
- [controller/relay.go](e:\sunzone\MyDocument\github\new-api\controller\relay.go:345-383)
- [controller/channel-test.go](e:\sunzone\MyDocument\github\new-api\controller\channel-test.go:642-666)
- [service/channel.go](e:\sunzone\MyDocument\github\new-api\service\channel.go)

**自动测试：**
- 函数：`AutomaticallyTestChannels()`
- 频率：`AutoTestChannelMinutes`（可配置）
- 执行：仅在主节点执行（`common.IsMasterNode` 检查）
- 测量响应时间并自动禁用慢速通道

**自动禁用条件：**
- HTTP状态码：401、403（认证失败）
- 特定提供商错误消息（invalid_api_key、insufficient_quota）
- 响应超时超过 `ChannelDisableThreshold`
- 多密钥模式：禁用特定密钥，而非整个通道

**自动启用：**
- `ShouldEnableChannel()` 在成功请求时重新启用
- 跟踪禁用原因和时间戳

### 3.4 故障转移机制

**相关文件：** [controller/relay.go](e:\sunzone\MyDocument\github\new-api\controller\relay.go:174-343)

**重试循环：**
```go
for retry <= RetryTimes:
    channel = getChannel(c, relayInfo, retryParam)
    newAPIError = relayHandler(c, relayInfo)

    if newAPIError == nil:
        return success

    processChannelError()  // 需要时自动禁用

    if !shouldRetry(c, newAPIError, retryTimes - retry):
        break

    retryParam.IncreaseRetry()
```

**重试决策（shouldRetry）：**
- **重试情况：** 429（限流）、5xx（除504/524外）、通道错误、307重定向
- **不重试情况：** 2xx（成功）、400（错误请求）、408/504（超时）、4xx（客户端错误）

---

## 4. 数据流程：完整请求生命周期

```
1. HTTP请求 (POST /v1/chat/completions)
   ↓
2. 中间件链
   ├─ TokenAuth → 验证API密钥，设置userId/tokenId/groups
   ├─ ModelRequestRateLimit
   │   ├─ 检查成功请求计数（滑动窗口）
   │   ├─ 检查总请求计数（令牌桶）
   │   └─ 如果超限 → 429
   └─ Distribute
       ├─ 从请求中提取model
       ├─ 检查token模型限制
       └─ CacheGetRandomSatisfiedChannel()
           └─ SetupContextForSelectedChannel()
   ↓
3. 控制器: Relay()
   ├─ 解析请求体
   ├─ 创建relayInfo对象
   ├─ EstimateRequestToken() → 计算提示词token数
   └─ ModelPriceHelper() → 计算定价
       ├─ 获取模型/分组比率
       ├─ 计算预消费配额
       └─ 检查是否为免费模型
   ↓
4. 服务: PreConsumeQuota()
   ├─ GetUserQuota() → userQuota
   ├─ 信任配额: 如果 userQuota > trustQuota → 跳过预消费
   ├─ PreConsumeTokenQuota()
   ├─ DecreaseUserQuota()
   └─ relayInfo.FinalPreConsumedQuota = 实际扣除量
   ↓
5. 重试循环 (retry = 0 to RetryTimes)
   ├─ getChannel() → 重试时刷新通道
   ├─ relayHandler() → 路由到提供商
   │   └─ 转发到上游（OpenAI、Claude等）
   ├─ 如果出错:
   │   ├─ processChannelError() → 需要时自动禁用通道
   │   └─ shouldRetry() → 是（429、5xx）/ 否（4xx、超时）
   └─ 如果成功 → 跳出循环
   ↓
6. 服务: 后结算
   ├─ 从响应中提取实际token
   ├─ 应用所有比率（模型、补全、缓存、音频、图片）
   ├─ quotaDelta = actual - preConsumed
   ├─ PostConsumeQuota(quotaDelta) → 扣除/退还
   ├─ UpdateUserUsedQuotaAndRequestCount()
   ├─ UpdateChannelUsedQuota()
   └─ RecordConsumeLog() → 审计追踪
   ↓
7. 清理与响应
   ├─ 如果出错且 preConsumed > 0: ReturnPreConsumedQuota()
   ├─ 记录模型限流成功（如果 status < 400）
   └─ 返回响应给客户端
```

---

## 5. 扩展点与建议

### 5.1 添加新的限流类型

**模式：**
```go
// 1. 定义新的限流函数 (middleware/rate-limit.go)
func NewCustomRateLimit() func(c *gin.Context) {
    if common.CustomRateLimitEnable {
        return rateLimitFactory(
            common.CustomRateLimitNum,
            common.CustomRateLimitDuration,
            "CM",  // 唯一的标记前缀
        )
    }
    return defNext
}

// 2. 添加配置 (common/constants.go)
var (
    CustomRateLimitEnable   bool
    CustomRateLimitNum      int
    CustomRateLimitDuration int64
)

// 3. 在路由中注册 (router/relay-router.go)
router.Use(middleware.NewCustomRateLimit())
```

### 5.2 添加新的定价比率

**模式：**
```go
// 1. 添加比率获取器 (setting/ratio_setting/)
func GetCustomRatio(modelName string) float64 {
    // 检查模型特定配置
    // 回退到默认值
    return 1.0
}

// 2. 更新PriceData结构 (types/price_data.go)
type PriceData struct {
    // ... 现有字段
    CustomRatio float64
}

// 3. 在ModelPriceHelper中应用 (relay/helper/price.go)
customRatio := ratio_setting.GetCustomRatio(modelName)
priceData.CustomRatio = customRatio

// 4. 在PostConsumeQuota计算中使用 (service/quota.go)
quota = (tokens × customRatio) × modelRatio × groupRatio
```

### 5.3 添加新的通道选择策略

**当前策略：** 优先级加权随机

**扩展点：** [model/channel_cache.go:96-191](e:\sunzone\MyDocument\github\new-api\model\channel_cache.go)

**建议的策略：**

1. **最近最少使用（LRU）：**
   - 跟踪 `channel.LastUsedTime`
   - 在目标优先级选择最旧的LastUsedTime通道

2. **基于响应时间：**
   - 使用健康检查的 `channel.ResponseTime`
   - Weight = 1000 / ResponseTime（偏好更快的通道）

3. **成本优化：**
   - 跟踪 `channel.UsedQuota / channel.Balance`
   - 选择有可用配额的最便宜通道

4. **地理亲和性：**
   - 添加 `channel.Region` 字段
   - 匹配 `user.Region` 以最小化延迟

**实现示例：**
```go
func GetRandomSatisfiedChannel(group, model string, retry int, strategy string) (*Channel, error) {
    switch strategy {
    case "weighted-random":
        return weightedRandomSelection(targetChannels)
    case "lru":
        return lruSelection(targetChannels)
    case "response-time":
        return responseTimeSelection(targetChannels)
    case "cost-optimized":
        return costOptimizedSelection(targetChannels)
    default:
        return weightedRandomSelection(targetChannels)
    }
}
```

### 5.4 动态自动分组排序

**当前实现：** 从 `setting.GetAutoGroups()` 获取固定分组顺序

**增强方案：**
```go
func GetUserAutoGroupDynamic(userGroup, modelName string) []string {
    allGroups := setting.GetAutoGroups()

    // 基于以下因素对每个分组评分：
    // 1. 该模型的成功率（从日志中）
    // 2. 平均响应时间（从健康检查）
    // 3. 用户的首选分组（从设置）

    scores := make(map[string]float64)
    for _, group := range allGroups {
        successRate := getGroupModelSuccessRate(group, modelName)
        avgResponseTime := getGroupModelAvgResponseTime(group, modelName)

        score := (successRate * 0.7) + ((1000.0 / avgResponseTime) * 0.3)
        scores[group] = score
    }

    return sortGroupsByScore(allGroups, scores)
}
```

### 5.5 配额池（团队共享）

**增强方案：**
```go
type QuotaPool struct {
    ID          int
    Name        string
    TotalQuota  int
    UsedQuota   int
    MemberIDs   []int
}

func PreConsumeQuotaWithPool(quota int, relayInfo *RelayInfo) error {
    // 先尝试用户配额
    userQuota, _ := model.GetUserQuota(relayInfo.UserId)

    if userQuota >= quota {
        return PreConsumeQuota(quota, relayInfo)
    }

    // 回退到池配额
    pool, err := model.GetUserQuotaPool(relayInfo.UserId)
    if err == nil && pool.TotalQuota - pool.UsedQuota >= quota {
        return pool.DecreaseQuota(quota)
    }

    return errors.New("用户和池中配额不足")
}
```

---

## 6. 性能优化机会

### 6.1 通道缓存同步

**当前实现：** 每 `syncFrequency` 秒完全重新加载（[model/channel_cache.go:88-94](e:\sunzone\MyDocument\github\new-api\model\channel_cache.go)）

**优化方案：** 使用数据库触发器或CDC（变更数据捕获）实现增量更新

### 6.2 限流中的Redis往返

**当前实现：** 每个请求4-5次Redis调用（LLen、LIndex、LPush、LTrim、Expire）

**优化方案：** 使用Lua脚本原子执行所有操作
```lua
local rateLimitScript = `
local key = KEYS[1]
local max = tonumber(ARGV[1])
local duration = tonumber(ARGV[2])
local now = ARGV[3]

local len = redis.call('LLEN', key)
if len < max then
    redis.call('LPUSH', key, now)
    redis.call('EXPIRE', key, duration)
    return 1
end

local oldest = redis.call('LINDEX', key, -1)
if tonumber(now) - tonumber(oldest) < duration then
    return 0
end

redis.call('LPUSH', key, now)
redis.call('LTRIM', key, 0, max - 1)
redis.call('EXPIRE', key, duration)
return 1
`
```

### 6.3 GetUserQuota缓存

**当前实现：** 每个请求都查询数据库（[service/pre_consume_quota.go:34](e:\sunzone\MyDocument\github\new-api\service\pre_consume_quota.go)）

**优化方案：** 在Redis中缓存，带TTL，配额变更时失效

---

## 7. 关键文件索引

### 限流相关：
- [middleware/rate-limit.go](e:\sunzone\MyDocument\github\new-api\middleware\rate-limit.go) - 全局IP限流
- [middleware/model-rate-limit.go](e:\sunzone\MyDocument\github\new-api\middleware\model-rate-limit.go) - 模型级用户限流
- [common/limiter/lua/rate_limit.lua](e:\sunzone\MyDocument\github\new-api\common\limiter\lua\rate_limit.lua) - 令牌桶Lua脚本
- [setting/rate_limit.go](e:\sunzone\MyDocument\github\new-api\setting\rate_limit.go) - 限流配置

### 配额管理：
- [service/pre_consume_quota.go](e:\sunzone\MyDocument\github\new-api\service\pre_consume_quota.go) - 预消费逻辑
- [service/quota.go](e:\sunzone\MyDocument\github\new-api\service\quota.go) - 后结算
- [relay/helper/price.go](e:\sunzone\MyDocument\github\new-api\relay\helper\price.go) - 定价计算
- [relay/compatible_handler.go](e:\sunzone\MyDocument\github\new-api\relay\compatible_handler.go) - 高级配额计算
- [model/user.go](e:\sunzone\MyDocument\github\new-api\model\user.go) - 用户配额变更
- [model/token.go](e:\sunzone\MyDocument\github\new-api\model\token.go) - Token配额变更
- [model/log.go](e:\sunzone\MyDocument\github\new-api\model\log.go) - 消费日志

### 通道路由：
- [service/channel_select.go](e:\sunzone\MyDocument\github\new-api\service\channel_select.go) - 自动分组跨组重试
- [model/channel_cache.go](e:\sunzone\MyDocument\github\new-api\model\channel_cache.go) - 加权随机选择
- [model/channel.go](e:\sunzone\MyDocument\github\new-api\model\channel.go) - 通道模型、多密钥支持
- [controller/channel-test.go](e:\sunzone\MyDocument\github\new-api\controller\channel-test.go) - 健康监控

### 集成与编排：
- [controller/relay.go](e:\sunzone\MyDocument\github\new-api\controller\relay.go) - 主转发控制器、重试循环
- [middleware/distributor.go](e:\sunzone\MyDocument\github\new-api\middleware\distributor.go) - 通道选择中间件
- [router/relay-router.go](e:\sunzone\MyDocument\github\new-api\router\relay-router.go) - 中间件链配置

---

## 8. 核心架构洞察

1. **基于优先级的加权随机选择** 防止单通道雪崩
2. **跨组重试** 在不重启重试链的情况下回退到低优先级分组
3. **双重健康监控**：实时响应时间 + 定期自动测试
4. **优雅的通道降级**：明确失败时自动禁用，成功时自动恢复
5. **配额预消费** 在信任充足用户/Token的同时防止过度使用
6. **多层限流**：全局（IP）、每用户、每模型
7. **智能故障转移** 区分可重试错误（429、5xx）和永久失败（401、400）
8. **多比率定价** 支持复杂计费：模型、分组、补全、缓存、音频、图片比率
9. **免费模型处理** 允许零成本模型同时保持审计日志
10. **多密钥支持** 实现通道级负载均衡和密钥轮换

---

## 9. 未来扩展指南

扩展这些系统时，请遵循以下原则：

1. **保持关注点分离**：限流 → 配额 → 路由是独立的层
2. **保持向后兼容**：将新功能添加为可选配置
3. **使用现有模式**：遵循限流器的工厂模式、定价的比率获取器模式
4. **积极缓存**：Redis缓存对大规模性能至关重要
5. **全面记录**：每个配额交易都应可审计
6. **优雅处理错误**：错误时退还配额，失败时自动禁用通道
7. **测试边缘情况**：零权重、零配额、通道耗尽场景
8. **考虑多租户**：为不同租户层级提供基于分组的配置
9. **持续监控**：跟踪限流拒绝、配额消费模式、通道健康
10. **增量优化**：优化前先分析，变更后测量

---

本分析为理解和扩展NewAPI网关的限流、配额管理和通道路由系统提供了完整的架构蓝图。
