# Error Code Improvement Proposal

> **Document Version**: 1.0
> **Created**: 2026-02-26
> **Status**: Proposed

## 📊 Current State Analysis

### ✅ Strengths

1. **Centralized Management**: Error codes are centrally defined in `types/error.go`
2. **Unified Encapsulation**: `NewAPIError` structure provides unified error wrapping
3. **Multi-format Support**: Supports error format conversion for multiple AI services (OpenAI, Claude, Gemini, etc.)
4. **Sensitive Data Masking**: Automatic masking of URLs, IPs, and other sensitive information
5. **Flexible Options Pattern**: Supports skip retry, hide error messages, etc.
6. **Clear Categorization**: Errors categorized by functional modules (channel, quota, client, response)
7. **Request Tracing**: Supports request ID tracking for debugging

### ❌ Problems Identified

#### 1. Missing Numeric Error Code System
```go
// Current: Only string error codes, inefficient for quick identification and log search
ErrorCode = "invalid_request"

// Should have: Numeric code for efficient processing
ErrorCode = 1001  // with string representation "invalid_request"
```

#### 2. Inconsistent HTTP Status Code Mapping
HTTP status codes are scattered throughout the codebase:
```go
// controller/relay.go
newAPIError = types.NewErrorWithStatusCode(err, types.ErrorCodeInvalidRequest, http.StatusRequestEntityTooLarge)
newAPIError = types.NewError(err, types.ErrorCodeInvalidRequest)
```

No unified mapping rule between error codes and HTTP status codes.

#### 3. Inconsistent Error Code Naming
```go
// Some have prefixes
ErrorCodeChannelNoAvailableKey ErrorCode = "channel:no_available_key"
// Some don't
ErrorCodeInvalidRequest ErrorCode = "invalid_request"
```

#### 4. Missing Error Level Definitions
No distinction between Critical/Error/Warning/Info levels, making monitoring and alerting difficult.

#### 5. Inconsistent Error Handling
```go
// Some places return raw error
func SomeFunc() error {
    return errors.New("xxx")  // ❌ Should return *NewAPIError
}
// Some places return *NewAPIError
func OtherFunc() *types.NewAPIError {
    return types.NewError(...)
}
```

#### 6. Missing Error Code Documentation
No centralized list of error codes; developers must search through code to find available codes.

#### 7. No Backend Internationalization
Error messages are in English only; i18n is entirely frontend-dependent.

---

## 💡 Improvement Proposals

### Proposal 1: Introduce Numeric Error Code System

**File**: `types/error_code.go`

