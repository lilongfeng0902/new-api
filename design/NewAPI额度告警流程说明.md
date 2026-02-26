# 额度告警完整流程说明

## 📋 流程概述

额度告警系统从API请求处理开始，经过额度消耗计算、阈值判断、通知内容生成、频率限制检查，最终发送通知给用户。整个流程涉及多个模块的协同工作。

## 🔄 详细流程图

```
API请求处理 → 额度消耗 → 告警触发判断 → 通知内容生成 → 频率限制检查 → 通知发送 → 错误处理
     ↓              ↓              ↓              ↓              ↓              ↓              ↓
  relay层        PostConsumeQuota checkAndSendQuotaNotify 内容格式化    CheckNotificationLimit  NotifyUser    SysError日志
```

## 📝 流程步骤详解

### 1. 触发阶段：API请求与额度消耗

#### 1.1 请求处理入口
```
位置：relay/compatible_handler.go 或 relay/mjproxy_handler.go
函数：API请求处理函数
```

当用户发起API请求时，系统会：
- 解析请求参数
- 验证用户身份和Token
- 计算本次请求的消耗额度
- 调用额度消耗处理

#### 1.2 额度消耗处理
```
位置：service/quota.go
函数：PostConsumeQuota()
```

```go
func PostConsumeQuota(relayInfo *relaycommon.RelayInfo, quota int, preConsumedQuota int, sendEmail bool)
```

**参数说明**：
- `relayInfo`: 请求上下文信息，包含用户信息、用户设置等
- `quota`: 本次实际消耗的额度
- `preConsumedQuota`: 预消耗的额度
- `sendEmail`: 是否启用邮件通知（历史参数名，实际控制是否发送通知）

**处理逻辑**：
1. 更新用户额度：`model.DecreaseUserQuota()` 或 `model.IncreaseUserQuota()`
2. 更新Token额度：`model.DecreaseTokenQuota()` 或 `model.IncreaseTokenQuota()`
3. **关键判断**：
   - `sendEmail` 必须为 `true`
   - `(quota + preConsumedQuota) != 0`（有实际额度变化）
   - 非Playground模式

**触发条件总结**：
```go
if sendEmail && (quota + preConsumedQuota) != 0 && !relayInfo.IsPlayground {
    checkAndSendQuotaNotify(relayInfo, quota, preConsumedQuota)
}
```

### 2. 判断阶段：告警触发条件

#### 2.1 告警检查函数
```
位置：service/quota.go
函数：checkAndSendQuotaNotify()
```

**异步执行**：使用 `gopool.Go()` 异步处理，不阻塞主请求流程

#### 2.2 阈值计算
```go
// 获取告警阈值
threshold := common.QuotaRemindThreshold  // 系统默认阈值
if userSetting.QuotaWarningThreshold != 0 {
    threshold = int(userSetting.QuotaWarningThreshold)  // 用户自定义阈值
}

// 计算剩余额度
consumeQuota := quota + preConsumedQuota
remainingQuota := relayInfo.UserQuota - consumeQuota

// 判断是否触发告警
quotaTooLow := remainingQuota < threshold
```

**阈值来源**：
1. 系统默认：`common.QuotaRemindThreshold`（通常为1000）
2. 用户自定义：`userSetting.QuotaWarningThreshold`

### 3. 内容生成阶段：通知内容格式化

#### 3.1 通知类型判断
```go
notifyType := userSetting.NotifyType
if notifyType == "" {
    notifyType = dto.NotifyTypeEmail  // 默认为邮件
}
```

**支持的通知类型**：
- `email`: 邮件通知
- `webhook`: Webhook通知
- `bark`: Bark推送
- `gotify`: Gotify推送
- `sms`: 阿里云短信（新增）

#### 3.2 内容格式化
根据不同通知类型生成相应格式的内容：

