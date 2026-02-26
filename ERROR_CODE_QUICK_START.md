# 错误码系统快速开始指南

> **New API 工程错误码系统**
> 更新时间: 2026-02-26

## ✅ 已完成的修改

所有新文件已经成功集成到 new-api 工程中：

### 📁 新增文件

```
new-api/
├── types/
│   ├── error.go              (已更新 - 添加 Level 字段和自动映射)
│   ├── error_code.go         (新增 - 数字错误码系统)
│   ├── error_level.go        (新增 - 错误级别定义)
│   └── error_i18n.go         (新增 - 国际化支持)
├── tools/
│   └── generate_error_doc.go (新增 - 文档生成工具)
└── docs/
    ├── README.md                      (新增)
    ├── ERROR_CODES.md                 (新增)
    ├── error-code-improvements.md     (新增)
    └── error-code-migration-guide.md  (新增)
```

### 🎯 核心功能

| 功能 | 说明 | 状态 |
|------|------|------|
| 数字错误码 | 34个错误码，分为7个类别 (1xxx-7xxx) | ✅ 已完成 |
| HTTP状态码自动映射 | 无需手动设置状态码 | ✅ 已完成 |
| 错误级别 | 4个级别: info/warning/error/critical | ✅ 已完成 |
| 国际化支持 | 6种语言: en/zh/ja/fr/ru/vi | ✅ 已完成 |
| 自动文档生成 | 从代码生成 ERROR_CODES.md | ✅ 已完成 |

## 🚀 快速开始

### 1. 基本用法

```go
package main

import (
    "errors"
    "github.com/Calcium-Ion/new-api/types"
)

// 创建错误 - 自动映射HTTP状态码和错误级别
func example() {
    err := types.NewError(
        errors.New("channel not available"),
        types.ErrorCodeChannelNoAvailableKey,
    )

    // 访问错误属性
    statusCode := err.StatusCode     // 自动: 503
    level := err.Level              // 自动: error
    code := err.errorCode.String()  // "channel_no_available_key"
}
```

### 2. 在 Controller 中使用

```go
package controller

import (
    "github.com/gin-gonic/gin"
    "github.com/Calcium-Ion/new-api/types"
)

func Relay(c *gin.Context) {
    // ... 业务逻辑 ...

    if err != nil {
        newApiErr := types.NewError(err, types.ErrorCodeInvalidRequest)

        // HTTP状态码已自动映射，无需手动设置
        c.JSON(newApiErr.StatusCode, gin.H{
            "error": newApiErr.ToOpenAIError(),
        })
        return
    }
}
```

### 3. 使用本地化消息

```go
func Relay(c *gin.Context) {
    if err != nil {
        newApiErr := types.NewError(err, types.ErrorCodeChannelNoAvailableKey)

        // 获取客户端语言
        acceptLang := c.GetHeader("Accept-Language")
        lang := types.GetLanguageFromContext(acceptLang)

        // 返回本地化错误消息
        c.JSON(newApiErr.StatusCode, gin.H{
            "error": newApiErr.ToOpenAIError(),
            "message": newApiErr.Localize(lang),  // 根据语言返回不同消息
        })
        return
    }
}
```

### 4. 自定义错误级别

```go
// 覆盖默认的错误级别
err := types.NewError(
    errors.New("critical system failure"),
    types.ErrorCodeGetChannelFailed,
    types.ErrOptionWithLevel(types.ErrorLevelCritical),
)
```

### 5. 错误选项组合

```go
err := types.NewError(
    errors.New("operation failed"),
    types.ErrorCodeDoRequestFailed,
    types.ErrOptionWithSkipRetry(),           // 跳过重试
    types.ErrOptionWithNoRecordErrorLog(),    // 不记录日志
    types.ErrOptionWithLevel(types.ErrorLevelWarning),  // 设置级别
)
```

## 📊 错误码快速参考

### 错误码分类

| 分类 | 范围 | HTTP状态 | 示例 |
|------|------|----------|------|
| 通用错误 | 1xxx | 400 | `ErrorCodeInvalidRequest` (1001) |
| 系统错误 | 2xxx | 500 | `ErrorCodeCountTokenFailed` (2001) |
| 渠道错误 | 3xxx | 503 | `ErrorCodeChannelNoAvailableKey` (3001) |
| 客户端错误 | 4xxx | 401 | `ErrorCodeAccessDenied` (4003) |
| 上游错误 | 5xxx | 502 | `ErrorCodeBadResponseStatusCode` (5002) |
| 数据库错误 | 6xxx | 500 | `ErrorCodeQueryDataError` (6001) |
| 配额错误 | 7xxx | 402 | `ErrorCodeInsufficientUserQuota` (7001) |

### 错误级别

| 级别 | 说明 | 颜色 | 使用场景 |
|------|------|------|----------|
| `ErrorLevelInfo` | 信息 | 青色 | 正常操作���息 |
| `ErrorLevelWarning` | 警告 | 黄色 | 非关键问题 |
| `ErrorLevelError` | 错误 | 红色 | 错误事件 |
| `ErrorLevelCritical` | 严重 | 洋红 | 系统故障 |

