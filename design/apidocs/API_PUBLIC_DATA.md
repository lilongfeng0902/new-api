# API接口文档

本文档定义了New API平台的API接口规范，包括公开数据接口等模块。

**基础URL**: `https://your-domain.com`
**协议**: HTTPS
**字符编码**: UTF-8

## 目录

- [公开数据模块](#公开数据模块)
  - [1. 获取公开数据看板图表数据](#1-获取公开数据看板图表数据)
  - [2. 获取公开统计数据](#2-获取公开统计数据)
  - [附录A: 快速开始](#附录a-快速开始)
  - [附录B: 常见问题](#附录b-常见问题)
  - [附录C: 通用响应格式](#附录c-通用响应格式)
  - [附录D: 错误码说明](#附录d-错误码说明)
  - [附录E: 版本历史](#附录e-版本历史)
  - [附录F: 技术支持](#附录f-技术支持)

---

## 公开数据模块

本模块提供公开数据相关接口，这些接口无需认证即可访问，用于公开数据看板展示。

### 通用说明

#### 速率限制
- **限制**: 100 请求/分钟/IP
- **超限响应**: HTTP 429 Too Many Requests

#### 时间规范
- 所有时间戳为 **Unix 秒级时间戳**（10位数字）
- **时区**: UTC+0（协调世界时）
- 示例: `1706054400` = `2024-01-24 10:00:00 UTC`

#### 响应格式
所有接口返回 HTTP 200 OK，业务成功/失败通过响应体中的 `success` 字段判断。

---

### 1. 获取公开数据看板图表数据

#### 1.1 接口概述

获取公开数据看板的时间序列数据，用于绘制消费趋势、模型占比等图表。该接口返回按模型和时间分组的聚合数据。

#### 1.2 基本信息

| 项目 | 说明 |
|------|------|
| 接口名称 | 获取公开数据看板图表数据 |
| 接口地址 | `/api/data/public` |
| 请求方式 | `GET` |
| 认证方式 | 无需认证（公开接口） |
| 接口版本 | v1.0 |

#### 1.3 请求参数

##### Query参数

| 参数名 | 类型 | 必填 | 默认值 | 说明 |
|--------|------|------|--------|------|
| start_timestamp | int64 | 否 | 当前时间-7天 | 查询开始时间，Unix秒级时间戳 |
| end_timestamp | int64 | 否 | 当前时间 | 查询结束时间，Unix秒级时间戳 |

##### 参数约束

- `start_timestamp` 和 `end_timestamp` 需同时提供或同时省略
- 时间范围不能超过31天（2,678,400秒）
- 时间戳为Unix秒级时间戳（10位数字）

#### 1.4 响应参数

##### 响应数据结构

| 字段名 | 类型 | 说明 |
|--------|------|------|
| success | boolean | 请求是否成功 |
| message | string | 错误信息，成功时为空字符串 |
| data | array | 数据记录数组 |

##### data数组元素字段

| 字段名 | 类型 | 说明 |
|--------|------|------|
| model_name | string | 模型名称 |
| count | int | 该模型在该时间点的总请求次数 |
| quota | int | 该模型在该时间点的总消耗额度 |
| token_used | int | 该模型在该时间点的总Token消耗 |
| created_at | int64 | 时间戳（Unix秒级），精确到小时 |

##### 数据聚合规则

- 按 `model_name` 和 `created_at`（精确到小时）分组聚合
- `count`、`quota`、`token_used` 字段为该时间段内的累计值
- 时间戳已对齐到小时边界，即分钟和秒都为00
  - 示例：`1706054400` 表示 `2024-01-24 10:00:00`（整点）
  - 说明：该小时内的所有请求都会聚合到这个整点时间戳下

#### 1.5 响应示例

##### 成功响应

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
    },
    {
      "model_name": "claude-3-opus",
      "count": 892,
      "quota": 89200,
      "token_used": 67500,
      "created_at": 1706054400
    }
  ]
}
```

##### 错误响应

```json
{
  "success": false,
  "message": "时间跨度不能超过31天"
}
```

---

### 2. 获取公开统计数据

#### 2.1 接口概述

获取公开数据统计信息，包括平台基础统计指标和消费Top用户排行。该接口返回实时统计数据，用于数据看板的统计卡片展示。

#### 2.2 基本信息

| 项目 | 说明 |
|------|------|
| 接口名称 | 获取公开统计数据 |
| 接口地址 | `/api/data/public/stats` |
| 请求方式 | `GET` |
| 认证方式 | 无需认证（公开接口） |
| 接口版本 | v1.0 |

#### 2.3 请求参数

##### Query参数

| 参数名 | 类型 | 必填 | 默认值 | 说明 |
|--------|------|------|--------|------|
| start_timestamp | int64 | 否 | 当前时间-24小时 | 查询开始时间，Unix秒级时间戳（用于Top用户查询） |
| end_timestamp | int64 | 否 | 当前时间 | 查询结束时间，Unix秒级时间戳（用于Top用户查询） |

##### 参数约束

- `start_timestamp` 和 `end_timestamp` 需同时提供或同时省略
- 时间范围不能超过31天（2,678,400秒）
- 时间戳为Unix秒级时间戳（10位数字）
- 时间范围仅影响Top用户排行，不影响基础统计指标

#### 2.4 响应参数

##### 响应数据结构

| 字段名 | 类型 | 说明 |
|--------|------|------|
| success | boolean | 请求是否成功 |
| message | string | 错误信息，成功时为空字符串 |
| data | object | 响应数据对象 |
| data.stats | object | 统计数据对象 |
| data.top_users | array | Top用户列表 |

##### stats对象字段

| 字段名 | 类型 | 说明 |
|--------|------|------|
| enabled_models_count | int64 | 当前启用的模型数量 |
| enabled_channels_count | int64 | 当前启用的渠道（服务商）数量 |
| active_tokens_count | int64 | 当前有效的令牌数量 |
| today_token_usage | int64 | 今日Token消耗量（按自然日0:00-23:59统计） |
| total_req_count | int64 | 总请求数（从quota_data表汇总） |
| total_quota | int64 | 总消耗额度（从quota_data表汇总） |
| total_token_usage | int64 | 总Token消耗（从quota_data表汇总） |
| total_data_count | int64 | 总数据记录数（quota_data表记录总数） |

##### top_users数组元素字段

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

#### 2.5 响应示例

##### 成功响应

```json
{
  "success": true,
  "message": "",
  "data": {
    "stats": {
      "enabled_models_count": 127,
      "enabled_channels_count": 15,
      "active_tokens_count": 342,
      "today_token_usage": 1250000
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
      },
      {
        "username": "bob***er",
        "quota": 98000,
        "token_used": 650000,
        "request_count": 312
      }
    ]
  }
}
```

##### 错误响应

```json
{
  "success": false,
  "message": "时间跨度不能超过31天"
}
```

---

### 附录A: 快速开始

#### cURL 示例

```bash
# 获取图表数据（默认最近7天）
curl -X GET "https://your-domain.com/api/data/public"

# 获取图表数据（指定时间范围）
curl -X GET "https://your-domain.com/api/data/public?start_timestamp=1706054400&end_timestamp=1706658000"

# 获取统计数据
curl -X GET "https://your-domain.com/api/data/public/stats"
```

#### JavaScript 示例

```javascript
// 获取图表数据
async function getChartData(startTimestamp, endTimestamp) {
  const params = new URLSearchParams();
  if (startTimestamp && endTimestamp) {
    params.append('start_timestamp', startTimestamp);
    params.append('end_timestamp', endTimestamp);
  }

  const response = await fetch(
    `https://your-domain.com/api/data/public?${params}`
  );
  const result = await response.json();

  if (!result.success) {
    throw new Error(result.message);
  }

  return result.data;
}

// 使用示例
const sevenDaysAgo = Math.floor(Date.now() / 1000) - (7 * 24 * 60 * 60);
const now = Math.floor(Date.now() / 1000);
const data = await getChartData(sevenDaysAgo, now);
```

---

### 附录B: 常见问题

**Q1: 时间戳应该使用什么时区？**
A: 所有时间戳均为 UTC+0（协调世界时）。

**Q2: 为什么返回的数据时间戳都是整点？**
A: 数据按小时粒度聚合，时间戳已对齐到小时边界。

**Q3: 速率限制是多少？**
A: 100 请求/分钟/IP。超限返回 HTTP 429。

**Q4: 是否支持CORS跨域请求？**
A: 是的，公开接口支持CORS，可从浏览器直接调用。

---

### 附录C: 通用响应格式

所有公开数据API接口均遵循统一的响应格式：

```json
{
  "success": boolean,   // 请求是否成功
  "message": string,    // 错误信息，成功时为空字符串
  "data": any          // 响应数据，格式根据接口而定
}
```

#### 成功响应

- `success`: `true`
- `message`: 空字符串 `""`
- `data`: 包含实际数据

#### 失败响应

- `success`: `false`
- `message`: 具体错误信息
- `data`: 通常为 `null` 或不返回

---

### 附录D: 错误码说明

#### HTTP状态码

| HTTP状态码 | 说明 | 处理建议 |
|-----------|------|---------|
| 200 | 请求成功（业务成功或失败通过 `success` 字段判断） | 检查响应体中的 `success` 字段 |
| 429 | 请求过于频繁（超过速率限制） | 降低请求频率，稍后重试 |
| 500 | 服务器内部错误 | 稍后重试，持续失败请联系技术支持 |
| 503 | 服务暂时不可用 | 系统维护中，稍后重试 |

#### 业务错误消息

| 错误消息 | 说明 | 解决方案 |
|---------|------|---------|
| 时间跨度不能超过31天 | 请求的时间范围超过31天限制 | 缩短查询时间范围 |
| 数据库查询失败: ... | 数据库连接或查询异常 | 检查数据库状态，稍后重试 |

---

### 附录E: 版本历史

| 版本 | 日期 | 变更内容 | 变更人 |
|------|------|---------|--------|
| v1.0 | 2026-01-22 | 初始版本，定义两个公开数据接口 | Development Team |
| v1.1 | 2026-01-23 | 添加基础URL、速率限制、时间规范说明；补充代码示例和常见问题 | Documentation Team |

---

### 附录F: 技术支持

#### 联系方式

- **GitHub Issues**: https://github.com/QuantumNous/new-api/issues
- **技术支持邮箱**: support@quantumnous.com

#### 相关资源

- **项目主页**: https://github.com/QuantumNous/new-api
- **完整文档**: https://docs.new-api.com

---

**文档结束**
