# 公开数据 API 文档

**版本**: v1.0
**最后更新**: 2026-01-23
**基础URL**: `https://your-domain.com`

---

## 文档概览

本文档定义了 New API 平台的公开数据接口规范。这些接口无需认证，用于公开数据看板展示。

### 快速导航

- [接口列表](#接口列表)
- [通用说明](#通用说明)
- [接口详情](#接口详情)
  - [获取公开数据看板图表数据](#1-获取公开数据看板图表数据)
  - [获取公开统计数据](#2-获取公开统计数据)
- [错误处理](#错误处理)
- [代码示例](#代码示例)
- [常见问题](#常见问题)
- [版本历史](#版本历史)

---

## 接口列表

| 接口名称 | 路径 | 方法 | 认证 | 说明 |
|---------|------|------|------|------|
| 获取图表数据 | `/api/data/public` | GET | ❌ 无需 | 获取时间序列数据用于绘制图表 |
| 获取统计数据 | `/api/data/public/stats` | GET | ❌ 无需 | 获取平台统计指标和Top用户 |

---

## 通用说明

### 请求规范

#### 基础信息
- **协议**: HTTPS（生产环境）/ HTTP（开发环境）
- **Content-Type**: `application/json`
- **字符编码**: UTF-8
- **认证方式**: 公开接口，无需认证

#### 速率限制
- **限制**: 100 请求/分钟/IP
- **超限响应**: HTTP 429 Too Many Requests
- **限制头信息**:
  ```
  X-RateLimit-Limit: 100
  X-RateLimit-Remaining: 95
  X-RateLimit-Reset: 1706054460
  ```

#### 时间说明
- 所有时间戳均为 **Unix 秒级时间戳**（10位数字）
- 时区: **UTC+0**（协调世界时）
- 示例: `1706054400` = `2024-01-24 10:00:00 UTC`

### 响应规范

#### 统一响应格式

所有接口均返回以下格式：

```json
{
  "success": boolean,   // 业务是否成功
  "message": string,    // 错误信息（成功时为空字符串）
  "data": any          // 响应数据（失败时可能为 null）
}
```

#### HTTP 状态码

| 状态码 | 说明 | 处理建议 |
|-------|------|---------|
| 200 | 请求成功 | 检查 `success` 字段判断业务成功/失败 |
| 429 | 请求过于频繁 | 降低请求频率，等待 `X-RateLimit-Reset` 时间后重试 |
| 500 | 服务器内部错误 | 稍后重试，持续失败请联系技术支持 |
| 503 | 服务暂时不可用 | 系统维护中，稍后重试 |

#### 响应头

```
Content-Type: application/json; charset=utf-8
Cache-Control: public, max-age=300
X-Request-ID: 550e8400-e29b-41d4-a716-446655440000
```

---

## 接口详情

### 1. 获取公开数据看板图表数据

获取按模型和时间分组的聚合数据，用于绘制消费趋势图、模型占比图等。

#### 基本信息

```
GET /api/data/public
```

| 项目 | 说明 |
|------|------|
| 接口名称 | 获取公开数据看板图表数据 |
| 接口地址 | `/api/data/public` |
| 请求方式 | `GET` |
| 认证方式 | 无需认证 |
| 接口版本 | v1.0 |
| 缓存时间 | 5分钟 |

#### 请求参数

##### Query 参数

| 参数名 | 类型 | 必填 | 默认值 | 说明 | 示例 |
|--------|------|------|--------|------|------|
| start_timestamp | int64 | 否 | 当前时间-7天 | 查询开始时间（Unix秒级时间戳） | 1706054400 |
| end_timestamp | int64 | 否 | 当前时间 | 查询结束时间（Unix秒级时间戳） | 1706658000 |

##### 参数约束

- ⚠️ `start_timestamp` 和 `end_timestamp` 需**同时提供**或**同时省略**
- ⚠️ 时间范围不能超过 **31天**（2,678,400秒）
- ⚠️ `start_timestamp` 必须小于 `end_timestamp`
- ℹ️ 时间戳为 UTC+0 时区

##### 请求示例

```bash
# 默认查询（最近7天）
GET /api/data/public

# 指定时间范围
GET /api/data/public?start_timestamp=1706054400&end_timestamp=1706658000
```

#### 响应参数

##### 响应数据结构

| 字段名 | 类型 | 说明 |
|--------|------|------|
| success | boolean | 请求是否成功 |
| message | string | 错误信息，成功时为空字符串 |
| data | array | 数据记录数组 |

##### data 数组元素字段

| 字段名 | 类型 | 说明 | 示例 |
|--------|------|------|------|
| model_name | string | 模型名称 | `gpt-4` |
| count | int | 该时间段内的总请求次数 | 1523 |
| quota | int | 该时间段内的总消耗额度 | 152300 |
| token_used | int | 该时间段内的总Token消耗 | 98500 |
| created_at | int64 | 时间戳（Unix秒级，精确到小时） | 1706054400 |

##### 数据聚合规则

- **分组维度**: `model_name` × `created_at`（小时粒度）
- **聚合方式**: SUM（`count`、`quota`、`token_used`）
- **时间对齐**: 时间戳已对齐到小时边界（分钟和秒为00）
  - 示例: `1706054400` = `2024-01-24 10:00:00 UTC`
  - 说明: 10:00-10:59 的所有请求聚合到 `10:00:00` 时间戳下

#### 响应示例

##### 成功响应 (200 OK)

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

##### 错误响应示例

**时间范围超限 (200 OK)**
```json
{
  "success": false,
  "message": "时间跨度不能超过31天"
}
```

**参数缺失 (200 OK)**
```json
{
  "success": false,
  "message": "start_timestamp 和 end_timestamp 必须同时提供"
}
```

**无效时间戳 (200 OK)**
```json
{
  "success": false,
  "message": "start_timestamp 必须小于 end_timestamp"
}
```

**服务器错误 (500 Internal Server Error)**
```json
{
  "success": false,
  "message": "数据库查询失败: connection timeout"
}
```

---

### 2. 获取公开统计数据

获取平台统计指标和消费Top用户排行，用于数据看板的统计卡片展示。

#### 基本信息

```
GET /api/data/public/stats
```

| 项目 | 说明 |
|------|------|
| 接口名称 | 获取公开统计数据 |
| 接口地址 | `/api/data/public/stats` |
| 请求方式 | `GET` |
| 认证方式 | 无需认证 |
| 接口版本 | v1.0 |
| 缓存时间 | 1分钟 |

#### 请求参数

##### Query 参数

| 参数名 | 类型 | 必填 | 默认值 | 说明 | 示例 |
|--------|------|------|--------|------|------|
| start_timestamp | int64 | 否 | 当前时间-24小时 | Top用户查询开始时间 | 1706054400 |
| end_timestamp | int64 | 否 | 当前时间 | Top用户查询结束时间 | 1706140800 |

##### 参数约束

- ⚠️ `start_timestamp` 和 `end_timestamp` 需**同时提供**或**同时省略**
- ⚠️ 时间范围不能超过 **31天**（2,678,400秒）
- ℹ️ 时间范围**仅影响 Top 用户排行**，不影响基础统计指标
- ℹ️ 基础统计指标为实时数据，不受时间参数影响

##### 请求示例

```bash
# 默认查询（最近24小时的Top用户）
GET /api/data/public/stats

# 指定时间范围
GET /api/data/public/stats?start_timestamp=1706054400&end_timestamp=1706140800
```

#### 响应参数

##### 响应数据结构

| 字段名 | 类型 | 说明 |
|--------|------|------|
| success | boolean | 请求是否成功 |
| message | string | 错误信息，成功时为空字符串 |
| data | object | 响应数据对象 |
| data.stats | object | 平台统计数据（实时） |
| data.top_users | array | Top用户列表（受时间参数影响） |

##### stats 对象字段

| 字段名 | 类型 | 说明 | 示例 |
|--------|------|------|------|
| enabled_models_count | int64 | 当前启用的模型数量 | 127 |
| enabled_channels_count | int64 | 当前启用的渠道（服务商）数量 | 15 |
| active_tokens_count | int64 | 当前有效的令牌数量 | 342 |
| today_token_usage | int64 | 今日Token消耗量（自然日0:00-23:59，UTC+8） | 1250000 |
| total_req_count | int64 | 总请求数（从quota_data表汇总） | 150000 |
| total_quota | int64 | 总消耗额度（从quota_data表汇总） | 50000000 |
| total_token_usage | int64 | 总Token消耗（从quota_data表汇总） | 75000000 |
| total_data_count | int64 | 总数据记录数（quota_data表记录总数） | 25000 |

##### top_users 数组元素字段

| 字段名 | 类型 | 说明 | 示例 |
|--------|------|------|------|
| username | string | 匿名化处理后的用户名 | `joh***oe` |
| quota | int64 | 指定时间范围内的总消耗额度 | 150000 |
| token_used | int64 | 指定时间范围内的总Token消耗 | 980000 |
| request_count | int64 | 指定时间范围内的总请求次数 | 452 |

##### 用户名匿名化规则

为保护用户隐私，用户名按以下规则匿名化：

| 原始长度 | 匿名化规则 | 示例 |
|---------|-----------|------|
| 0（空） | 显示为 "匿名用户" | `` → `匿名用户` |
| 1-2 | 保留首字符 + `***` | `ab` → `a***` |
| 3-4 | 保留首尾字符 + `***` | `alice` → `a***e` |
| ≥ 5 | 保留前3字符 + `***` + 保留后2字符 | `john_doe` → `joh***oe` |

#### 响应示例

##### 成功响应 (200 OK)

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

##### 错误响应示例

**时间范围超限 (200 OK)**
```json
{
  "success": false,
  "message": "时间跨度不能超过31天"
}
```

**数据库错误 (500 Internal Server Error)**
```json
{
  "success": false,
  "message": "数据库查询失败: connection timeout"
}
```

---

## 错误处理

### 错误响应格式

所有错误响应遵循统一格式：

```json
{
  "success": false,
  "message": "具体错误信息"
}
```

### 常见错误码

| HTTP状态码 | 错误消息 | 原因 | 解决方案 |
|-----------|---------|------|---------|
| 200 | `时间跨度不能超过31天` | 请求的时间范围超过31天 | 缩短查询时间范围 |
| 200 | `start_timestamp 和 end_timestamp 必须同时提供` | 只提供了其中一个时间参数 | 同时提供两个参数或都不提供 |
| 200 | `start_timestamp 必须小于 end_timestamp` | 开始时间大于结束时间 | 检查时间戳顺序 |
| 200 | `数据库查询失败: ...` | 数据库连接或查询异常 | 稍后重试，持续失败请联系技术支持 |
| 429 | `Too Many Requests` | 请求频率超过限制（100次/分钟） | 降低请求频率，等待限额重置 |
| 500 | `Internal Server Error` | 服务器内部错误 | 稍后重试，持续失败请联系技术支持 |
| 503 | `Service Unavailable` | 服务暂时不可用 | 系统维护中，稍后重试 |

### 错误处理最佳实践

1. **检查 HTTP 状态码**
   ```javascript
   if (response.status !== 200) {
     // 处理 HTTP 级别错误（429, 500, 503）
   }
   ```

2. **检查业务状态**
   ```javascript
   if (!response.data.success) {
     // 处理业务级别错误
     console.error(response.data.message);
   }
   ```

3. **实现重试逻辑**
   ```javascript
   // 对于 429 和 500 错误，建议使用指数退避重试
   async function fetchWithRetry(url, maxRetries = 3) {
     for (let i = 0; i < maxRetries; i++) {
       try {
         const response = await fetch(url);
         if (response.status === 429) {
           const retryAfter = response.headers.get('X-RateLimit-Reset');
           await sleep(calculateBackoff(i, retryAfter));
           continue;
         }
         return response;
       } catch (error) {
         if (i === maxRetries - 1) throw error;
         await sleep(calculateBackoff(i));
       }
     }
   }
   ```

---

## 代码示例

### JavaScript (Fetch API)

```javascript
// 获取图表数据
async function getChartData(startTimestamp, endTimestamp) {
  const params = new URLSearchParams();
  if (startTimestamp && endTimestamp) {
    params.append('start_timestamp', startTimestamp);
    params.append('end_timestamp', endTimestamp);
  }

  try {
    const response = await fetch(
      `https://your-domain.com/api/data/public?${params}`,
      {
        method: 'GET',
        headers: {
          'Content-Type': 'application/json'
        }
      }
    );

    const result = await response.json();

    if (!result.success) {
      throw new Error(result.message);
    }

    return result.data;
  } catch (error) {
    console.error('获取图表数据失败:', error);
    throw error;
  }
}

// 获取统计数据
async function getStats(startTimestamp, endTimestamp) {
  const params = new URLSearchParams();
  if (startTimestamp && endTimestamp) {
    params.append('start_timestamp', startTimestamp);
    params.append('end_timestamp', endTimestamp);
  }

  try {
    const response = await fetch(
      `https://your-domain.com/api/data/public/stats?${params}`,
      {
        method: 'GET',
        headers: {
          'Content-Type': 'application/json'
        }
      }
    );

    const result = await response.json();

    if (!result.success) {
      throw new Error(result.message);
    }

    return result.data;
  } catch (error) {
    console.error('获取统计数据失败:', error);
    throw error;
  }
}

// 使用示例
const sevenDaysAgo = Math.floor(Date.now() / 1000) - (7 * 24 * 60 * 60);
const now = Math.floor(Date.now() / 1000);

// 获取最近7天的图表数据
getChartData(sevenDaysAgo, now).then(data => {
  console.log('图表数据:', data);
});

// 获取最近24小时的统计数据
getStats().then(data => {
  console.log('统计数据:', data);
});
```

### Python (requests)

```python
import requests
import time
from datetime import datetime, timedelta

BASE_URL = "https://your-domain.com"

def get_chart_data(start_timestamp=None, end_timestamp=None):
    """获取图表数据"""
    url = f"{BASE_URL}/api/data/public"
    params = {}

    if start_timestamp and end_timestamp:
        params['start_timestamp'] = start_timestamp
        params['end_timestamp'] = end_timestamp

    try:
        response = requests.get(url, params=params, timeout=10)
        response.raise_for_status()

        result = response.json()

        if not result['success']:
            raise Exception(result['message'])

        return result['data']
    except requests.exceptions.RequestException as e:
        print(f"获取图表数据失败: {e}")
        raise

def get_stats(start_timestamp=None, end_timestamp=None):
    """获取统计数据"""
    url = f"{BASE_URL}/api/data/public/stats"
    params = {}

    if start_timestamp and end_timestamp:
        params['start_timestamp'] = start_timestamp
        params['end_timestamp'] = end_timestamp

    try:
        response = requests.get(url, params=params, timeout=10)
        response.raise_for_status()

        result = response.json()

        if not result['success']:
            raise Exception(result['message'])

        return result['data']
    except requests.exceptions.RequestException as e:
        print(f"获取统计数据失败: {e}")
        raise

# 使用示例
if __name__ == "__main__":
    # 计算时间戳（最近7天）
    now = int(time.time())
    seven_days_ago = now - (7 * 24 * 60 * 60)

    # 获取图表数据
    chart_data = get_chart_data(seven_days_ago, now)
    print(f"图表数据: {len(chart_data)} 条记录")

    # 获取统计数据
    stats = get_stats()
    print(f"启用模型数: {stats['stats']['enabled_models_count']}")
    print(f"Top用户数: {len(stats['top_users'])}")
```

### cURL

```bash
# 获取图表数据（默认最近7天）
curl -X GET "https://your-domain.com/api/data/public" \
  -H "Content-Type: application/json"

# 获取图表数据（指定时间范围）
curl -X GET "https://your-domain.com/api/data/public?start_timestamp=1706054400&end_timestamp=1706658000" \
  -H "Content-Type: application/json"

# 获取统计数据（默认最近24小时）
curl -X GET "https://your-domain.com/api/data/public/stats" \
  -H "Content-Type: application/json"

# 获取统计数据（指定时间范围）
curl -X GET "https://your-domain.com/api/data/public/stats?start_timestamp=1706054400&end_timestamp=1706140800" \
  -H "Content-Type: application/json"

# 格式化输出（使用 jq）
curl -s "https://your-domain.com/api/data/public/stats" | jq '.'
```

### Go

```go
package main

import (
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "net/url"
    "time"
)

const BaseURL = "https://your-domain.com"

type Response struct {
    Success bool        `json:"success"`
    Message string      `json:"message"`
    Data    interface{} `json:"data"`
}

type ChartDataItem struct {
    ModelName string `json:"model_name"`
    Count     int    `json:"count"`
    Quota     int    `json:"quota"`
    TokenUsed int    `json:"token_used"`
    CreatedAt int64  `json:"created_at"`
}

// GetChartData 获取图表数据
func GetChartData(startTimestamp, endTimestamp int64) ([]ChartDataItem, error) {
    apiURL := BaseURL + "/api/data/public"

    // 构建查询参数
    params := url.Values{}
    if startTimestamp > 0 && endTimestamp > 0 {
        params.Add("start_timestamp", fmt.Sprintf("%d", startTimestamp))
        params.Add("end_timestamp", fmt.Sprintf("%d", endTimestamp))
    }

    if len(params) > 0 {
        apiURL += "?" + params.Encode()
    }

    // 发送请求
    resp, err := http.Get(apiURL)
    if err != nil {
        return nil, fmt.Errorf("请求失败: %w", err)
    }
    defer resp.Body.Close()

    // 读取响应
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return nil, fmt.Errorf("读取响应失败: %w", err)
    }

    // 解析响应
    var result Response
    if err := json.Unmarshal(body, &result); err != nil {
        return nil, fmt.Errorf("解析JSON失败: %w", err)
    }

    if !result.Success {
        return nil, fmt.Errorf("业务错误: %s", result.Message)
    }

    // 转换数据
    dataBytes, _ := json.Marshal(result.Data)
    var chartData []ChartDataItem
    if err := json.Unmarshal(dataBytes, &chartData); err != nil {
        return nil, fmt.Errorf("转换数据失败: %w", err)
    }

    return chartData, nil
}

func main() {
    // 计算时间戳（最近7天）
    now := time.Now().Unix()
    sevenDaysAgo := now - (7 * 24 * 60 * 60)

    // 获取图表数据
    data, err := GetChartData(sevenDaysAgo, now)
    if err != nil {
        fmt.Printf("获取数据失败: %v\n", err)
        return
    }

    fmt.Printf("获取到 %d 条数据记录\n", len(data))
    for _, item := range data {
        fmt.Printf("模型: %s, 请求数: %d, 额度: %d\n",
            item.ModelName, item.Count, item.Quota)
    }
}
```

---

## 常见问题

### Q1: 时间戳应该使用什么时区？
**A**: 所有时间戳均为 **UTC+0**（协调世界时）。如果您的系统使用其他时区，请在转换时注意时区差异。

### Q2: 为什么返回的数据时间戳都是整点（10:00:00）？
**A**: 数据按小时粒度聚合，时间戳已对齐到小时边界。例如，10:00-10:59的所有请求都聚合到 `10:00:00` 时间戳下。

### Q3: 如何获取最近30天的数据？
**A**: 计算30天前的时间戳并传入参数：
```javascript
const thirtyDaysAgo = Math.floor(Date.now() / 1000) - (30 * 24 * 60 * 60);
const now = Math.floor(Date.now() / 1000);
getChartData(thirtyDaysAgo, now);
```

### Q4: 速率限制是针对每个接口还是所有接口？
**A**: 速率限制（100次/分钟）是**针对单个IP地址**的所有公开接口的总和。

### Q5: `today_token_usage` 按什么时区统计？
**A**: 按 **UTC+8**（北京时间）的自然日统计，即每天 00:00:00 - 23:59:59。

### Q6: Top用户列表最多返回多少条？
**A**: 默认返回 **Top 10** 用户。

### Q7: 数据多久更新一次？
**A**:
- 图表数据：**5分钟缓存**
- 统计数据：**1分钟缓存**

### Q8: 如果没有数据会返回什么？
**A**: 返回空数组或空对象：
```json
{
  "success": true,
  "message": "",
  "data": []  // 或 {"stats": {...}, "top_users": []}
}
```

### Q9: 是否支持CORS跨域请求？
**A**: 是的，公开接口支持CORS，可以从浏览器直接调用。

### Q10: 如何处理429错误（请求过于频繁）？
**A**: 实现指数退避重试机制，读取 `X-RateLimit-Reset` 响应头，等待限额重置后再重试。

---

## 性能优化建议

### 1. 使用缓存
接口响应已设置缓存头，建议在客户端也实现缓存：
```javascript
// 使用本地缓存（5分钟）
const CACHE_TTL = 5 * 60 * 1000;
const cache = new Map();

async function getCachedData(url) {
  const cached = cache.get(url);
  if (cached && Date.now() - cached.timestamp < CACHE_TTL) {
    return cached.data;
  }

  const data = await fetch(url).then(r => r.json());
  cache.set(url, { data, timestamp: Date.now() });
  return data;
}
```

### 2. 批量查询
如果需要多个时间段的数据，考虑合并为一次请求（扩大时间范围），然后在客户端分组。

### 3. 减少请求频率
- 图表数据缓存5分钟，不要频繁刷新
- 统计数据缓存1分钟，不要实时轮询
- 使用WebSocket或Server-Sent Events（如果支持）

### 4. 并发控制
如果需要请求多个接口，使用并发控制避免触发速率限制：
```javascript
// 使用 Promise.all 并发请求（但不超过速率限制）
const results = await Promise.all([
  getChartData(start1, end1),
  getStats(start2, end2)
]);
```

---

## 最佳实践

### 1. 错误处理
始终检查 `success` 字段并处理错误：
```javascript
const result = await fetch('/api/data/public').then(r => r.json());
if (!result.success) {
  console.error('API错误:', result.message);
  // 显示用户友好的错误提示
  return;
}
```

### 2. 时间范围验证
在客户端验证时间范围，避免不必要的请求：
```javascript
function validateTimeRange(start, end) {
  const maxDays = 31 * 24 * 60 * 60;
  if (end - start > maxDays) {
    throw new Error('时间范围不能超过31天');
  }
  if (start >= end) {
    throw new Error('开始时间必须小于结束时间');
  }
}
```

### 3. 数据可视化
使用图表库（如 ECharts、Chart.js）展示数据：
```javascript
// ECharts 示例
const chartData = await getChartData();
const modelNames = [...new Set(chartData.map(d => d.model_name))];
const times = [...new Set(chartData.map(d => d.created_at))];

const series = modelNames.map(model => ({
  name: model,
  type: 'line',
  data: times.map(time => {
    const item = chartData.find(d =>
      d.model_name === model && d.created_at === time
    );
    return item ? item.quota : 0;
  })
}));
```

### 4. 响应式设计
对于移动端，考虑减少数据粒度或时间范围：
```javascript
const isMobile = window.innerWidth < 768;
const days = isMobile ? 3 : 7;  // 移动端只显示3天
```

---

## 安全性说明

### 1. HTTPS 强制
生产环境必须使用 HTTPS 协议，保护数据传输安全。

### 2. 速率限制
接口实施了速率限制（100次/分钟/IP），防止滥用和DDoS攻击。

### 3. 数据匿名化
用户名已按规则匿名化，保护用户隐私。

### 4. 输入验证
所有参数在服务端进行严格验证，防止SQL注入等攻击。

---

## OpenAPI 规范

可以使用以下工具生成 OpenAPI/Swagger 规范：
- [Swagger Editor](https://editor.swagger.io/)
- [Postman](https://www.postman.com/) - 导入API后可导出OpenAPI规范

**OpenAPI 3.0 规范文件**: [查看 openapi.yaml](#)

---

## 版本历史

| 版本 | 日期 | 变更内容 | 变更人 |
|------|------|---------|--------|
| v1.0 | 2026-01-22 | 初始版本，定义两个公开数据接口 | Development Team |
| v1.1 | 2026-01-23 | 增强文档：添加代码示例、FAQ、性能优化建议 | Documentation Team |

---

## 技术支持

### 联系方式

- **GitHub Issues**: https://github.com/QuantumNous/new-api/issues
- **技术支持邮箱**: support@quantumnous.com
- **社区论坛**: https://community.new-api.com

### 相关资源

- **项目主页**: https://github.com/QuantumNous/new-api
- **完整文档**: https://docs.new-api.com
- **API测试工具**: https://api.new-api.com/playground
- **状态页面**: https://status.new-api.com

### 响应时间

- **工作日**: 24小时内响应
- **周末/节假日**: 48小时内响应
- **紧急问题**: 通过邮件标注 [URGENT] 优先处理

---

## 附录

### 附录A: Postman Collection

可以导入以下 Postman Collection 快速测试接口：

```json
{
  "info": {
    "name": "New API - Public Data APIs",
    "schema": "https://schema.getpostman.com/json/collection/v2.1.0/collection.json"
  },
  "item": [
    {
      "name": "Get Chart Data",
      "request": {
        "method": "GET",
        "header": [],
        "url": {
          "raw": "{{baseUrl}}/api/data/public?start_timestamp={{startTime}}&end_timestamp={{endTime}}",
          "host": ["{{baseUrl}}"],
          "path": ["api", "data", "public"],
          "query": [
            {"key": "start_timestamp", "value": "{{startTime}}"},
            {"key": "end_timestamp", "value": "{{endTime}}"}
          ]
        }
      }
    },
    {
      "name": "Get Stats",
      "request": {
        "method": "GET",
        "header": [],
        "url": {
          "raw": "{{baseUrl}}/api/data/public/stats",
          "host": ["{{baseUrl}}"],
          "path": ["api", "data", "public", "stats"]
        }
      }
    }
  ]
}
```

### 附录B: 数据格式示例

#### 完整响应示例（图表数据）

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
      "model_name": "gpt-4",
      "count": 1687,
      "quota": 168700,
      "token_used": 105000,
      "created_at": 1706058000
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

---

**文档结束** | New API Platform © 2026
