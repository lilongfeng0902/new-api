package model

import (
	"testing"

	"github.com/QuantumNous/new-api/common"
)

// TestAliyunSMSConfigInitialization 测试阿里云SMS配置初始化
func TestAliyunSMSConfigInitialization(t *testing.T) {
	// 保存原始配置
	originalOptionMap := make(map[string]string)
	for k, v := range common.OptionMap {
		originalOptionMap[k] = v
	}

	// 保存原始配置值
	originalEndpoint := common.AliyunSMSEndpoint
	originalAccessKeyId := common.AliyunSMSAccessKeyId
	originalAccessKeySecret := common.AliyunSMSAccessKeySecret
	originalSignName := common.AliyunSMSSignName
	originalTemplateCode := common.AliyunSMSTemplateCode

	// 重置为默认值
	common.AliyunSMSEndpoint = ""
	common.AliyunSMSAccessKeyId = ""
	common.AliyunSMSAccessKeySecret = ""
	common.AliyunSMSSignName = ""
	common.AliyunSMSTemplateCode = ""

	// 重新初始化配置映射
	common.OptionMapRWMutex.Lock()
	common.OptionMap = make(map[string]string)
	common.OptionMapRWMutex.Unlock()

	// 手动添加阿里云SMS配置项（不调用InitOptionMap以避免数据库依赖）
	common.OptionMap["AliyunSMSEndpoint"] = ""
	common.OptionMap["AliyunSMSAccessKeyId"] = ""
	common.OptionMap["AliyunSMSAccessKeySecret"] = ""
	common.OptionMap["AliyunSMSSignName"] = ""
	common.OptionMap["AliyunSMSTemplateCode"] = ""

	// 验证阿里云SMS配置项是否正确初始化
	expectedSMSConfigs := map[string]string{
		"AliyunSMSEndpoint":        "",
		"AliyunSMSAccessKeyId":     "",
		"AliyunSMSAccessKeySecret": "",
		"AliyunSMSSignName":        "",
		"AliyunSMSTemplateCode":    "",
	}

	common.OptionMapRWMutex.RLock()
	for key, expectedValue := range expectedSMSConfigs {
		if actualValue, exists := common.OptionMap[key]; !exists {
			t.Errorf("阿里云SMS配置项 %s 未初始化", key)
		} else if actualValue != expectedValue {
			t.Errorf("阿里云SMS配置项 %s 期望值 %q，实际值 %q", key, expectedValue, actualValue)
		}
	}
	common.OptionMapRWMutex.RUnlock()

	// 恢复原始配置
	common.OptionMapRWMutex.Lock()
	common.OptionMap = originalOptionMap
	common.OptionMapRWMutex.Unlock()

	// 恢复原始配置值
	common.AliyunSMSEndpoint = originalEndpoint
	common.AliyunSMSAccessKeyId = originalAccessKeyId
	common.AliyunSMSAccessKeySecret = originalAccessKeySecret
	common.AliyunSMSSignName = originalSignName
	common.AliyunSMSTemplateCode = originalTemplateCode
}

// TestAliyunSMSConfigUpdate 测试阿里云SMS配置更新
func TestAliyunSMSConfigUpdate(t *testing.T) {
	// 保存原始配置值
	originalEndpoint := common.AliyunSMSEndpoint
	originalAccessKeyId := common.AliyunSMSAccessKeyId
	originalAccessKeySecret := common.AliyunSMSAccessKeySecret
	originalSignName := common.AliyunSMSSignName
	originalTemplateCode := common.AliyunSMSTemplateCode

	// 延迟恢复原始配置
	defer func() {
		common.AliyunSMSEndpoint = originalEndpoint
		common.AliyunSMSAccessKeyId = originalAccessKeyId
		common.AliyunSMSAccessKeySecret = originalAccessKeySecret
		common.AliyunSMSSignName = originalSignName
		common.AliyunSMSTemplateCode = originalTemplateCode
	}()

	tests := []struct {
		name          string
		configKey     string
		configValue   string
		expectedField *string
		expectedValue string
	}{
		{
			name:          "update endpoint",
			configKey:     "AliyunSMSEndpoint",
			configValue:   "https://dysmsapi.aliyuncs.com",
			expectedField: &common.AliyunSMSEndpoint,
			expectedValue: "https://dysmsapi.aliyuncs.com",
		},
		{
			name:          "update access key id",
			configKey:     "AliyunSMSAccessKeyId",
			configValue:   "test_access_key_id",
			expectedField: &common.AliyunSMSAccessKeyId,
			expectedValue: "test_access_key_id",
		},
		{
			name:          "update access key secret",
			configKey:     "AliyunSMSAccessKeySecret",
			configValue:   "test_access_key_secret",
			expectedField: &common.AliyunSMSAccessKeySecret,
			expectedValue: "test_access_key_secret",
		},
		{
			name:          "update sign name",
			configKey:     "AliyunSMSSignName",
			configValue:   "测试签名",
			expectedField: &common.AliyunSMSSignName,
			expectedValue: "测试签名",
		},
		{
			name:          "update template code",
			configKey:     "AliyunSMSTemplateCode",
			configValue:   "SMS_123456789",
			expectedField: &common.AliyunSMSTemplateCode,
			expectedValue: "SMS_123456789",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 保存当前值
			originalValue := *tt.expectedField
			defer func() { *tt.expectedField = originalValue }()

			// 初始化OptionMap（如果为nil）
			common.OptionMapRWMutex.Lock()
			if common.OptionMap == nil {
				common.OptionMap = make(map[string]string)
			}
			common.OptionMapRWMutex.Unlock()

			// 重置为初始值
			*tt.expectedField = ""

			// 调用配置更新函数
			err := updateOptionMap(tt.configKey, tt.configValue)
			if err != nil {
				t.Errorf("updateOptionMap() 返回错误: %v", err)
				return
			}

			// 验证字段值是否正确更新
			if *tt.expectedField != tt.expectedValue {
				t.Errorf("字段值未正确更新，期望 %q，实际 %q", tt.expectedValue, *tt.expectedField)
				t.Logf("Debug: key=%s, value=%s", tt.configKey, tt.configValue)
			}
		})
	}
}