```go
type ErrorCode int

const (
    // General Errors (1xxx)
    ErrorCodeInvalidRequest ErrorCode = 1001
    ErrorCodeSensitiveWordsDetected ErrorCode = 1002

    // System Errors (2xxx)
    ErrorCodeCountTokenFailed ErrorCode = 2001
    ErrorCodeModelPriceError ErrorCode = 2002
    ErrorCodeInvalidApiType ErrorCode = 2003
    ErrorCodeJsonMarshalFailed ErrorCode = 2004
    ErrorCodeDoRequestFailed ErrorCode = 2005
    ErrorCodeGetChannelFailed ErrorCode = 2006
    ErrorCodeGenRelayInfoFailed ErrorCode = 2007

    // Channel Errors (3xxx)
    ErrorCodeChannelNoAvailableKey ErrorCode = 3001
    ErrorCodeChannelParamOverrideInvalid ErrorCode = 3002
    ErrorCodeChannelHeaderOverrideInvalid ErrorCode = 3003
    ErrorCodeChannelModelMappedError ErrorCode = 3004
    ErrorCodeChannelAwsClientError ErrorCode = 3005
    ErrorCodeChannelInvalidKey ErrorCode = 3006
    ErrorCodeChannelResponseTimeExceeded ErrorCode = 3007

    // Client Errors (4xxx)
    ErrorCodeReadRequestBodyFailed ErrorCode = 4001
    ErrorCodeConvertRequestFailed ErrorCode = 4002
    ErrorCodeAccessDenied ErrorCode = 4003
    ErrorCodeBadRequestBody ErrorCode = 4004

    // Upstream Errors (5xxx)
    ErrorCodeReadResponseBodyFailed ErrorCode = 5001
    ErrorCodeBadResponseStatusCode ErrorCode = 5002
    ErrorCodeBadResponse ErrorCode = 5003
    ErrorCodeBadResponseBody ErrorCode = 5004
    ErrorCodeEmptyResponse ErrorCode = 5005
    ErrorCodeAwsInvokeError ErrorCode = 5006
    ErrorCodeModelNotFound ErrorCode = 5007
    ErrorCodePromptBlocked ErrorCode = 5008

    // Database Errors (6xxx)
    ErrorCodeQueryDataError ErrorCode = 6001
    ErrorCodeUpdateDataError ErrorCode = 6002

    // Quota Errors (7xxx)
    ErrorCodeInsufficientUserQuota ErrorCode = 7001
    ErrorCodePreConsumeTokenQuotaFailed ErrorCode = 7002
)

// String returns the string representation of error code
func (c ErrorCode) String() string {
    // Map implementation
}

// HTTPStatusCode returns the corresponding HTTP status code
func (c ErrorCode) HTTPStatusCode() int {
    // Map implementation
}
```

**Benefits**:
- Fast numeric comparison for error handling logic
- Easy to log and search
- Clear categorization by number ranges
- Type safety with numeric constants

---

### Proposal 2: Add Error Level Definitions

**File**: `types/error_level.go`

```go
type ErrorLevel int

const (
    ErrorLevelInfo ErrorLevel = iota
    ErrorLevelWarning
    ErrorLevelError
    ErrorLevelCritical
)

func (l ErrorLevel) String() string {
    switch l {
    case ErrorLevelInfo:
        return "info"
    case ErrorLevelWarning:
        return "warning"
    case ErrorLevelError:
        return "error"
    case ErrorLevelCritical:
        return "critical"
    default:
        return "unknown"
    }
}

func (l ErrorLevel) Color() string {
    switch l {
    case ErrorLevelInfo:
        return "\033[36m" // Cyan
    case ErrorLevelWarning:
        return "\033[33m" // Yellow
    case ErrorLevelError:
        return "\033[31m" // Red
    case ErrorLevelCritical:
        return "\033[35m" // Magenta
    default:
        return "\033[0m" // Reset
    }
}
```

**Integration with NewAPIError**:
```go
type NewAPIError struct {
    // ... existing fields
    Level ErrorLevel
}

// Usage
func (e *NewAPIError) Log() {
    level := e.Level
    logger.Logf(level, e.Error())
}
```

**Benefits**:
- Clear severity indication
- Easy to integrate with monitoring systems (Prometheus, Sentry)
- Supports alert filtering by severity

---

### Proposal 3: Unified HTTP Status Code Mapping

```go
// In error_code.go
func (c ErrorCode) HTTPStatusCode() int {
    switch c {
    // Client errors (4xx)
    case ErrorCodeInvalidRequest,
         ErrorCodeReadRequestBodyFailed,
         ErrorCodeConvertRequestFailed,
         ErrorCodeBadRequestBody:
        return http.StatusBadRequest

    case ErrorCodeAccessDenied:
        return http.StatusUnauthorized

    case ErrorCodeInsufficientUserQuota:
        return http.StatusPaymentRequired

    case ErrorCodeModelNotFound:
        return http.StatusNotFound

    case ErrorCodeChannelResponseTimeExceeded:
        return http.StatusRequestTimeout

    // Server errors (5xx)
    case ErrorCodeCountTokenFailed,
         ErrorCodeModelPriceError,
         ErrorCodeGetChannelFailed,
         ErrorCodeChannelNoAvailableKey,
         ErrorCodeBadResponseStatusCode,
         ErrorCodeBadResponse:
        return http.StatusInternalServerError

    default:
        return http.StatusInternalServerError
    }
}
```

