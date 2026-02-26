# Error Code System Documentation

This directory contains comprehensive documentation for the New API error code system.

## 📚 Documents

### 1. [Error Code Improvements Proposal](./error-code-improvements.md)
**📋 Design proposal and architecture overview**

This document details the proposed improvements to the error code system, including:
- Analysis of current state (strengths and weaknesses)
- Detailed improvement proposals
- Implementation strategy and migration plan
- Priority matrix for improvements

**Best for**: Understanding the "why" behind the changes

---

### 2. [Error Codes Reference](./ERROR_CODES.md)
**📖 Complete error code reference**

Auto-generated reference of all error codes, including:
- All error codes with numeric values
- HTTP status code mappings
- Error level assignments
- Usage examples
- Category organization

**Best for**: Looking up specific error codes and their meanings

**How to regenerate**:
```bash
go run tools/generate_error_doc.go > docs/ERROR_CODES.md
```

---

### 3. [Migration Guide](./error-code-migration-guide.md)
**🔄 Step-by-step migration instructions**

Practical guide for migrating from the old string-based error codes to the new numeric system:
- Before/after code examples
- Migration steps
- Common issues and solutions
- Testing strategies

**Best for**: Developers updating existing code

---

## 🏗️ Architecture Overview

The new error code system consists of several components:

### Core Files

| File | Description |
|------|-------------|
| `types/error.go` | Core error implementation with `NewAPIError` struct |
| `types/error_code.go` | Numeric error code definitions and HTTP mapping |
| `types/error_level.go` | Error severity level definitions |
| `types/error_i18n.go` | Internationalization support for error messages |

### Tools

| File | Description |
|------|-------------|
| `tools/generate_error_doc.go` | Auto-generates ERROR_CODES.md documentation |

---

## 🎯 Quick Reference

### Error Code Categories

| Category | Range | HTTP Status | Example |
|----------|-------|-------------|---------|
| General Errors | 1xxx | 400 | Invalid Request (1001) |
| System Errors | 2xxx | 500 | Count Token Failed (2001) |
| Channel Errors | 3xxx | 503 | Channel Not Available (3008) |
| Client Errors | 4xxx | 401 | Access Denied (4003) |
| Upstream Errors | 5xxx | 502 | Bad Response (5003) |
| Database Errors | 6xxx | 500 | Query Data Error (6001) |
| Quota Errors | 7xxx | 402 | Insufficient Quota (7001) |

### Error Levels

| Level | Description | Color |
|-------|-------------|-------|
| info | Informational messages | Cyan |
| warning | Non-critical warnings | Yellow |
| error | Error events | Red |
| critical | Critical failures | Magenta |

### Supported Languages

| Code | Language |
|------|----------|
| en | English |
| zh | 中文 (Chinese) |
| ja | 日本語 (Japanese) |
| fr | Français (French) |
| ru | Русский (Russian) |
| vi | Tiếng Việt (Vietnamese) |

---

## 💻 Usage Examples

### Creating an Error

```go
import "github.com/QuantumNous/new-api/types"

// Auto-maps HTTP status code and error level
err := types.NewError(
    errors.New("channel not available"),
    types.ErrorCodeChannelNoAvailableKey,
)

// Access error properties
fmt.Println(err.StatusCode)         // 503
fmt.Println(err.Level)              // error
fmt.Println(err.errorCode.String()) // "channel_no_available_key"
```

### Localized Error Messages

```go
// Get localized error message
lang := types.GetLanguageFromContext("zh-CN")
message := err.Localize(lang)
// Returns: "渠道不可用"
```

### Custom Error Level

```go
// Override default error level
err := types.NewError(
    errors.New("custom error"),
    types.ErrorCodeInvalidRequest,
    types.ErrOptionWithLevel(types.ErrorLevelCritical),
)
```

### Error Options

```go
err := types.NewError(
    errors.New("something went wrong"),
    types.ErrorCodeInvalidRequest,
    types.ErrOptionWithSkipRetry(),           // Skip retry
    types.ErrOptionWithNoRecordErrorLog(),    // Don't log
    types.ErrOptionWithHideErrMsg("Hidden"),  // Hide message
)
```

---

## 🔄 Migration Summary

### Key Changes

1. **ErrorCode Type**: Changed from `string` to `int`
2. **HTTP Status Auto-mapping**: No need to manually set status codes
3. **Error Levels**: Built-in severity levels
4. **Internationalization**: Support for 6 languages
5. **Better Performance**: Integer comparison vs string comparison

### Breaking Changes

- `ErrorCode` is now `int` instead of `string`
- Use `.String()` method for string representation
- Old constants renamed to `LegacyErrorCodeString` (backward compatible)

### Benefits

✅ Type safety (compile-time checking)
✅ Faster error comparison
✅ Automatic HTTP status mapping
✅ Built-in error levels for monitoring
✅ Easy internationalization
✅ Better developer experience

---

## 📊 Statistics

- **Total Error Codes**: 34
- **Categories**: 7 (1xxx-7xxx)
- **Languages Supported**: 6
- **Documentation Files**: 3
- **Code Files**: 4 (types/ + tools/)

---

## 🛠️ Maintenance

### Adding a New Error Code

1. Add to `types/error_code.go`:
   ```go
   ErrorCodeMyNewError ErrorCode = 1009
   ```

2. Add string mapping:
   ```go
   errorCodeStrings[ErrorCodeMyNewError] = "my_new_error"
   ```

3. Add HTTP status mapping:
   ```go
   errorCodeHTTPStatusMap[ErrorCodeMyNewError] = http.StatusBadRequest
   ```

4. Add level mapping:
   ```go
   errorCodeLevelMap[ErrorCodeMyNewError] = ErrorLevelWarning
   ```

5. Add to `tools/generate_error_doc.go`:
   ```go
   {1009, "ErrorCodeMyNewError", "General Errors (1xxx)", 400, "warning", "My new error description"},
   ```

6. Add localized messages to `types/error_i18n.go`

7. Regenerate documentation:
   ```bash
   go run tools/generate_error_doc.go > docs/ERROR_CODES.md
   ```

---

## 📞 Support

For questions, issues, or suggestions:

1. Check the [Migration Guide](./error-code-migration-guide.md) for common issues
2. Review the [Improvements Proposal](./error-code-improvements.md) for design details
3. Search the [Error Codes Reference](./ERROR_CODES.md) for specific codes
4. Open a GitHub issue for bugs or feature requests

---

**Last Updated**: 2026-02-26
**Version**: 1.0
**Maintainer**: New API Team