```go
// 邮件/Webhook格式（支持HTML）
content = "{{value}}，当前剩余额度为 {{value}}，为了不影响您的使用，请及时充值。<br/>充值链接：<a href='{{value}}'>{{value}}</a>"
values = []interface{}{prompt, formattedQuota, topUpLink, topUpLink}

// Bark/Gotify/SMS格式（纯文本）
content = "{{value}}，剩余额度：{{value}}，请及时充值"
values = []interface{}{prompt, formattedQuota}
```

**模板参数说明**：
- `{{value}}`：会被 `data.Values` 中的值依次替换
- `prompt`：告警提示文字（"您的额度即将用尽"）
- `formattedQuota`：格式化的额度显示（例如："1,000.50"）
- `topUpLink`：充值页面链接

### 4. 频率控制阶段：通知限制检查

#### 4.1 频率限制检查
```
位置：service/notify-limit.go
函数：CheckNotificationLimit()
```

**限制规则**：
- 每小时最多发送 `constant.NotifyLimitCount` 条同类型通知
- 默认限制：每小时2条
- 基于Redis或内存存储计数

**计数Key格式**：
```
notify_limit:{userId}:{notifyType}:{YYYYMMDDHH}
例如：notify_limit:123:quota_exceed:2024012014
```

**检查逻辑**：
1. 获取当前小时的通知计数
2. 如果超过限制，返回 `false`
3. 如果未超过限制，计数+1，返回 `true`

### 5. 发送阶段：通知分发

#### 5.1 统一通知接口
```
位置：service/user_notify.go
函数：NotifyUser()
```

**参数**：
- `userId`: 用户ID
- `userEmail`: 用户邮箱
- `userSetting`: 用户设置
- `data`: 通知数据（`dto.Notify`结构体）

#### 5.2 通知类型路由
根据用户设置的通知类型，调用对应的发送函数：

```go
switch notifyType {
case dto.NotifyTypeEmail:
    return sendEmailNotify(emailToUse, data)
case dto.NotifyTypeWebhook:
    return SendWebhookNotify(webhookURL, webhookSecret, data)
case dto.NotifyTypeBark:
    return sendBarkNotify(barkURL, data)
case dto.NotifyTypeGotify:
    return sendGotifyNotify(gotifyURL, gotifyToken, priority, data)
case dto.NotifyTypeSMS:
    return sendSMSNotify(phoneNumber, data)
}
```

#### 5.3 短信发送流程
```
位置：service/user_notify.go
函数：sendSMSNotify()
```

**处理步骤**：
1. **占位符替换**：将模板中的 `{{value}}` 替换为实际值
2. **参数提取**：从 `data.Values` 中提取剩余额度信息
3. **调用阿里云SMS**：`common.SendQuotaWarningSMS(phoneNumber, remainingQuota)`

### 6. 阿里云SMS发送流程

#### 6.1 SMS专用函数
```
位置：common/alisms.go
函数：SendQuotaWarningSMS()
```

**处理逻辑**：
1. 构造模板参数：`{"quota": remainingQuota}`
2. 调用通用SMS发送函数：`SendAliyunSMS()`

#### 6.2 阿里云API调用
```
位置：common/alisms.go
函数：SendAliyunSMS()
```

**API调用步骤**：
1. **参数验证**：检查配置完整性
2. **请求构建**：
   - 生成签名和规范化的查询字符串
   - 构建HMAC-SHA1签名
   - 构造完整的API请求URL
3. **HTTP请求**：发送GET请求到阿里云SMS API
4. **响应处理**：解析API响应，判断发送成功与否

### 7. 错误处理与日志

#### 7.1 发送失败处理
```go
err := NotifyUser(relayInfo.UserId, relayInfo.UserEmail, relayInfo.UserSetting, notifyData)
if err != nil {
    common.SysError(fmt.Sprintf("failed to send quota notify to user %d: %s", relayInfo.UserId, err.Error()))
}
```

#### 7.2 日志记录
- **成功发送**：不记录额外日志
- **发送失败**：记录错误日志到系统日志
- **频率限制**：记录跳过发送的原因
- **配置缺失**：记录用户缺少配置信息

## 🔄 完整流程时序图