**Usage**:
```go
// Before
newAPIError = types.NewErrorWithStatusCode(err, types.ErrorCodeInvalidRequest, http.StatusBadRequest)

// After: Status code is automatically determined
newAPIError = types.NewError(err, ErrorCodeInvalidRequest)
statusCode := newAPIError.StatusCode  // Auto-mapped
```

---

### Proposal 4: Standardized Error Code Naming Convention

**Category by Numeric Ranges**:
- `1xxx` - General Errors
- `2xxx` - System Errors
- `3xxx` - Channel Errors
- `4xxx` - Client Errors
- `5xxx` - Upstream Errors
- `6xxx` - Database Errors
- `7xxx` - Quota Errors
- `8xxx` - Auth Errors (reserved)
- `9xxx` - Miscellaneous (reserved)

**Naming Rules**:
1. Use PascalCase for constant names
2. No prefixes needed (use numeric ranges instead)
3. Descriptive names indicating the error cause
4. Include module context when ambiguous

**Examples**:
```go
// ✅ Good
ErrorCodeChannelNoAvailableKey ErrorCode = 3001
ErrorCodeInsufficientUserQuota ErrorCode = 7001

// ❌ Bad (string-based with prefix)
ErrorCodeChannelNoAvailableKey ErrorCode = "channel:no_available_key"
```

---

### Proposal 5: Add Error Code Documentation Generator

**File**: `tools/generate_error_doc.go`

```go
// +build ignore

package main

import (
    "fmt"
    "os"
    "reflect"
    "strings"
)

type ErrorDoc struct {
    Code       int
    Name       string
    Message    string
    HTTPStatus int
    Level      string
}

func main() {
    // Scan all ErrorCode constants
    // Generate ERROR_CODES.md documentation
    // Include code, name, description, HTTP status, level
}

// Generate markdown table
func generateMarkdownTable(errors []ErrorDoc) string {
    // Markdown table generation logic
}
```

**Usage**:
```bash
go run tools/generate_error_doc.go > docs/ERROR_CODES.md
```

**Output**: `docs/ERROR_CODES.md` with comprehensive error code reference

---

### Proposal 6: Unified Error Handling Middleware

**File**: `middleware/error_handler.go`

```go
func ErrorHandler() gin.HandlerFunc {
    return func(c *gin.Context) {
        c.Next()

        // Handle errors returned from handlers
        if len(c.Errors) > 0 {
            err := c.Errors.Last().Err

            if newApiErr, ok := err.(*types.NewAPIError); ok {
                // Log error with level
                logger.LogWithLevel(c, newApiErr.Level, newApiErr.Error())

                // Return response based on relay format
                switch c.GetHeader("Relay-Format") {
                case "claude":
                    c.JSON(newApiErr.StatusCode, gin.H{
                        "type":  "error",
                        "error": newApiErr.ToClaudeError(),
                    })
                default:
                    c.JSON(newApiErr.StatusCode, gin.H{
                        "error": newApiErr.ToOpenAIError(),
                    })
                }
            } else {
                // Wrap raw errors
                c.JSON(http.StatusInternalServerError, gin.H{
                    "error": gin.H{
                        "message": "Internal server error",
                        "type":    "internal_error",
                        "code":    2000, // Unknown system error
                    },
                })
            }
        }
    }
}
```

**Benefits**:
- Centralized error handling logic
- Consistent error response format
- Automatic logging and monitoring

---

### Proposal 7: Backend Internationalization Support

**File**: `types/error_i18n.go`