### 支持的语言

| 代码 | 语言 | 使用示例 |
|------|------|----------|
| `en` | English | `GetLanguageFromContext("en")` |
| `zh` | 中文 | `GetLanguageFromContext("zh-CN")` |
| `ja` | 日本語 | `GetLanguageFromContext("ja")` |
| `fr` | Français | `GetLanguageFromContext("fr")` |
| `ru` | Русский | `GetLanguageFromContext("ru")` |
| `vi` | Tiếng Việt | `GetLanguageFromContext("vi")` |

## 🔄 向后兼容

旧的字符串错误码仍然可用（已重命名为 `LegacyErrorCodeString`）：

```go
// 旧方式（仍然可用，但已弃用）
oldCode := types.LegacyErrorCodeInvalidRequest  // "invalid_request"

// 新方式（推荐）
newCode := types.ErrorCodeInvalidRequest  // 1001

// 转换
code := types.ErrorCodeFromString("invalid_request")  // 1001
str := types.ErrorCodeInvalidRequest.String()         // "invalid_request"
```

## 📚 文档索引

| 文档 | 路径 | 用途 |
|------|------|------|
| 总览 | `docs/README.md` | 系统整体介绍 |
| 错误码参考 | `docs/ERROR_CODES.md` | 完整错误码列表 |
| 迁移指南 | `docs/error-code-migration-guide.md` | 从旧系统迁移 |
| 改进建议 | `docs/error-code-improvements.md` | 设计和架构 |

## 🛠️ 维护命令

### 重新生成文档

```bash
cd /home/lilf/lilf/new-api
go run tools/generate_error_doc.go > docs/ERROR_CODES.md
```

### 编译检查

```bash
# 检查 types 包
cd /home/lilf/lilf/new-api
go build ./types

# 完整编译（需要先安装依赖）
go mod tidy
go build .
```

## 💡 常见使用场景

### 场景1: 处理渠道错误

```go
if err != nil {
    newApiErr := types.NewError(err, types.ErrorCodeChannelNoAvailableKey)

    // 检查是否为渠道错误
    if types.IsChannelError(newApiErr) {
        // 渠道错误 (3xxx范围)
        log.Printf("Channel error: %s", newApiErr.Error())
    }

    c.JSON(newApiErr.StatusCode, gin.H{"error": newApiErr.ToOpenAIError()})
}
```

### 场景2: 跳过重试的错误

```go
// 创建错误并标记为跳过重试
err := types.NewError(
    errors.New("invalid API key"),
    types.ErrorCodeChannelInvalidKey,
    types.ErrOptionWithSkipRetry(),
)

// 检查是否应该跳过重试
if types.IsSkipRetryError(err) {
    // 不重试这个错误
    return err
}
```

### 场景3: 多语言错误响应

```go
func HandleError(c *gin.Context, err error) {
    newApiErr := types.NewError(err, types.ErrorCodeInvalidRequest)

    // 获取客户端语言
    acceptLang := c.GetHeader("Accept-Language")
    lang := types.GetLanguageFromContext(acceptLang)

    // 返回多语言响应
    response := gin.H{
        "error": newApiErr.ToOpenAIError(),
        "code": int(newApiErr.errorCode),
        "level": newApiErr.Level.String(),
        "message": newApiErr.Localize(lang),
    }

    c.JSON(newApiErr.StatusCode, response)
}
```

### 场景4: 配额不足处理

```go
if quota.Insufficient() {
    err := types.NewError(
        errors.New("quota exceeded"),
        types.ErrorCodeInsufficientUserQuota,
    )

    // 自动映射到 402 Payment Required
    c.JSON(err.StatusCode, gin.H{
        "error": err.ToOpenAIError(),
        "quota_required": true,
    })
    return
}
```

## ⚠️ 注意事项

1. **编译通过**: types 包和主要控制器已验证编译通过
2. **向后兼容**: 旧的错误码常量作为 `LegacyErrorCodeString` 仍然可用
3. **渐进式迁移**: 新旧系统可以共存，无需一次性全部修改
4. **自动映射**: HTTP 状态码和错误级别会自动从 ErrorCode 映射

## 🎉 下一步

1. ✅ 查看文档：`docs/ERROR_CODES.md` - 了解所有可用错误码
2. ✅ 阅读迁移指南：`docs/error-code-migration-guide.md` - 学习如何更新代码
3. ✅ 在新代码中使用新的错误码系统
4. ✅ 逐步迁移现有代码（可选）

## 📞 需要帮助？

- 查看完整文档：`docs/` 目录
- 查看错误码列表：`docs/ERROR_CODES.md`
- 阅读迁移指南：`docs/error-code-migration-guide.md`

---

**状态**: ✅ 已集成到 new-api 工程
**版本**: 1.0
**最后更新**: 2026-02-26
