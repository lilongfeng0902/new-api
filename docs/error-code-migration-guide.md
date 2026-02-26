# Error Code Migration Guide

> **Version**: 1.0
> **Last Updated**: 2026-02-26

## Overview

This guide helps you migrate from the legacy string-based error codes to the new numeric error code system.

## What Changed?

### Before (Legacy)

```go
type ErrorCode string

const (
    ErrorCodeInvalidRequest ErrorCode = "invalid_request"
    ErrorCodeChannelNoAvailableKey ErrorCode = "channel:no_available_key"
    // ...
)
```

### After (New)

```go
type ErrorCode int

const (
    ErrorCodeInvalidRequest ErrorCode = 1001
    ErrorCodeChannelNoAvailableKey ErrorCode = 3001
    // ...
)

// Helper methods
func (c ErrorCode) String() string
func (c ErrorCode) HTTPStatusCode() int
func (c ErrorCode) DefaultLevel() ErrorLevel
```

## Benefits of New System

1. **Type Safety**: Numeric codes prevent typos and provide compile-time checking
2. **Performance**: Integer comparison is faster than string comparison
3. **Auto-mapping**: HTTP status codes are automatically determined
4. **Error Levels**: Built-in severity levels for monitoring
5. **Internationalization**: Easy to add localized error messages

## Breaking Changes

### 1. ErrorCode Type Changed

```go
// Old
var code types.ErrorCode = "invalid_request"

// New
var code types.ErrorCode = types.ErrorCodeInvalidRequest  // 1001
```

### 2. String Conversion

```go
// Old
fmt.Sprintf("Error: %s", errorCode)  // "invalid_request"

// New
fmt.Sprintf("Error: %s", errorCode.String())  // "invalid_request"
fmt.Sprintf("Error: %d", errorCode)  // 1001
```

### 3. Legacy Constants Renamed

Old string constants are renamed to `LegacyErrorCodeString`:

```go
// Still available for backward compatibility
types.LegacyErrorCodeInvalidRequest         // "invalid_request"
types.LegacyErrorCodeChannelNoAvailableKey  // "channel:no_available_key"
```

## Migration Steps

### Step 1: Update Error Creation

**Before:**
```go
newApiErr := types.NewError(err, types.ErrorCodeInvalidRequest)
newApiErr.StatusCode = http.StatusBadRequest  // Manual setting
```

**After:**
```go
// HTTP status code is auto-mapped
newApiErr := types.NewError(err, types.ErrorCodeInvalidRequest)
// newApiErr.StatusCode is automatically set to 400
```

### Step 2: Update Error Checking

**Before:**
```go
if err.errorCode == "invalid_request" {
    // handle invalid request
}
```

**After:**
```go
if err.errorCode == types.ErrorCodeInvalidRequest {
    // handle invalid request
}

// Or use helper methods
if types.IsChannelError(err) {
    // handle channel errors
}
```

### Step 3: Update Error Level Usage

**New Feature:**
```go
// Access error level
level := err.Level  // ErrorLevelWarning, ErrorLevelError, etc.

// Log with level
logger.Logf(err.Level, "Error occurred: %s", err.Error())

// Override default level
err := types.NewError(
    errors.New("custom"),
    types.ErrorCodeInvalidRequest,
    types.ErrOptionWithLevel(types.ErrorLevelCritical),
)
```

### Step 4: Use Localized Messages

**New Feature:**
```go
// Get localized error message
lang := types.GetLanguageFromContext(c.GetHeader("Accept-Language"))
message := err.Localize(lang)

// Examples:
// "en" -> "Invalid request parameters"
// "zh" -> "请求参数无效"
// "ja" -> "無効なリクエストパラメータ"
```

## Code Examples

### Controller Error Handling

**Before:**
```go
func Relay(c *gin.Context) {
    if err != nil {
        newApiErr := types.NewError(err, types.ErrorCodeInvalidRequest)
        c.JSON(http.StatusBadRequest, gin.H{
            "error": newApiErr.ToOpenAIError(),
        })
        return
    }
}
```

**After:**
```go
func Relay(c *gin.Context) {
    if err != nil {
        newApiErr := types.NewError(err, types.ErrorCodeInvalidRequest)
        // StatusCode is auto-mapped from error code
        c.JSON(newApiErr.StatusCode, gin.H{
            "error": newApiErr.ToOpenAIError(),
        })
        return
    }
}
```

### Error Comparison

**Before:**
```go
if newApiErr.errorCode == "channel:no_available_key" {
    // handle error
}
```

**After:**
```go
if newApiErr.errorCode == types.ErrorCodeChannelNoAvailableKey {
    // handle error
}

// Or use helper function
if types.IsChannelError(newApiErr) {
    // handle channel errors (3xxx range)
}

if types.IsSkipRetryError(newApiErr) {
    // skip retry logic
}
```

### Custom Error Options