```go
type ErrorMessage map[string]string // lang -> message

var errorMessages = map[ErrorCode]ErrorMessage{
    ErrorCodeInvalidRequest: {
        "en": "Invalid request",
        "zh": "请求无效",
        "ja": "無効なリクエスト",
        "fr": "Requête invalide",
    },
    ErrorCodeInsufficientUserQuota: {
        "en": "Insufficient user quota",
        "zh": "用户配额不足",
        "ja": "ユーザークォータが不足しています",
        "fr": "Quota utilisateur insuffisant",
    },
    // ... more error messages
}

// Localize returns localized error message based on Accept-Language header
func (e *NewAPIError) Localize(lang string) string {
    if msgs, ok := errorMessages[e.errorCode]; ok {
        if msg, ok := msgs[lang]; ok {
            return msg
        }
        return msgs["en"] // fallback to English
    }
    return e.Error()
}

// GetLanguageFromContext extracts language from gin.Context
func GetLanguageFromContext(c *gin.Context) string {
    lang := c.GetHeader("Accept-Language")
    if lang == "" {
        return "en" // default
    }
    // Parse language header (e.g., "zh-CN" -> "zh")
    parts := strings.Split(lang, "-")
    if len(parts) > 0 {
        return parts[0]
    }
    return "en"
}
```

**Usage in Controllers**:
```go
func Relay(c *gin.Context) {
    // ...
    lang := types.GetLanguageFromContext(c)
    errorMsg := newAPIError.Localize(lang)
    // ...
}
```

**Benefits**:
- Consistent error messages across all languages
- Easy to add new languages
- Backend can return localized errors in API responses

---

### Proposal 8: Error Code Registry System

**File**: `types/error_registry.go`

```go
type ErrorInfo struct {
    Code       ErrorCode
    Name       string
    Message    string
    HTTPStatus int
    Level      ErrorLevel
}

var errorRegistry = map[ErrorCode]ErrorInfo{}
var registryMutex sync.RWMutex

// RegisterError registers an error code with metadata
func RegisterError(info ErrorInfo) {
    registryMutex.Lock()
    defer registryMutex.Unlock()
    errorRegistry[info.Code] = info
}

// GetErrorInfo retrieves error metadata by code
func GetErrorInfo(code ErrorCode) (ErrorInfo, bool) {
    registryMutex.RLock()
    defer registryMutex.RUnlock()
    info, ok := errorRegistry[code]
    return info, ok
}

// ListAllErrors returns all registered error codes
func ListAllErrors() []ErrorInfo {
    registryMutex.RLock()
    defer registryMutex.RUnlock()

    errors := make([]ErrorInfo, 0, len(errorRegistry))
    for _, info := range errorRegistry {
        errors = append(errors, info)
    }
    return errors
}

// Initialization
func init() {
    RegisterError(ErrorInfo{
        Code:       ErrorCodeInvalidRequest,
        Name:       "invalid_request",
        Message:    "Invalid request parameters",
        HTTPStatus: http.StatusBadRequest,
        Level:      ErrorLevelWarning,
    })

    RegisterError(ErrorInfo{
        Code:       ErrorCodeInsufficientUserQuota,
        Name:       "insufficient_user_quota",
        Message:    "User quota is insufficient",
        HTTPStatus: http.StatusPaymentRequired,
        Level:      ErrorLevelError,
    })
    // ... register all error codes
}
```

**Benefits**:
- Centralized error metadata management
- Easy to query and document error codes
- Supports runtime error code validation

---

## 📋 Implementation Priority

| Priority | Improvement | Impact | Effort | Dependencies |
|----------|-------------|--------|--------|--------------|
| **P0** | Unified HTTP status code mapping | High | Medium | None |
| **P0** | Add error level definitions | High | Low | None |
| **P1** | Introduce numeric error codes | Medium | High | None |
| **P1** | Unified error handling middleware | Medium | Medium | Error levels |
| **P2** | Generate error code documentation | Low | Low | Numeric codes |
| **P2** | Backend internationalization | Low | Medium | None |
| **P3** | Error code registry system | Low | High | Numeric codes |

