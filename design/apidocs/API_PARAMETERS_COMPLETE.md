# Neolink API 平台 - 完整API参数说明文档

**版本**: v1.0
**最后更新**: 2026-01-23
**说明**: 本文档补充了所有API接口的详细请求和响应参数

---

## 目录

- [用户认证模块](#用户认证模块)
  - [用户注册](#用户注册)
  - [用户登录](#用户登录)
  - [双因素认证登录](#双因素认证登录)
  - [获取2FA状态](#获取2fa状态)
  - [设置2FA](#设置2fa)
  - [启用2FA](#启用2fa)
  - [禁用2FA](#禁用2fa)
  - [重新生成备用码](#重新生成备用码)
  - [发送登录验证码](#发送登录验证码)
  - [短信登录](#短信登录)
  - [发送重置验证码](#发送重置验证码)
  - [重置密码](#重置密码)
- [用户管理模块](#用户管理模块)
  - [获取个人信息](#获取个人信息)
  - [更新个人信息](#更新个人信息)
  - [更新用户设置](#更新用户设置)
  - [生成访问令牌](#生成访问令牌)
  - [转移推广额度](#转移推广额度)
  - [发送绑定验证码](#发送绑定验证码)
  - [绑定手机号](#绑定手机号)
  - [解绑手机号](#解绑手机号)
  - [获取所有用户（管理员）](#获取所有用户管理员)
  - [搜索用户](#搜索用户)
  - [创建用户（管理员）](#创建用户管理员)
  - [更新用户（管理员）](#更新用户管理员)
  - [管理用户（批量操作）](#管理用户批量操作)
- [令牌管理模块](#令牌管理模块)
  - [获取令牌列表](#获取令牌列表)
  - [创建令牌](#创建令牌)
  - [更新令牌](#更新令牌)
  - [删除令牌](#删除令牌)
  - [批量删除令牌](#批量删除令牌)
  - [获取令牌使用统计](#获取令牌使用统计)
- [渠道管理模块](#渠道管理模块)
  - [获取所有渠道](#获取所有渠道)
  - [创建渠道](#创建渠道)
  - [更新渠道](#更新渠道)
  - [管理多密钥状态](#管理多密钥状态)
  - [批量设置渠道标签](#批量设置渠道标签)
  - [编辑标签渠道](#编辑标签渠道)
  - [复制渠道](#复制渠道)
- [日志与监控模块](#日志与监控模块)
  - [获取所有日志](#获取所有日志)
  - [获取用户日志](#获取用户日志)
  - [获取日志统计](#获取日志统计)
  - [删除历史日志](#删除历史日志)
- [数据统计模块](#数据统计模块)
  - [获取用户配额数据](#获取用户配额数据)
  - [获取公开配额数据](#获取公开配额数据)
  - [获取公开统计数据](#获取公开统计数据)
- [支付充值模块](#支付充值模块)
  - [获取充值信息](#获取充值信息)
  - [请求易支付](#请求易支付)
  - [请求金额计算](#请求金额计算)
  - [获取用户充值记录](#获取用户充值记录)
  - [管理员补单](#管理员补单)
- [兑换码模块](#兑换码模块)
  - [获取所有兑换码](#获取所有兑换码)
  - [创建兑换码](#创建兑换码)
  - [更新兑换码](#更新兑换码)
  - [删除兑换码](#删除兑换码)
  - [删除失效兑换码](#删除失效兑换码)
- [通用数据结构](#通用数据结构)
  - [User 模型](#user-模型)
  - [Token 模型](#token-模型)
  - [Channel 模型](#channel-模型)
  - [分页响应格式](#分页响应格式)
  - [状态常量](#状态常量)

---

## 用户认证模块

### 用户注册

```
POST /api/user/register
```

#### 请求参数

| 参数名 | 类型 | 必填 | 约束 | 说明 |
|--------|------|------|------|------|
| username | string | 是 | 3-20字符 | 用户名 |
| password | string | 是 | 8-20字符 | 密码 |
| display_name | string | 否 | 最长20字符 | 显示名称 |
| email | string | 否 | 最长50字符 | 邮箱地址 |
| verification_code | string | 条件 | - | 邮箱验证码（开启邮件验证时必填） |
| aff_code | string | 否 | - | 邀请码 |

#### 请求示例

```json
{
  "username": "john_doe",
  "password": "SecurePass123",
  "email": "john@example.com",
  "verification_code": "123456",
  "aff_code": "INVITE2024"
}
```

#### 响应参数

##### 成功响应 (200 OK)

```json
{
  "success": true,
  "message": "注册成功",
  "data": null
}
```

##### 失败响应

| success | message | 说明 |
|---------|---------|------|
| false | "用户名已存在" | 用户名重复 |
| false | "邮箱已被使用" | 邮箱重复 |
| false | "验证码错误" | 邮箱验证码不正确 |
| false | "邀请码无效" | 邀请码不存在或已失效 |
| false | "密码长度不符合要求" | 密码不满足8-20字符 |

---

### 用户登录

```
POST /api/user/login
```

#### 请求参数

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| username | string | 是 | 用户名或邮箱 |
| password | string | 是 | 密码 |

#### 请求示例

```json
{
  "username": "john_doe",
  "password": "SecurePass123"
}
```

#### 响应参数

##### 成功响应（无2FA）

```json
{
  "success": true,
  "message": "登录成功",
  "data": {
    "id": 123,
    "username": "john_doe",
    "display_name": "John Doe",
    "role": 1,
    "status": 1,
    "email": "john@example.com",
    "group": "default",
    "quota": 1000000,
    "used_quota": 50000,
    "aff_code": "ABC123"
  }
}
```

##### 需要2FA验证

```json
{
  "success": false,
  "message": "需要双因素认证",
  "data": {
    "require_2fa": true,
    "session_id": "temp_session_abc123"
  }
}
```

##### 失败响应

| success | message | 说明 |
|---------|---------|------|
| false | "用户名或密码错误" | 凭据无效 |
| false | "用户已被禁用" | 账户被管理员禁用 |
| false | "管理员关闭了密码登录" | 系统禁用了密码登录 |

---

### 双因素认证登录

```
POST /api/user/login/2fa
```

#### 请求参数

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| session_id | string | 是 | 临时会话ID（从登录接口获取） |
| code | string | 是 | 6位验证码或备用码 |

#### 请求示例

```json
{
  "session_id": "temp_session_abc123",
  "code": "123456"
}
```

#### 响应参数

##### 成功响应

```json
{
  "success": true,
  "message": "验证成功",
  "data": {
    "id": 123,
    "username": "john_doe",
    "role": 1,
    "status": 1
  }
}
```

##### 失败响应

```json
{
  "success": false,
  "message": "验证码错误"
}
```

---

### 获取2FA状态

```
GET /api/user/2fa/status
```

**认证**: 👤 UserAuth

#### 响应参数

```json
{
  "success": true,
  "message": "",
  "data": {
    "enabled": true,
    "backup_codes_remaining": 8
  }
}
```

---

### 设置2FA

```
POST /api/user/2fa/setup
```

**认证**: 👤 UserAuth

#### 响应参数

```json
{
  "success": true,
  "message": "",
  "data": {
    "secret": "BASE32_SECRET_KEY_HERE",
    "qr_code": "data:image/png;base64,iVBORw0KG...",
    "backup_codes": [
      "12345678",
      "23456789",
      "34567890",
      "45678901",
      "56789012",
      "67890123",
      "78901234",
      "89012345",
      "90123456",
      "01234567"
    ]
  }
}
```

---

### 启用2FA

```
POST /api/user/2fa/enable
```

**认证**: 👤 UserAuth

#### 请求参数

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| code | string | 是 | 6位验证码 |

#### 请求示例

```json
{
  "code": "123456"
}
```

#### 响应参数

```json
{
  "success": true,
  "message": "双因素认证已启用"
}
```

---

### 禁用2FA

```
POST /api/user/2fa/disable
```

**认证**: 👤 UserAuth

#### 请求参数

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| code | string | 是 | 6位验证码或备用码 |

#### 请求示例

```json
{
  "code": "123456"
}
```

---

### 重新生成备用码

```
POST /api/user/2fa/backup_codes
```

**认证**: 👤 UserAuth

#### 响应参数

```json
{
  "success": true,
  "message": "",
  "data": {
    "backup_codes": [
      "12345678",
      "23456789",
      ...
    ]
  }
}
```

---

### 发送登录验证码

```
POST /api/user/send-sms-code
```

#### 请求参数

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| mobile | string | 是 | 11位手机号 |

#### 请求示例

```json
{
  "mobile": "13800138000"
}
```

#### 响应参数

##### 成功响应

```json
{
  "success": true,
  "message": "验证码已发送",
  "data": {
    "expire_minutes": 5
  }
}
```

##### 失败响应（频率限制）

```json
{
  "success": false,
  "message": "发送过于频繁,请 45 秒后再试"
}
```

---

### 短信登录

```
POST /api/user/sms-login
```

#### 请求参数

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| mobile | string | 是 | 11位手机号 |
| code | string | 是 | 6位验证码 |

#### 请求示例

```json
{
  "mobile": "13800138000",
  "code": "123456"
}
```

#### 响应参数

##### 成功响应

```json
{
  "success": true,
  "message": "",
  "data": {
    "id": 123,
    "username": "user_138000",
    "display_name": "user_138000",
    "role": 1,
    "status": 1,
    "group": "default"
  }
}
```

##### 失败响应

```json
{
  "success": false,
  "message": "验证码不存在或已使用"
}
```

---

### 发送重置验证码

```
POST /api/user/send-reset-code
```

#### 请求参数

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| mobile | string | 是 | 11位手机号 |

#### 请求示例

```json
{
  "mobile": "13900139000"
}
```

#### 响应参数

##### 成功响应

```json
{
  "success": true,
  "message": "验证码已发送",
  "data": {
    "expire_minutes": 5
  }
}
```

##### 失败响应

```json
{
  "success": false,
  "message": "手机号未绑定任何账号"
}
```

---

### 重置密码

```
POST /api/user/reset-password
```

#### 请求参数

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| mobile | string | 是 | 11位手机号 |
| code | string | 是 | 6位验证码 |
| new_password | string | 是 | 新密码（8-20位） |

#### 请求示例

```json
{
  "mobile": "13900139000",
  "code": "111222",
  "new_password": "NewPassword123"
}
```

#### 响应参数

##### 成功响应

```json
{
  "success": true,
  "message": "密码重置成功，请重新登录"
}
```

##### 失败响应

```json
{
  "success": false,
  "message": "密码长度必须为8-20位"
}
```

---

## 用户管理模块

### 获取个人信息

```
GET /api/user/self
```

**认证**: 👤 UserAuth

#### 响应参数

```json
{
  "success": true,
  "message": "",
  "data": {
    "id": 123,
    "username": "john_doe",
    "display_name": "John Doe",
    "role": 1,
    "status": 1,
    "email": "john@example.com",
    "github_id": "",
    "discord_id": "",
    "oidc_id": "",
    "wechat_id": "",
    "telegram_id": "",
    "linux_do_id": "",
    "group": "default",
    "quota": 1000000,
    "used_quota": 50000,
    "request_count": 1500,
    "aff_code": "ABC123",
    "aff_count": 5,
    "aff_quota": 200000,
    "aff_history_quota": 500000,
    "inviter_id": 0,
    "setting": "{}",
    "stripe_customer": "",
    "sidebar_modules": "{}",
    "permissions": {
      "sidebar_settings": true,
      "sidebar_modules": {}
    }
  }
}
```

##### 字段说明

| 字段名 | 类型 | 说明 |
|--------|------|------|
| id | int | 用户ID |
| username | string | 用户名 |
| display_name | string | 显示名称 |
| role | int | 角色（1=普通用户, 2=管理员, 3=超级管理员） |
| status | int | 状态（1=启用, 2=禁用） |
| email | string | 邮箱地址 |
| group | string | 用户分组 |
| quota | int | 总额度 |
| used_quota | int | 已使用额度 |
| request_count | int | 请求次数 |
| aff_code | string | 推广码 |
| aff_count | int | 推广人数 |
| aff_quota | int | 推广可用额度 |
| aff_history_quota | int | 推广历史总额度 |

---

### 更新个人信息

```
PUT /api/user/self
```

**认证**: 👤 UserAuth

#### 请求参数

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| username | string | 否 | 新用户名 |
| display_name | string | 否 | 新显示名称 |
| password | string | 否 | 新密码 |
| original_password | string | 条件 | 修改密码时需提供原密码 |
| sidebar_modules | string | 否 | 侧边栏模块配置（JSON字符串） |

#### 请求示例

```json
{
  "display_name": "John Smith",
  "password": "NewSecurePass456",
  "original_password": "SecurePass123"
}
```

#### 响应参数

```json
{
  "success": true,
  "message": "更新成功"
}
```

---

### 更新用户设置

```
PUT /api/user/setting
```

**认证**: 👤 UserAuth

#### 请求参数

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| notify_type | string | 是 | 通知类型: email\|webhook\|bark\|gotify\|sms |
| quota_warning_threshold | float64 | 是 | 额度预警阈值（必须>0） |
| webhook_url | string | 条件 | Webhook URL（notify_type=webhook时） |
| webhook_secret | string | 否 | Webhook密钥 |
| notification_email | string | 条件 | 通知邮箱（notify_type=email时） |
| bark_url | string | 条件 | Bark推送URL（notify_type=bark时） |
| gotify_url | string | 条件 | Gotify URL（notify_type=gotify时） |
| gotify_token | string | 条件 | Gotify Token |
| gotify_priority | int | 否 | Gotify优先级（0-10，默认5） |
| sms_phone_number | string | 条件 | 手机号（notify_type=sms时，格式: 1[3-9]XXXXXXXXX） |
| accept_unset_model_ratio_model | bool | 否 | 接受未设置比率的模型 |
| record_ip_log | bool | 否 | 记录IP日志 |

#### 请求示例

```json
{
  "notify_type": "email",
  "quota_warning_threshold": 100000,
  "notification_email": "alerts@example.com",
  "record_ip_log": true
}
```

#### 响应参数

```json
{
  "success": true,
  "message": "设置已更新"
}
```

---

### 生成访问令牌

```
GET /api/user/token
```

**认证**: 👤 UserAuth

#### 响应参数

```json
{
  "success": true,
  "message": "",
  "data": "abc123def456ghi789jkl012mno345pqr678stu901vwx"
}
```

**说明**: 返回一个48字符的访问令牌，用于API调用。

---

### 转移推广额度

```
POST /api/user/aff_transfer
```

**认证**: 👤 UserAuth

#### 请求参数

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| quota | int | 是 | 要转移的额度数量 |

#### 请求示例

```json
{
  "quota": 50000
}
```

#### 响应参数

```json
{
  "success": true,
  "message": "额度转移成功"
}
```

---

### 发送绑定验证码

```
POST /api/user/self/send-bind-code
```

**认证**: 👤 UserAuth

#### 请求参数

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| mobile | string | 是 | 11位手机号 |

#### 请求示例

```json
{
  "mobile": "13900139000"
}
```

#### 响应参数

##### 成功响应

```json
{
  "success": true,
  "message": "验证码已发送",
  "data": {
    "expire_minutes": 5
  }
}
```

##### 失败响应

```json
{
  "success": false,
  "message": "该手机号已被其他用户绑定"
}
```

---

### 绑定手机号

```
POST /api/user/self/bind-mobile
```

**认证**: 👤 UserAuth

#### 请求参数

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| mobile | string | 是 | 11位手机号 |
| code | string | 是 | 6位验证码 |

#### 请求示例

```json
{
  "mobile": "13900139000",
  "code": "654321"
}
```

#### 响应参数

##### 成功响应

```json
{
  "success": true,
  "message": "手机号绑定成功",
  "data": {
    "mobile": "13900139000"
  }
}
```

##### 失败响应

```json
{
  "success": false,
  "message": "验证码错误"
}
```

---

### 解绑手机号

```
POST /api/user/self/unbind-mobile
```

**认证**: 👤 UserAuth

#### 响应参数

##### 成功响应

```json
{
  "success": true,
  "message": "手机号解绑成功"
}
```

##### 失败响应

```json
{
  "success": false,
  "message": "用户未绑定手机号"
}
```

---

### 获取所有用户（管理员）

```
GET /api/user/
```

**认证**: 👨‍💼 AdminAuth

#### Query参数

| 参数名 | 类型 | 必填 | 默认值 | 说明 |
|--------|------|------|--------|------|
| p | int | 否 | 0 | 页码（从0开始） |
| size | int | 否 | 10 | 每页数量 |

#### 响应参数

```json
{
  "success": true,
  "message": "",
  "data": [
    {
      "id": 123,
      "username": "john_doe",
      "display_name": "John Doe",
      "role": 1,
      "status": 1,
      "email": "john@example.com",
      "group": "default",
      "quota": 1000000,
      "used_quota": 50000,
      "request_count": 1500
    }
  ]
}
```

---

### 搜索用户

```
GET /api/user/search
```

**认证**: 👨‍💼 AdminAuth

#### Query参数

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| keyword | string | 是 | 搜索关键词（用户名/邮箱/ID） |

---

### 创建用户（管理员）

```
POST /api/user/
```

**认证**: 👨‍💼 AdminAuth

#### 请求参数

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| username | string | 是 | 用户名 |
| password | string | 是 | 密码 |
| display_name | string | 否 | 显示名称 |
| role | int | 否 | 角色（默认1） |
| quota | int | 否 | 初始额度 |
| group | string | 否 | 用户分组 |

#### 请求示例

```json
{
  "username": "new_user",
  "password": "InitialPass123",
  "display_name": "New User",
  "role": 1,
  "quota": 500000,
  "group": "vip"
}
```

---

### 更新用户（管理员）

```
PUT /api/user/
```

**认证**: 👨‍💼 AdminAuth

#### 请求参数

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| id | int | 是 | 用户ID |
| username | string | 否 | 新用户名 |
| password | string | 否 | 新密码（为空则不更新） |
| display_name | string | 否 | 显示名称 |
| role | int | 否 | 角色 |
| status | int | 否 | 状态 |
| quota | int | 否 | 额度 |
| group | string | 否 | 分组 |

#### 请求示例

```json
{
  "id": 123,
  "role": 2,
  "quota": 2000000,
  "group": "admin"
}
```

---

### 管理用户（批量操作）

```
POST /api/user/manage
```

**认证**: 👨‍💼 AdminAuth

#### 请求参数

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| id | int | 是 | 用户ID |
| action | string | 是 | 操作: disable\|enable\|delete\|promote\|demote |

#### 请求示例

```json
{
  "id": 123,
  "action": "disable"
}
```

##### Action说明

| Action | 说明 | 权限要求 |
|--------|------|---------|
| disable | 禁用用户 | AdminAuth |
| enable | 启用用户 | AdminAuth |
| delete | 删除用户 | AdminAuth |
| promote | 提升为管理员 | RootAuth |
| demote | 降级为普通用户 | RootAuth |

---

## 令牌管理模块

### 获取令牌列表

```
GET /api/token/
```

**认证**: 👤 UserAuth

#### Query参数

| 参数名 | 类型 | 必填 | 默认值 | 说明 |
|--------|------|------|--------|------|
| p | int | 否 | 0 | 页码 |
| size | int | 否 | 10 | 每页数量 |

#### 响应参数

```json
{
  "success": true,
  "message": "",
  "data": [
    {
      "id": 1,
      "user_id": 123,
      "key": "sk-abc123...xyz789",
      "status": 1,
      "name": "Production API Key",
      "created_time": 1706054400,
      "accessed_time": 1706140800,
      "expired_time": -1,
      "remain_quota": 100000,
      "unlimited_quota": false,
      "model_limits_enabled": false,
      "model_limits": "",
      "allow_ips": "",
      "used_quota": 50000,
      "group": "default",
      "cross_group_retry": false
    }
  ]
}
```

##### 字段说明

| 字段名 | 类型 | 说明 |
|--------|------|------|
| id | int | 令牌ID |
| key | string | API密钥 |
| status | int | 状态（1=正常, 2=禁用, 3=过期, 4=用尽） |
| name | string | 令牌名称 |
| created_time | int64 | 创建时间戳 |
| accessed_time | int64 | 最后访问时间戳 |
| expired_time | int64 | 过期时间戳（-1=永不过期） |
| remain_quota | int | 剩余额度 |
| unlimited_quota | bool | 无限额度 |
| model_limits_enabled | bool | 是否启用模型限制 |
| model_limits | string | 模型限制（JSON字符串） |
| allow_ips | string | 允许的IP列表（换行分隔） |
| used_quota | int | 已使用额度 |
| cross_group_retry | bool | 跨分组重试 |

---

### 创建令牌

```
POST /api/token/
```

**认证**: 👤 UserAuth

#### 请求参数

| 参数名 | 类型 | 必填 | 约束 | 说明 |
|--------|------|------|------|------|
| name | string | 是 | 最长50字符 | 令牌名称 |
| expired_time | int64 | 否 | - | 过期时间戳（-1=永不过期） |
| remain_quota | int | 条件 | ≥0 | 剩余额度（unlimited_quota=false时必填） |
| unlimited_quota | bool | 否 | - | 无限额度（默认false） |
| model_limits_enabled | bool | 否 | - | 启用模型限制（默认false） |
| model_limits | string | 否 | JSON格式 | 模型限制配置 |
| allow_ips | string | 否 | 换行分隔 | 允许的IP列表 |
| group | string | 否 | - | 所属分组 |
| cross_group_retry | bool | 否 | - | 跨分组重试（默认false） |

#### 请求示例

```json
{
  "name": "Production API Key",
  "expired_time": -1,
  "remain_quota": 1000000,
  "unlimited_quota": false,
  "model_limits_enabled": true,
  "model_limits": "[\"gpt-4\", \"gpt-3.5-turbo\"]",
  "allow_ips": "192.168.1.1\n10.0.0.1",
  "group": "default",
  "cross_group_retry": false
}
```

#### 响应参数

```json
{
  "success": true,
  "message": "",
  "data": {
    "id": 2,
    "key": "sk-abc123def456ghi789jkl012mno345pqr678stu901vwx",
    "name": "Production API Key"
  }
}
```

---

### 更新令牌

```
PUT /api/token/
```

**认证**: 👤 UserAuth

#### Query参数

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| status_only | string | 否 | 设为"true"时只更新状态 |

#### 请求参数

同创建令牌，额外需要 `id` 字段：

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| id | int | 是 | 令牌ID |
| ... | ... | ... | 其他字段同创建令牌 |

#### 请求示例

```json
{
  "id": 1,
  "name": "Updated API Key",
  "status": 1,
  "remain_quota": 2000000
}
```

---

### 删除令牌

```
DELETE /api/token/:id
```

**认证**: 👤 UserAuth

#### Path参数

| 参数名 | 类型 | 说明 |
|--------|------|------|
| id | int | 令牌ID |

#### 响应参数

```json
{
  "success": true,
  "message": "删除成功"
}
```

---

### 批量删除令牌

```
POST /api/token/batch
```

**认证**: 👤 UserAuth

#### 请求参数

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| ids | array | 是 | 令牌ID数组 |

#### 请求示例

```json
{
  "ids": [1, 2, 3, 5, 8]
}
```

---

### 获取令牌使用统计

```
GET /api/usage/token/
```

**认证**: 🔑 TokenAuth
**速率限制**: CriticalRateLimit

#### 请求头

```
Authorization: Bearer sk-abc123...
```

#### 响应参数

```json
{
  "success": true,
  "message": "ok",
  "data": {
    "object": "token_usage",
    "name": "Production API Key",
    "total_granted": 1000000,
    "total_used": 250000,
    "total_available": 750000,
    "unlimited_quota": false,
    "model_limits": {
      "gpt-4": true,
      "gpt-3.5-turbo": true
    },
    "model_limits_enabled": true,
    "expires_at": 1727536000
  }
}
```

---

## 渠道管理模块

### 获取所有渠道

```
GET /api/channel/
```

**认证**: 👨‍💼 AdminAuth

#### Query参数

| 参数名 | 类型 | 必填 | 默认值 | 说明 |
|--------|------|------|--------|------|
| page | int | 否 | 0 | 页码 |
| page_size | int | 否 | 10 | 每页数量 |
| id_sort | bool | 否 | false | 按ID排序 |
| tag_mode | bool | 否 | false | 标签模式 |
| status | string | 否 | "" | 筛选状态: enabled\|disabled\|"" |
| type | int | 否 | 0 | 筛选渠道类型 |

#### 响应参数

```json
{
  "success": true,
  "message": "",
  "data": {
    "items": [
      {
        "id": 1,
        "type": 1,
        "key": "sk-...",
        "status": 1,
        "name": "OpenAI Official",
        "weight": 100,
        "created_time": 1706054400,
        "test_time": 1706140800,
        "response_time": 250,
        "base_url": "",
        "other": "",
        "balance": 10.50,
        "models": "gpt-4,gpt-3.5-turbo",
        "group": "default",
        "used_quota": 500000,
        "priority": 10,
        "tag": "openai"
      }
    ],
    "total": 50,
    "page": 0,
    "page_size": 10,
    "type_counts": {
      "1": 15,
      "2": 10,
      "3": 5
    }
  }
}
```

##### 字段说明

| 字段名 | 类型 | 说明 |
|--------|------|------|
| id | int | 渠道ID |
| type | int | 渠道类型（1=OpenAI, 2=Claude, 3=Gemini等） |
| key | string | API密钥 |
| status | int | 状态（1=启用, 2=禁用） |
| name | string | 渠道名称 |
| weight | uint | 权重（用于负载均衡） |
| response_time | int | 响应时间（毫秒） |
| base_url | string | API基础URL |
| balance | float64 | 余额 |
| models | string | 支持的模型（逗号分隔） |
| group | string | 授权分组 |
| priority | int64 | 优先级 |
| tag | string | 标签 |

---

### 创建渠道

```
POST /api/channel/
```

**认证**: 👨‍💼 AdminAuth

#### 请求参数

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| mode | string | 是 | 模式: single\|multi_to_single\|batch |
| multi_key_mode | string | 条件 | 多密钥模式（mode≠single时） |
| batch_add_set_key_prefix_2_name | bool | 否 | 批量添加时设置密钥前缀为名称 |
| channel | object | 是 | 渠道对象（见下表） |

##### channel对象参数

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| type | int | 是 | 渠道类型 |
| key | string | 是 | API密钥（多密钥用\\n分隔） |
| status | int | 否 | 状态（默认1） |
| name | string | 是 | 渠道名称 |
| weight | uint | 否 | 权重 |
| base_url | string | 否 | API基础URL |
| models | string | 是 | 支持的模型（逗号分隔） |
| group | string | 否 | 授权分组（逗号分隔） |
| model_mapping | string | 否 | 模型映射（JSON） |
| priority | int64 | 否 | 优先级 |
| auto_ban | int | 否 | 自动封禁（0=关闭, 1=开启） |
| tag | string | 否 | 标签 |

#### 请求示例（单密钥模式）

```json
{
  "mode": "single",
  "channel": {
    "type": 1,
    "key": "sk-abc123...",
    "name": "OpenAI Production",
    "weight": 100,
    "base_url": "https://api.openai.com/v1",
    "models": "gpt-4,gpt-3.5-turbo",
    "group": "default,vip",
    "priority": 10,
    "tag": "openai"
  }
}
```

#### 请求示例（批量创建）

```json
{
  "mode": "batch",
  "batch_add_set_key_prefix_2_name": true,
  "channel": {
    "type": 1,
    "key": "sk-key1...\nsk-key2...\nsk-key3...",
    "name": "OpenAI Batch",
    "models": "gpt-4,gpt-3.5-turbo",
    "group": "default"
  }
}
```

#### 响应参数

```json
{
  "success": true,
  "message": "渠道创建成功",
  "data": null
}
```

---

### 更新渠道

```
PUT /api/channel/
```

**认证**: 👨‍💼 AdminAuth

#### 请求参数

同创建渠道，但channel对象需包含 `id` 字段。额外参数：

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| multi_key_mode | string | 否 | 更新多密钥模式 |
| key_mode | string | 否 | 密钥模式: append\|replace（多密钥时） |

#### 请求示例

```json
{
  "channel": {
    "id": 1,
    "name": "Updated Channel Name",
    "weight": 150,
    "models": "gpt-4,gpt-3.5-turbo,claude-3-opus"
  }
}
```

---

### 管理多密钥状态

```
POST /api/channel/multi_key/manage
```

**认证**: 👨‍💼 AdminAuth

#### 请求参数

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| channel_id | int | 是 | 渠道ID |
| action | string | 是 | 操作（见下表） |
| key_index | int | 条件 | 密钥索引（部分操作需要） |
| page | int | 否 | 页码（get_key_status时） |
| page_size | int | 否 | 每页数量 |
| status | int | 否 | 筛选状态（1=启用, 2=手动禁用, 3=自动禁用） |

##### Action操作说明

| Action | 说明 | 需要key_index |
|--------|------|--------------|
| get_key_status | 获取密钥状态 | 否 |
| disable_key | 禁用指定密钥 | 是 |
| enable_key | 启用指定密钥 | 是 |
| delete_key | 删除指定密钥 | 是 |
| delete_disabled_keys | 删除所有禁用密钥 | 否 |
| enable_all_keys | 启用所有密钥 | 否 |
| disable_all_keys | 禁用所有密钥 | 否 |

#### 请求示例（获取密钥状态）

```json
{
  "channel_id": 1,
  "action": "get_key_status",
  "page": 1,
  "page_size": 20
}
```

#### 响应参数（get_key_status）

```json
{
  "success": true,
  "message": "",
  "data": {
    "keys": [
      {
        "index": 0,
        "status": 1,
        "key_preview": "sk-abc123...",
        "disabled_time": 0,
        "reason": ""
      },
      {
        "index": 1,
        "status": 3,
        "key_preview": "sk-def456...",
        "disabled_time": 1706054400,
        "reason": "API错误: 余额不足"
      }
    ],
    "total": 10,
    "page": 1,
    "page_size": 20,
    "total_pages": 1,
    "enabled_count": 7,
    "manual_disabled_count": 1,
    "auto_disabled_count": 2
  }
}
```

---

### 批量设置渠道标签

```
POST /api/channel/batch/tag
```

**认证**: 👨‍💼 AdminAuth

#### 请求参数

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| ids | array | 是 | 渠道ID数组 |
| tag | string | 否 | 标签（null表示删除标签） |

#### 请求示例

```json
{
  "ids": [1, 2, 3, 5],
  "tag": "high-priority"
}
```

---

### 编辑标签渠道

```
PUT /api/channel/tag
```

**认证**: 👨‍💼 AdminAuth

#### 请求参数

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| tag | string | 是 | 当前标签 |
| new_tag | string | 否 | 新标签名 |
| priority | int64 | 否 | 优先级 |
| weight | uint | 否 | 权重 |
| model_mapping | string | 否 | 模型映射（JSON） |
| models | string | 否 | 模型列表 |
| groups | string | 否 | 分组列表 |
| param_override | string | 否 | 参数覆盖（JSON） |
| header_override | string | 否 | 请求头覆盖（JSON） |

#### 请求示例

```json
{
  "tag": "openai",
  "priority": 20,
  "weight": 150,
  "models": "gpt-4,gpt-4-turbo,gpt-3.5-turbo"
}
```

---

### 复制渠道

```
POST /api/channel/copy/:id
```

**认证**: 👨‍💼 AdminAuth

#### Path参数

| 参数名 | 类型 | 说明 |
|--------|------|------|
| id | int | 源渠道ID |

#### Query参数

| 参数名 | 类型 | 必填 | 默认值 | 说明 |
|--------|------|------|--------|------|
| suffix | string | 否 | "_复制" | 名称后缀 |
| reset_balance | bool | 否 | true | 是否重置余额 |

#### 响应参数

```json
{
  "success": true,
  "message": "复制成功",
  "data": {
    "id": 101
  }
}
```

---

## 日志与监控模块

### 获取所有日志

```
GET /api/log/
```

**认证**: 👨‍💼 AdminAuth

#### Query参数

| 参数名 | 类型 | 必填 | 默认值 | 说明 |
|--------|------|------|--------|------|
| page | int | 否 | 0 | 页码 |
| page_size | int | 否 | 10 | 每页数量 |
| type | int | 否 | 0 | 日志类型（0=全部, 1=充值, 2=消费, 3=管理, 4=系统） |
| start_timestamp | int64 | 否 | - | 开始时间戳 |
| end_timestamp | int64 | 否 | - | 结束时间戳 |
| username | string | 否 | - | 用户名 |
| token_name | string | 否 | - | 令牌名称 |
| model_name | string | 否 | - | 模型名称 |
| channel | int | 否 | 0 | 渠道ID |
| group | string | 否 | - | 分组 |

#### 响应参数

```json
{
  "success": true,
  "message": "",
  "data": [
    {
      "id": 10240,
      "user_id": 123,
      "created_at": 1706054400,
      "type": 2,
      "content": "使用 gpt-4 消费 100 额度",
      "username": "john_doe",
      "token_name": "Production API",
      "model_name": "gpt-4",
      "quota": 100,
      "prompt_tokens": 50,
      "completion_tokens": 30,
      "use_time": 1200,
      "is_stream": false,
      "channel": 1,
      "channel_name": "OpenAI Official",
      "token_id": 1,
      "group": "default",
      "ip": "192.168.1.1",
      "other": "{}"
    }
  ]
}
```

##### 字段说明

| 字段名 | 类型 | 说明 |
|--------|------|------|
| id | int | 日志ID |
| user_id | int | 用户ID |
| created_at | int64 | 创建时间戳 |
| type | int | 日志类型 |
| content | string | 日志内容 |
| username | string | 用户名 |
| token_name | string | 令牌名称 |
| model_name | string | 模型名称 |
| quota | int | 消费额度 |
| prompt_tokens | int | 输入Token数 |
| completion_tokens | int | 输出Token数 |
| use_time | int | 使用时长（毫秒） |
| is_stream | bool | 是否流式 |
| channel | int | 渠道ID |
| channel_name | string | 渠道名称 |
| ip | string | IP地址 |

---

### 获取用户日志

```
GET /api/log/self
```

**认证**: 👤 UserAuth

#### Query参数

同"获取所有日志"，但不包括 `username` 和 `channel` 参数。

---

### 获取日志统计

```
GET /api/log/stat
```

**认证**: 👨‍💼 AdminAuth

#### Query参数

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| type | int | 否 | 日志类型 |
| start_timestamp | int64 | 否 | 开始时间戳 |
| end_timestamp | int64 | 否 | 结束时间戳 |
| username | string | 否 | 用户名 |
| token_name | string | 否 | 令牌名称 |
| model_name | string | 否 | 模型名称 |
| channel | int | 否 | 渠道ID |

#### 响应参数

```json
{
  "success": true,
  "message": "",
  "data": {
    "quota": 1500000,
    "rpm": 1200,
    "tpm": 75000
  }
}
```

##### 字段说明

| 字段名 | 说明 |
|--------|------|
| quota | 总消费额度 |
| rpm | 请求/分钟（Requests Per Minute） |
| tpm | Token/分钟（Tokens Per Minute） |

---

### 删除历史日志

```
DELETE /api/log/
```

**认证**: 👨‍💼 AdminAuth

#### Query参数

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| target_timestamp | int64 | 是 | 删除该时间戳之前的所有日志 |

#### 响应参数

```json
{
  "success": true,
  "message": "",
  "data": 5000
}
```

**说明**: data字段返回删除的记录数。

---

## 数据统计模块

### 获取用户配额数据

```
GET /api/data/self
```

**认证**: 👤 UserAuth

#### Query参数

| 参数名 | 类型 | 必填 | 默认值 | 说明 |
|--------|------|------|--------|------|
| start_timestamp | int64 | 否 | 当前-30天 | 开始时间戳 |
| end_timestamp | int64 | 否 | 当前时间 | 结束时间戳 |

#### 参数约束

- 时间跨度不能超过30天（2,592,000秒）

#### 响应参数

```json
{
  "success": true,
  "message": "",
  "data": [
    {
      "id": 1,
      "user_id": 123,
      "username": "john_doe",
      "model_name": "gpt-4",
      "created_at": 1706054400,
      "token_used": 150,
      "count": 10,
      "quota": 15000
    },
    {
      "id": 2,
      "user_id": 123,
      "username": "john_doe",
      "model_name": "gpt-3.5-turbo",
      "created_at": 1706054400,
      "token_used": 500,
      "count": 30,
      "quota": 5000
    }
  ]
}
```

---

### 获取公开配额数据

```
GET /api/data/public
```

**认证**: 🌐 Public（无需认证）

#### Query参数

| 参数名 | 类型 | 必填 | 默认值 | 说明 |
|--------|------|------|--------|------|
| start_timestamp | int64 | 否 | 当前-7天 | 开始时间戳 |
| end_timestamp | int64 | 否 | 当前时间 | 结束时间戳 |

#### 参数约束

- `start_timestamp` 和 `end_timestamp` 需同时提供或同时省略
- 时间跨度不能超过31天（2,678,400秒）
- `start_timestamp` 必须小于 `end_timestamp`

#### 响应参数

```json
{
  "success": true,
  "message": "",
  "data": [
    {
      "model_name": "gpt-4",
      "count": 1523,
      "quota": 152300,
      "token_used": 98500,
      "created_at": 1706054400
    },
    {
      "model_name": "gpt-3.5-turbo",
      "count": 3821,
      "quota": 38210,
      "token_used": 256000,
      "created_at": 1706054400
    }
  ]
}
```

##### 数据聚合规则

- 按 `model_name` 和 `created_at`（精确到小时）分组聚合
- `count`、`quota`、`token_used` 为该时间段内的累计值
- 时间戳已对齐到小时边界（分钟和秒为00）

---

### 获取公开统计数据

```
GET /api/data/public/stats
```

**认证**: 🌐 Public（无需认证）

#### Query参数

| 参数名 | 类型 | 必填 | 默认值 | 说明 |
|--------|------|------|--------|------|
| start_timestamp | int64 | 否 | 当前-24小时 | Top用户查询开始时间 |
| end_timestamp | int64 | 否 | 当前时间 | Top用户查询结束时间 |

#### 参数约束

- 时间范围不能超过31天
- 时间范围仅影响Top用户排行，不影响基础统计指标

#### 响应参数

```json
{
  "success": true,
  "message": "",
  "data": {
    "stats": {
      "enabled_models_count": 127,
      "enabled_channels_count": 15,
      "active_tokens_count": 342,
      "today_token_usage": 1250000,
      "total_req_count": 150000,
      "total_quota": 50000000,
      "total_token_usage": 75000000,
      "total_data_count": 25000
    },
    "top_users": [
      {
        "username": "joh***oe",
        "quota": 150000,
        "token_used": 980000,
        "request_count": 452
      },
      {
        "username": "ali***ce",
        "quota": 120000,
        "token_used": 750000,
        "request_count": 389
      }
    ]
  }
}
```

##### stats字段说明

| 字段名 | 类型 | 说明 |
|--------|------|------|
| enabled_models_count | int64 | 当前启用的模型数量 |
| enabled_channels_count | int64 | 当前启用的渠道（服务商）数量 |
| active_tokens_count | int64 | 当前有效的令牌数量 |
| today_token_usage | int64 | 今日Token消耗量（按自然日0:00-23:59统计，UTC+8） |
| total_req_count | int64 | 总请求数（从quota_data表汇总） |
| total_quota | int64 | 总消耗额度（从quota_data表汇总） |
| total_token_usage | int64 | 总Token消耗（从quota_data表汇总） |
| total_data_count | int64 | 总数据记录数（quota_data表记录总数） |

##### top_users字段说明

| 字段名 | 类型 | 说明 |
|--------|------|------|
| username | string | 匿名化处理后的用户名 |
| quota | int64 | 该用户在指定时间范围内的总消耗额度 |
| token_used | int64 | 该用户在指定时间范围内的总Token消耗 |
| request_count | int64 | 该用户在指定时间范围内的总请求次数 |

##### 用户名匿名化规则

| 原始长度 | 匿名化规则 | 示例 |
|---------|-----------|------|
| 0（空） | "匿名用户" | `` → `匿名用户` |
| ≤ 2 | 保留首字符 + "***" | `ab` → `a***` |
| 3-4 | 保留首尾字符 + "***" | `alice` → `a***e` |
| ≥ 5 | 保留前3字符 + "***" + 保留后2字符 | `john_doe` → `joh***oe` |

---

## 支付充值模块

### 获取充值信息

```
GET /api/user/topup/info
```

**认证**: 👤 UserAuth

#### 响应参数

```json
{
  "success": true,
  "message": "",
  "data": {
    "enable_online_topup": true,
    "enable_stripe_topup": true,
    "enable_creem_topup": false,
    "creem_products": "[]",
    "pay_methods": [
      {
        "name": "支付宝",
        "type": "alipay",
        "color": "rgba(0, 122, 255, 1)"
      },
      {
        "name": "微信支付",
        "type": "wechat",
        "color": "rgba(9, 187, 7, 1)"
      }
    ],
    "min_topup": 100,
    "stripe_min_topup": 5,
    "amount_options": [100, 500, 1000, 5000, 10000],
    "discount": {
      "1000": 0.95,
      "5000": 0.90,
      "10000": 0.85
    }
  }
}
```

---

### 请求易支付

```
POST /api/user/pay
```

**认证**: 👤 UserAuth
**速率限制**: CriticalRateLimit

#### 请求参数

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| amount | int64 | 是 | 充值数量（额度） |
| payment_method | string | 是 | 支付方式: alipay\|wechat |
| top_up_code | string | 否 | 优惠码 |

#### 请求示例

```json
{
  "amount": 1000,
  "payment_method": "alipay",
  "top_up_code": "DISCOUNT2024"
}
```

#### 响应参数

```json
{
  "success": true,
  "message": "success",
  "data": {
    "trade_no": "USR123NO20240124...",
    "amount": 1000,
    "money": 9.50
  },
  "url": "https://payment.example.com/pay?order=..."
}
```

---

### 请求金额计算

```
POST /api/user/amount
```

**认证**: 👤 UserAuth

#### 请求参数

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| amount | int64 | 是 | 充值数量 |
| top_up_code | string | 否 | 优惠码 |

#### 请求示例

```json
{
  "amount": 1000,
  "top_up_code": "DISCOUNT2024"
}
```

#### 响应参数

```json
{
  "success": true,
  "message": "success",
  "data": "9.50"
}
```

**说明**: data字段返回需要支付的金额（字符串格式）。

---

### 获取用户充值记录

```
GET /api/user/topup/self
```

**认证**: 👤 UserAuth

#### Query参数

| 参数名 | 类型 | 必填 | 默认值 | 说明 |
|--------|------|------|--------|------|
| page | int | 否 | 0 | 页码 |
| page_size | int | 否 | 10 | 每页数量 |
| keyword | string | 否 | - | 搜索关键词 |

#### 响应参数

```json
{
  "success": true,
  "message": "",
  "data": [
    {
      "id": 1,
      "user_id": 123,
      "amount": 1000,
      "money": 9.50,
      "trade_no": "USR123NO20240124...",
      "payment_method": "alipay",
      "create_time": 1706054400,
      "complete_time": 1706054500,
      "status": "success"
    }
  ]
}
```

##### 字段说明

| 字段名 | 类型 | 说明 |
|--------|------|------|
| id | int | 充值记录ID |
| user_id | int | 用户ID |
| amount | int | 充值额度 |
| money | float64 | 支付金额 |
| trade_no | string | 交易号 |
| payment_method | string | 支付方式 |
| create_time | int64 | 创建时间戳 |
| complete_time | int64 | 完成时间戳 |
| status | string | 状态: pending\|success\|failed |

---

### 管理员补单

```
POST /api/user/topup/complete
```

**认证**: 👨‍💼 AdminAuth

#### 请求参数

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| trade_no | string | 是 | 交易号 |

#### 请求示例

```json
{
  "trade_no": "USR123NO20240124..."
}
```

#### 响应参数

```json
{
  "success": true,
  "message": "补单成功"
}
```

---

## 兑换码模块

### 获取所有兑换码

```
GET /api/redemption/
```

**认证**: 👨‍💼 AdminAuth

#### Query参数

| 参数名 | 类型 | 必填 | 默认值 | 说明 |
|--------|------|------|--------|------|
| page | int | 否 | 0 | 页码 |
| page_size | int | 否 | 10 | 每页数量 |

#### 响应参数

```json
{
  "success": true,
  "message": "",
  "data": [
    {
      "id": 1,
      "user_id": 123,
      "key": "abc123def456ghi789...",
      "status": 1,
      "name": "Summer Promotion",
      "quota": 50000,
      "created_time": 1706054400,
      "redeemed_time": 0,
      "used_user_id": 0,
      "expired_time": 1727536000
    }
  ]
}
```

##### 字段说明

| 字段名 | 类型 | 说明 |
|--------|------|------|
| id | int | 兑换码ID |
| user_id | int | 创建者用户ID |
| key | string | 兑换码 |
| status | int | 状态（1=未使用, 2=已使用, 3=已过期） |
| name | string | 名称 |
| quota | int | 额度 |
| created_time | int64 | 创建时间戳 |
| redeemed_time | int64 | 兑换时间戳 |
| used_user_id | int | 使用者用户ID |
| expired_time | int64 | 过期时间戳（0=不过期） |

---

### 创建兑换码

```
POST /api/redemption/
```

**认证**: 👨‍💼 AdminAuth

#### 请求参数

| 参数名 | 类型 | 必填 | 约束 | 说明 |
|--------|------|------|------|------|
| name | string | 是 | 1-20字符 | 兑换码名称 |
| quota | int | 是 | >0 | 每个兑换码的额度 |
| count | int | 是 | 1-100 | 生成的个数 |
| expired_time | int64 | 否 | - | 过期时间戳（0=不过期） |

#### 请求示例

```json
{
  "name": "New Year 2024",
  "quota": 100000,
  "count": 50,
  "expired_time": 1735660800
}
```

#### 响应参数

```json
{
  "success": true,
  "message": "",
  "data": [
    "abc123def456ghi789jkl012mno345pqr678stu901vwx",
    "bcd234efg567hij890klm123nop456qrs789tuv012wxy",
    ...
  ]
}
```

**说明**: data字段返回生成的所有兑换码数组。

---

### 更新兑换码

```
PUT /api/redemption/
```

**认证**: 👨‍💼 AdminAuth

#### Query参数

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| status_only | string | 否 | 设为"true"时只更新状态 |

#### 请求参数

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| id | int | 是 | 兑换码ID |
| name | string | 否 | 名称 |
| quota | int | 否 | 额度 |
| status | int | 否 | 状态 |
| expired_time | int64 | 否 | 过期时间戳 |

#### 请求示例

```json
{
  "id": 1,
  "name": "Spring Sale 2024",
  "quota": 150000,
  "expired_time": 1740758400
}
```

---

### 删除兑换码

```
DELETE /api/redemption/:id
```

**认证**: 👨‍💼 AdminAuth

#### Path参数

| 参数名 | 类型 | 说明 |
|--------|------|------|
| id | int | 兑换码ID |

---

### 删除失效兑换码

```
DELETE /api/redemption/invalid
```

**认证**: 👨‍💼 AdminAuth

#### 响应参数

```json
{
  "success": true,
  "message": "",
  "data": 25
}
```

**说明**: data字段返回删除的兑换码数量。

---

## 通用数据结构

### User 模型

```go
{
  "id": int,
  "username": string,
  "password": string,           // 返回时不包含
  "display_name": string,
  "role": int,                  // 1=普通用户, 2=管理员, 3=超级管理员
  "status": int,                // 1=启用, 2=禁用
  "email": string,
  "github_id": string,
  "discord_id": string,
  "oidc_id": string,
  "wechat_id": string,
  "telegram_id": string,
  "linux_do_id": string,
  "group": string,              // 用户分组
  "quota": int,                 // 总额度
  "used_quota": int,            // 已使用额度
  "request_count": int,         // 请求次数
  "aff_code": string,          // 推广码
  "aff_count": int,            // 推广人数
  "aff_quota": int,            // 推广可用额度
  "aff_history_quota": int,    // 推广历史总额度
  "inviter_id": int,           // 邀请人ID
  "setting": string,           // JSON格式设置
  "stripe_customer": string    // Stripe客户ID
}
```

### Token 模型

```go
{
  "id": int,
  "user_id": int,
  "key": string,                      // API密钥
  "status": int,                      // 1=正常, 2=禁用, 3=过期, 4=用尽
  "name": string,                     // 令牌名称
  "created_time": int64,              // 创建时间戳
  "accessed_time": int64,             // 最后访问时间戳
  "expired_time": int64,              // 过期时间戳(-1=永不过期)
  "remain_quota": int,                // 剩余额度
  "unlimited_quota": bool,            // 无限额度
  "model_limits_enabled": bool,       // 启用模型限制
  "model_limits": string,             // 模型限制(JSON)
  "allow_ips": string,                // 允许的IP列表(换行分隔)
  "used_quota": int,                  // 已使用额度
  "group": string,                    // 所属分组
  "cross_group_retry": bool           // 跨分组重试
}
```

### Channel 模型

```go
{
  "id": int,
  "type": int,                        // 渠道类型
  "key": string,                      // API密钥
  "status": int,                      // 1=启用, 2=禁用
  "name": string,                     // 渠道名称
  "weight": uint,                     // 权重
  "created_time": int64,              // 创建时间戳
  "test_time": int64,                 // 测试时间戳
  "response_time": int,               // 响应时间(毫秒)
  "base_url": string,                 // API基础URL
  "other": string,                    // 其他配置
  "balance": float64,                 // 余额
  "models": string,                   // 支持的模型(逗号分隔)
  "group": string,                    // 授权分组
  "used_quota": int,                  // 已使用额度
  "priority": int64,                  // 优先级
  "tag": string,                      // 标签
  "model_mapping": string,            // 模型映射(JSON)
  "status_code_mapping": string,      // 状态码映射(JSON)
  "auto_ban": int,                    // 自动封禁
  "setting": string,                  // 设置(JSON)
  "param_override": string,           // 参数覆盖(JSON)
  "header_override": string,          // 请求头覆盖(JSON)
  "remark": string                    // 备注
}
```

### 分页响应格式

```json
{
  "success": true,
  "message": "",
  "data": {
    "items": [...],
    "total": 100,
    "page": 1,
    "page_size": 20
  }
}
```

### 状态常量

#### 用户角色

| 值 | 说明 |
|----|------|
| 1 | 普通用户 |
| 2 | 管理员 |
| 3 | 超级管理员 |

#### 用户状态

| 值 | 说明 |
|----|------|
| 1 | 启用 |
| 2 | 禁用 |

#### 令牌状态

| 值 | 说明 |
|----|------|
| 1 | 正常 |
| 2 | 禁用 |
| 3 | 已过期 |
| 4 | 额度已用尽 |

#### 渠道状态

| 值 | 说明 |
|----|------|
| 1 | 启用 |
| 2 | 禁用 |

#### 日志类型

| 值 | 说明 |
|----|------|
| 0 | 全部 |
| 1 | 充值 |
| 2 | 消费 |
| 3 | 管理 |
| 4 | 系统 |

---

**文档结束**