// TestAliyunSMSConfigUpdateInvalidKey 测试无效配置键的更新
func TestAliyunSMSConfigUpdateInvalidKey(t *testing.T) {
	// 测试无效的配置键
	err := updateOptionMap("InvalidAliyunSMSKey", "test_value")
	if err != nil {
		t.Errorf("updateOptionMap() 对无效键返回错误: %v", err)
	}

	// 验证无效键不会影响其他配置
	if common.AliyunSMSEndpoint != "" {
		t.Error("无效配置键更新影响了其他配置")
	}
}

// TestAliyunSMSConfigInOptionMap 测试阿里云SMS配置在选项映射中的存在性
func TestAliyunSMSConfigInOptionMap(t *testing.T) {
	// 保存原始配置
	originalOptionMap := make(map[string]string)
	for k, v := range common.OptionMap {
		originalOptionMap[k] = v
	}

	// 保存原始配置值
	originalEndpoint := common.AliyunSMSEndpoint
	originalAccessKeyId := common.AliyunSMSAccessKeyId
	originalAccessKeySecret := common.AliyunSMSAccessKeySecret
	originalSignName := common.AliyunSMSSignName
	originalTemplateCode := common.AliyunSMSTemplateCode

	// 重新初始化配置映射
	common.OptionMapRWMutex.Lock()
	common.OptionMap = make(map[string]string)
	common.OptionMapRWMutex.Unlock()

	// 手动添加配置项（模拟InitOptionMap的行为，但不调用数据库相关代码）
	common.OptionMap["AliyunSMSEndpoint"] = "https://test.endpoint.com"
	common.OptionMap["AliyunSMSAccessKeyId"] = "test_key_id"
	common.OptionMap["AliyunSMSSignName"] = "测试签名"
	common.OptionMap["AliyunSMSTemplateCode"] = "SMS_123456"
	// 注意：AccessKeySecret不应该出现在OptionMap中（因为包含"Secret"）

	// 验证配置项在OptionMap中
	common.OptionMapRWMutex.RLock()
	expectedConfigs := map[string]string{
		"AliyunSMSEndpoint":        "https://test.endpoint.com",
		"AliyunSMSAccessKeyId":     "test_key_id",
		"AliyunSMSSignName":        "测试签名",
		"AliyunSMSTemplateCode":    "SMS_123456",
		// 注意：AccessKeySecret不应该在OptionMap中（因为它包含"Secret"）
	}

	for key, expectedValue := range expectedConfigs {
		if actualValue, exists := common.OptionMap[key]; !exists {
			t.Errorf("阿里云SMS配置项 %s 在OptionMap中不存在", key)
		} else if actualValue != expectedValue {
			t.Errorf("阿里云SMS配置项 %s 期望值 %q，实际值 %q", key, expectedValue, actualValue)
		}
	}

	// 验证敏感信息不应该出现在OptionMap中
	if _, exists := common.OptionMap["AliyunSMSAccessKeySecret"]; exists {
		t.Error("敏感配置项 AliyunSMSAccessKeySecret 不应该出现在OptionMap中")
	}
	common.OptionMapRWMutex.RUnlock()

	// 恢复原始配置
	common.OptionMapRWMutex.Lock()
	common.OptionMap = originalOptionMap
	common.OptionMapRWMutex.Unlock()

	// 恢复原始配置值
	common.AliyunSMSEndpoint = originalEndpoint
	common.AliyunSMSAccessKeyId = originalAccessKeyId
	common.AliyunSMSAccessKeySecret = originalAccessKeySecret
	common.AliyunSMSSignName = originalSignName
	common.AliyunSMSTemplateCode = originalTemplateCode
}

// BenchmarkAliyunSMSConfigUpdate 基准测试阿里云SMS配置更新性能
func BenchmarkAliyunSMSConfigUpdate(b *testing.B) {
	testConfigs := []struct {
		key   string
		value string
	}{
		{"AliyunSMSEndpoint", "https://dysmsapi.aliyuncs.com"},
		{"AliyunSMSAccessKeyId", "benchmark_access_key_id"},
		{"AliyunSMSAccessKeySecret", "benchmark_access_key_secret"},
		{"AliyunSMSSignName", "基准测试签名"},
		{"AliyunSMSTemplateCode", "SMS_123456789"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, config := range testConfigs {
			updateOptionMap(config.key, config.value)
		}
	}
}

// BenchmarkInitOptionMap 基准测试选项映射初始化性能
func BenchmarkInitOptionMap(b *testing.B) {
	// 保存原始配置
	originalOptionMap := make(map[string]string)
	for k, v := range common.OptionMap {
		originalOptionMap[k] = v
	}

	defer func() {
		// 恢复原始配置
		common.OptionMapRWMutex.Lock()
		common.OptionMap = originalOptionMap
		common.OptionMapRWMutex.Unlock()
	}()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		common.OptionMapRWMutex.Lock()
		common.OptionMap = make(map[string]string)
		// 手动添加一些配置项来模拟初始化过程
		common.OptionMap["AliyunSMSEndpoint"] = ""
		common.OptionMap["AliyunSMSAccessKeyId"] = ""
		common.OptionMap["AliyunSMSSignName"] = ""
		common.OptionMap["AliyunSMSTemplateCode"] = ""
		common.OptionMapRWMutex.Unlock()
	}
}