---

## 🔄 Migration Strategy

### Phase 1: Foundation (P0)
1. ✅ Add `ErrorLevel` definitions
2. ✅ Implement `HTTPStatusCode()` method
3. ✅ Update `NewAPIError` to include `Level` field

### Phase 2: Core Refactoring (P1)
1. ✅ Migrate to numeric error codes
2. ✅ Update all error creation to use new system
3. ✅ Implement unified error handling middleware

### Phase 3: Enhancement (P2)
1. ✅ Implement internationalization
2. ✅ Generate error code documentation
3. ✅ Add error code registry

### Phase 4: Advanced Features (P3)
1. ⏳ Error code registry with metadata
2. ⏳ Runtime error validation
3. ⏳ Error analytics integration

---

## 📝 Code Examples

### Before Current System

```go
// Creating an error
err := types.NewError(
    errors.New("channel not available"),
    types.ErrorCodeChannelNoAvailableKey,
)

// Using error
if err != nil {
    c.JSON(http.StatusInternalServerError, gin.H{
        "error": err.ToOpenAIError(),
    })
}
```

### After Proposed System

```go
// Creating an error with automatic status code and level
err := types.NewError(
    errors.New("channel not available"),
    types.ErrorCodeChannelNoAvailableKey,
    types.ErrOptionWithLevel(types.ErrorLevelError),
)

// Using error (status code auto-mapped)
if err != nil {
    // Status code automatically determined from error code
    c.JSON(err.StatusCode, gin.H{
        "error": err.ToOpenAIError(),
    })
}

// Localized error message
lang := types.GetLanguageFromContext(c)
errorMsg := err.Localize(lang)
```

---

## 🎯 Expected Outcomes

### Maintainability
- ✅ Clear numeric error codes for fast identification
- ✅ Centralized error metadata management
- ✅ Auto-generated documentation

### Reliability
- ✅ Consistent HTTP status code mapping
- ✅ Error level-based logging and alerting
- ✅ Reduced human error in error handling

### Developer Experience
- ✅ Easy to find and use error codes
- ✅ Clear error categorization
- ✅ Better IDE auto-completion
- ✅ Comprehensive error code documentation

### User Experience
- ✅ Localized error messages
- ✅ Consistent error format across all API endpoints
- ✅ Better error messages for debugging

---

## 🔒 Backward Compatibility

To maintain backward compatibility during migration:

1. **Keep string error code type** as alias
   ```go
   type ErrorCodeString string
   // Map old string codes to new numeric codes
   ```

2. **Dual support** during transition period
   ```go
   func (c ErrorCode) LegacyString() string {
       // Convert numeric to legacy string format
   }
   ```

3. **Deprecation warnings** for old error creation methods
   ```go
   // Deprecated: Use NewError with numeric ErrorCode
   func NewErrorLegacy(err error, code string) *NewAPIError {
       // Implementation with deprecation warning
   }
   ```

4. **Gradual migration**:
   - Phase 1: Add new system alongside old
   - Phase 2: Migrate all new code to use new system
   - Phase 3: Migrate existing code incrementally
   - Phase 4: Remove old system after full migration

---

## 📚 References

- Current Implementation: `types/error.go`
- Error Handling Best Practices: [Go Error Handling](https://go.dev/doc/error-handle)
- HTTP Status Codes: [MDN HTTP Status](https://developer.mozilla.org/en-US/docs/Web/HTTP/Status)
- i18n Patterns: [Go i18n](https://github.com/nicksnyder/go-i18n)

---

## 📞 Contact & Discussion

For questions or suggestions regarding this proposal:
1. Create an issue in the repository
2. Start a discussion in the team communication channel
3. Submit a PR for any improvements to this proposal

---

**Last Updated**: 2026-02-26
**Next Review**: After Phase 1 completion