```
1. API请求到达
   ↓
2. relay层处理 → 计算消耗额度
   ↓
3. service.PostConsumeQuota()
   ↓
4. 检查sendEmail参数 → 如果为true，调用checkAndSendQuotaNotify()
   ↓
5. checkAndSendQuotaNotify() - 异步执行
   ↓
6. 计算剩余额度 < 阈值？ → 是：继续 | 否：结束
   ↓
7. 根据通知类型生成内容格式
   ↓
8. service.NotifyUser() → 检查频率限制
   ↓
9. 频率限制检查通过？ → 是：继续 | 否：跳过发送
   ↓
10. 根据通知类型调用相应发送函数
    ↓
11. sendSMSNotify() → 占位符替换 → 调用阿里云SMS API
    ↓
12. 阿里云SMS API响应 → 成功：完成 | 失败：记录错误日志
```

## ⚙️ 配置参数说明

### 系统配置
- `QuotaRemindThreshold`: 系统默认告警阈值（1000）
- `NotifyLimitCount`: 每小时最大通知次数（2）
- `NotificationLimitDurationMinute`: 通知限制时间窗口（10分钟）

### 用户配置
- `NotifyType`: 通知类型（email/webhook/bark/gotify/sms）
- `QuotaWarningThreshold`: 用户自定义阈值
- `SMSPhoneNumber`: 短信接收手机号（SMS类型时需要）

### 阿里云SMS配置
- `AliyunSMSEndpoint`: API端点
- `AliyunSMSAccessKeyId`: 访问密钥ID
- `AliyunSMSAccessKeySecret`: 访问密钥Secret
- `AliyunSMSSignName`: 短信签名
- `AliyunSMSTemplateCode`: 短信模板CODE

## 🎯 关键设计特点

1. **异步处理**：告警检查不阻塞主请求流程
2. **频率控制**：防止通知轰炸
3. **多渠道支持**：支持多种通知方式
4. **容错设计**：单个通知失败不影响整体流程
5. **用户自定义**：支持个性化阈值和通知方式
6. **模板化内容**：统一的模板替换机制

## 🚨 边界条件与特殊情况

### 触发条件边界
1. **零消耗请求**：`quota + preConsumedQuota == 0` 时不触发告警
2. **Playground模式**：测试环境不发送告警通知
3. **额度增加**：当 `quota < 0` 时表示额度增加，不触发告警
4. **用户无设置**：`userSetting` 为空时使用默认配置

### 频率限制边界
1. **新用户**：首次发送通知时建立计数器
2. **时间窗口**：按小时重置计数器
3. **类型隔离**：不同通知类型分别计数
4. **并发安全**：Redis计数器保证线程安全

### 通知内容边界
1. **额度格式化**：自动添加千分位分隔符（如：1,000.50）
2. **链接生成**：动态生成充值页面链接
3. **字符编码**：确保中文字符正确显示
4. **长度限制**：不同渠道有不同的内容长度限制

### 错误处理边界
1. **网络异常**：阿里云API调用失败时记录错误日志
2. **配置缺失**：缺少必要配置时跳过发送并记录日志
3. **无效数据**：手机号格式错误或其他参数异常时返回错误
4. **频率超限**：超过发送限制时跳过发送（不记录错误）

## 📊 监控与调试

### 关键监控点
1. **告警触发率**：统计告警触发的频率
2. **发送成功率**：各渠道通知的成功率
3. **频率限制命中**：统计被频率限制拦截的通知数
4. **处理耗时**：异步告警处理的平均耗时

### 调试信息
```go
// 系统日志示例
"failed to send quota notify to user 123: 阿里云短信发送失败: isv.INVALID_PARAMETERS"
"user 123 has no sms phone number, skip sending sms"
"notification limit exceeded for user 123 with type sms"
```

### 配置验证
```bash
# 检查阿里云SMS配置
go test -v ./common -run TestAliyunSMSIntegrationSetup

# 测试额度告警逻辑
go test -v ./service -run TestSMSQuotaNotificationContentFormat
```

这个流程确保了额度告警系统的高效、可靠和用户友好性。