**New Feature:**
```go
err := types.NewError(
    errors.New("something went wrong"),
    types.ErrorCodeInvalidRequest,
    types.ErrOptionWithSkipRetry(),           // Skip retry
    types.ErrOptionWithNoRecordErrorLog(),    // Don't log
    types.ErrOptionWithLevel(types.ErrorLevelWarning),  // Custom level
    types.ErrOptionWithHideErrMsg("Hidden"),  // Hide message
)
```

## Error Code Reference

### Numeric Ranges

| Range | Category | Example |
|-------|----------|---------|
| 1xxx | General Errors | 1001: Invalid Request |
| 2xxx | System Errors | 2001: Count Token Failed |
| 3xxx | Channel Errors | 3001: Channel No Available Key |
| 4xxx | Client Errors | 4001: Read Request Body Failed |
| 5xxx | Upstream Errors | 5001: Read Response Body Failed |
| 6xxx | Database Errors | 6001: Query Data Error |
| 7xxx | Quota Errors | 7001: Insufficient User Quota |

### HTTP Status Mapping

All error codes automatically map to appropriate HTTP status codes:

- **400**: Bad Request (validation errors)
- **401**: Unauthorized (authentication errors)
- **402**: Payment Required (quota errors)
- **403**: Forbidden (access denied)
- **404**: Not Found (model not found)
- **429**: Too Many Requests (rate limit)
- **500**: Internal Server Error (system errors)
- **502**: Bad Gateway (upstream errors)
- **503**: Service Unavailable (no channels)
- **504**: Gateway Timeout (timeout)

### Error Levels

| Level | Description | Use Case |
|-------|-------------|----------|
| **info** | Informational | Normal operations |
| **warning** | Warning | Non-critical issues |
| **error** | Error | Error events |
| **critical** | Critical | System failures |

## Backward Compatibility

### Legacy Support

Old string-based error codes are still available as `LegacyErrorCodeString`:

```go
// Legacy constants (deprecated but still available)
types.LegacyErrorCodeInvalidRequest
types.LegacyErrorCodeChannelNoAvailableKey
// etc.
```

### Conversion Function

Convert between old and new systems:

```go
// Old to New
newCode := types.ErrorCodeFromString("invalid_request")  // 1001

// New to Old
oldCode := types.ErrorCodeInvalidRequest.String()  // "invalid_request"
```

## Testing Your Migration

### 1. Unit Tests

Update error assertions in tests:

```go
// Before
assert.Equal(t, "invalid_request", err.GetErrorCode())

// After
assert.Equal(t, types.ErrorCodeInvalidRequest, err.GetErrorCode())

// Or
assert.Equal(t, 1001, int(err.GetErrorCode()))
```

### 2. Integration Tests

Test HTTP status codes:

```go
// Before
assert.Equal(t, http.StatusBadRequest, w.Code)

// After (still works, status code is auto-mapped)
assert.Equal(t, http.StatusBadRequest, w.Code)
```

### 3. Error Messages

Test error messages:

```go
// String representation works as before
assert.Contains(t, err.Error(), "invalid_request")
```

## Common Issues and Solutions

### Issue 1: Type Mismatch

**Error:**
```go
cannot use "invalid_request" (type untyped string) as type ErrorCode
```

**Solution:**
```go
// Use the constant instead of string
types.ErrorCodeInvalidRequest  // ✅

// Or convert from string
types.ErrorCodeFromString("invalid_request")  // ✅
```

### Issue 2: HTTP Status Code Missing

**Error:**
```go
NewApiErr.StatusCode undefined
```

**Solution:**
```go
// StatusCode is now auto-mapped from ErrorCode
// No need to set it manually
err := types.NewError(e, types.ErrorCodeInvalidRequest)
statusCode := err.StatusCode  // Auto-set to 400
```

### Issue 3: Import Errors

**Error:**
```go
undefined: ErrorCode
```

**Solution:**
```go
// Make sure to import the new types package
import "github.com/QuantumNous/new-api/types"

// Use the ErrorCode type
types.ErrorCodeInvalidRequest
```

## Rollback Plan

If you need to rollback to the old system:

1. Revert `types/error.go` to use `ErrorCode` as `string`
2. Revert `types/error_code.go` to old version
3. Remove `types/error_level.go`
4. Remove `types/error_i18n.go`

However, we recommend keeping the new system for its benefits.

## Getting Help

- **Documentation**: See [ERROR_CODES.md](./ERROR_CODES.md) for complete error code reference
- **Proposal**: See [error-code-improvements.md](./error-code-improvements.md) for design details
- **Issues**: Report problems on GitHub Issues

## Checklist

- [ ] Update error creation code
- [ ] Update error comparison code
- [ ] Update unit tests
- [ ] Update integration tests
- [ ] Update error logging
- [ ] Add error level handling
- [ ] Add localized messages (if needed)
- [ ] Update documentation
- [ ] Remove legacy code (after migration complete)

---

**Questions?** Check the [ERROR_CODES.md](./ERROR_CODES.md) reference or [error-code-improvements.md](./error-code-improvements.md) proposal for more details.
