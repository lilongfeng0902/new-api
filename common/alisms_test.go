package common

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"
)

// TestPercentEncode 测试URL编码函数
func TestPercentEncode(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "normal string",
			input:    "hello world",
			expected: "hello%20world",
		},
		{
			name:     "special characters",
			input:    "test@example.com",
			expected: "test%40example.com",
		},
		{
			name:     "chinese characters",
			input:    "测试",
			expected: "%E6%B5%8B%E8%AF%95",
		},
		{
			name:     "plus and asterisk",
			input:    "a+b*c",
			expected: "a%2Bb%2Ac",
		},
		{
			name:     "tilde character",
			input:    "~test~",
			expected: "~test~",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := percentEncode(tt.input)
			if result != tt.expected {
				t.Errorf("percentEncode(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

// TestSignStringToString 测试签名生成函数
func TestSignStringToString(t *testing.T) {
	method := "GET"
	canonicalizedQueryString := "AccessKeyId=test&Action=SendSms&Format=JSON&PhoneNumbers=13800000000&RegionId=cn-hangzhou&SignName=test&SignatureMethod=HMAC-SHA1&SignatureNonce=test&SignatureVersion=1.0&TemplateCode=SMS_123&TemplateParam=%7B%22code%22%3A%221234%22%7D&Timestamp=2024-01-01T12%3A00%3A00Z&Version=2017-05-25"
	accessKeySecret := "test_secret"

	result := signStringToString(method, canonicalizedQueryString, accessKeySecret)

	// 由于签名算法的确定性，我们可以验证结果不为空且格式正确
	if result == "" {
		t.Error("signStringToString returned empty result")
	}

	// Base64编码的结果应该只包含特定的字符集
	validChars := "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/="
	for _, char := range result {
		if !strings.ContainsRune(validChars, char) && char != '=' {
			t.Errorf("signStringToString returned invalid base64 character: %c", char)
		}
	}
}

// TestBuildCanonicalizedQueryString 测试规范化查询字符串构建
func TestBuildCanonicalizedQueryString(t *testing.T) {
	tests := []struct {
		name     string
		params   map[string]string
		expected string
	}{
		{
			name: "single parameter",
			params: map[string]string{
				"key1": "value1",
			},
			expected: "key1=value1",
		},
		{
			name: "multiple parameters",
			params: map[string]string{
				"key2": "value2",
				"key1": "value1",
			},
			expected: "key1=value1&key2=value2",
		},
		{
			name: "parameters with special characters",
			params: map[string]string{
				"key with spaces": "value with spaces",
				"key@domain.com":  "value@domain.com",
			},
			expected: "key%20with%20spaces=value%20with%20spaces&key%40domain.com=value%40domain.com",
		},
		{
			name: "empty value parameter",
			params: map[string]string{
				"key1": "",
				"key2": "value2",
			},
			expected: "key2=value2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := buildCanonicalizedQueryString(tt.params)
			if result != tt.expected {
				t.Errorf("buildCanonicalizedQueryString() = %q, want %q", result, tt.expected)
			}
		})
	}
}

// TestSendAliyunSMS 测试阿里云短信发送（使用mock服务器）
func TestSendAliyunSMS(t *testing.T) {
	// 保存原始配置
	originalEndpoint := AliyunSMSEndpoint
	originalAccessKeyId := AliyunSMSAccessKeyId
	originalAccessKeySecret := AliyunSMSAccessKeySecret
	originalSignName := AliyunSMSSignName
	originalTemplateCode := AliyunSMSTemplateCode

	// 设置测试配置
	AliyunSMSEndpoint = ""
	AliyunSMSAccessKeyId = "test_key_id"
	AliyunSMSAccessKeySecret = "test_key_secret"
	AliyunSMSSignName = "test_sign"
	AliyunSMSTemplateCode = "SMS_123456"

	// 延迟恢复原始配置
	defer func() {
		AliyunSMSEndpoint = originalEndpoint
		AliyunSMSAccessKeyId = originalAccessKeyId
		AliyunSMSAccessKeySecret = originalAccessKeySecret
		AliyunSMSSignName = originalSignName
		AliyunSMSTemplateCode = originalTemplateCode
	}()

	// 测试配置不完整的情况
	t.Run("missing endpoint", func(t *testing.T) {
		err := SendAliyunSMS("13800000000", nil, "")
		if err == nil {
			t.Error("Expected error for missing endpoint, but got nil")
		}
		if !strings.Contains(err.Error(), "阿里云短信配置未完成") {
			t.Errorf("Expected configuration error, got: %v", err)
		}
	})

	// 设置端点
	AliyunSMSEndpoint = "https://dysmsapi.aliyuncs.com"

	// 创建mock服务器
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 验证请求方法
		if r.Method != "GET" {
			t.Errorf("Expected GET request, got %s", r.Method)
		}

		// 验证必要的查询参数
		query := r.URL.Query()
		requiredParams := []string{"AccessKeyId", "Action", "Format", "PhoneNumbers", "SignName", "TemplateCode"}
		for _, param := range requiredParams {
			if query.Get(param) == "" {
				t.Errorf("Missing required parameter: %s", param)
			}
		}

		// 返回成功的响应
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"RequestId":"test-request-id","Code":"OK","Message":"OK","BizId":"test-biz-id"}`))
	}))
	defer server.Close()

	AliyunSMSEndpoint = server.URL

	t.Run("successful send", func(t *testing.T) {
		templateParam := map[string]string{
			"code": "1234",
		}
		err := SendAliyunSMS("13800000000", templateParam, "")
		if err != nil {
			t.Errorf("Expected successful send, got error: %v", err)
		}
	})

	// 测试失败响应
	server.Close()
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"RequestId":"test-request-id","Code":"isv.INVALID_PARAMETERS","Message":"Invalid parameters","BizId":""}`))
	}))
	defer server.Close()

	AliyunSMSEndpoint = server.URL

	t.Run("api error response", func(t *testing.T) {
		err := SendAliyunSMS("13800000000", nil, "")
		if err == nil {
			t.Error("Expected error for API failure, but got nil")
		}
		if !strings.Contains(err.Error(), "isv.INVALID_PARAMETERS") {
			t.Errorf("Expected API error, got: %v", err)
		}
	})
}

// TestSendQuotaWarningSMS 测试额度告警短信发送
func TestSendQuotaWarningSMS(t *testing.T) {
	// 保存和恢复配置
	originalEndpoint := AliyunSMSEndpoint
	originalAccessKeyId := AliyunSMSAccessKeyId
	originalAccessKeySecret := AliyunSMSAccessKeySecret
	originalSignName := AliyunSMSSignName
	originalTemplateCode := AliyunSMSTemplateCode

	defer func() {
		AliyunSMSEndpoint = originalEndpoint
		AliyunSMSAccessKeyId = originalAccessKeyId
		AliyunSMSAccessKeySecret = originalAccessKeySecret
		AliyunSMSSignName = originalSignName
		AliyunSMSTemplateCode = originalTemplateCode
	}()

	// 设置测试配置
	AliyunSMSEndpoint = "https://dysmsapi.aliyuncs.com"
	AliyunSMSAccessKeyId = "test_key_id"
	AliyunSMSAccessKeySecret = "test_key_secret"
	AliyunSMSSignName = "test_sign"
	AliyunSMSTemplateCode = "SMS_123456"

	// 创建mock服务器
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()

		// 验证TemplateParam参数
		templateParam := query.Get("TemplateParam")
		if templateParam == "" {
			t.Error("Missing TemplateParam in request")
		}

		// 这里可以进一步验证JSON内容，但为简单起见只检查存在性

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"RequestId":"test-request-id","Code":"OK","Message":"OK","BizId":"test-biz-id"}`))
	}))
	defer server.Close()

	AliyunSMSEndpoint = server.URL

	t.Run("quota warning sms", func(t *testing.T) {
		err := SendQuotaWarningSMS("13800000000", "1000.50")
		if err != nil {
			t.Errorf("Expected successful quota warning SMS send, got error: %v", err)
		}
	})
}

// BenchmarkPercentEncode 基准测试URL编码性能
func BenchmarkPercentEncode(b *testing.B) {
	testStrings := []string{
		"hello world",
		"test@example.com",
		"https://api.aliyun.com/path?param=value&another=test",
		"中文测试字符串",
		"special chars: !@#$%^&*()_+-=[]{}|;':\",./<>?",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, str := range testStrings {
			percentEncode(str)
		}
	}
}

// BenchmarkBuildCanonicalizedQueryString 基准测试查询字符串构建性能
func BenchmarkBuildCanonicalizedQueryString(b *testing.B) {
	params := map[string]string{
		"AccessKeyId":      "test_key",
		"Action":           "SendSms",
		"Format":           "JSON",
		"PhoneNumbers":     "13800000000",
		"RegionId":         "cn-hangzhou",
		"SignName":         "test_sign",
		"SignatureMethod":  "HMAC-SHA1",
		"SignatureNonce":   "test_nonce",
		"SignatureVersion": "1.0",
		"TemplateCode":     "SMS_123",
		"TemplateParam":    "{\"code\":\"1234\"}",
		"Timestamp":        "2024-01-01T12:00:00Z",
		"Version":          "2017-05-25",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buildCanonicalizedQueryString(params)
	}
}

// TestSendAliyunSMSIntegration 集成测试：对接真实的阿里云SMS服务
// 此测试需要真实的阿里云SMS配置，会产生实际的短信费用
// 运行方式：go test -tags integration -run TestSendAliyunSMSIntegration -v ./common
//
// 环境变量配置：
//
//	ALIYUN_SMS_INTEGRATION_TEST=true          # 启用集成测试
//	ALIYUN_SMS_ACCESS_KEY_ID=your_key_id      # AccessKey ID
//	ALIYUN_SMS_ACCESS_KEY_SECRET=your_secret  # AccessKey Secret
//	ALIYUN_SMS_SIGN_NAME=your_sign_name       # 短信签名
//	ALIYUN_SMS_TEMPLATE_CODE=SMS_xxxxxxxxx    # 短信模板CODE
//	ALIYUN_SMS_TEST_PHONE=13800000000        # 测试手机号
//
// 注意：
//  1. 此测试会发送真实短信，请谨慎使用
//  2. 确保账户有足够余额
//  3. 短信签名和模板需要审核通过
//  4. 建议只在staging环境运行
func TestSendAliyunSMSIntegration(t *testing.T) {
	// 检查是否启用集成测试
	if os.Getenv("ALIYUN_SMS_INTEGRATION_TEST") != "true" {
		t.Skip("集成测试未启用，请设置 ALIYUN_SMS_INTEGRATION_TEST=true 启用")
	}

	// 获取配置
	accessKeyId := os.Getenv("ALIYUN_SMS_ACCESS_KEY_ID")
	accessKeySecret := os.Getenv("ALIYUN_SMS_ACCESS_KEY_SECRET")
	signName := os.Getenv("ALIYUN_SMS_SIGN_NAME")
	templateCode := os.Getenv("ALIYUN_SMS_TEMPLATE_CODE")
	testPhone := os.Getenv("ALIYUN_SMS_TEST_PHONE")

	// 验证必需的配置
	requiredEnvVars := map[string]string{
		"ALIYUN_SMS_ACCESS_KEY_ID":     accessKeyId,
		"ALIYUN_SMS_ACCESS_KEY_SECRET": accessKeySecret,
		"ALIYUN_SMS_SIGN_NAME":         signName,
		"ALIYUN_SMS_TEMPLATE_CODE":     templateCode,
		"ALIYUN_SMS_TEST_PHONE":        testPhone,
	}

	missingVars := []string{}
	for envVar, value := range requiredEnvVars {
		if value == "" {
			missingVars = append(missingVars, envVar)
		}
	}

	if len(missingVars) > 0 {
		t.Fatalf("缺少必需的环境变量: %s", strings.Join(missingVars, ", "))
	}

	// 保存原始配置
	originalEndpoint := AliyunSMSEndpoint
	originalAccessKeyId := AliyunSMSAccessKeyId
	originalAccessKeySecret := AliyunSMSAccessKeySecret
	originalSignName := AliyunSMSSignName
	originalTemplateCode := AliyunSMSTemplateCode

	// 设置测试配置
	AliyunSMSEndpoint = "https://dysmsapi.aliyuncs.com"
	AliyunSMSAccessKeyId = accessKeyId
	AliyunSMSAccessKeySecret = accessKeySecret
	AliyunSMSSignName = signName
	AliyunSMSTemplateCode = templateCode

	// 延迟恢复原始配置
	defer func() {
		AliyunSMSEndpoint = originalEndpoint
		AliyunSMSAccessKeyId = originalAccessKeyId
		AliyunSMSAccessKeySecret = originalAccessKeySecret
		AliyunSMSSignName = originalSignName
		AliyunSMSTemplateCode = originalTemplateCode
	}()

	t.Run("send_simple_sms", func(t *testing.T) {
		// 发送简单短信（无模板参数）
		err := SendAliyunSMS(testPhone, nil, "")
		if err != nil {
			// 检查是否为预期的API错误（比如模板需要参数）
			if strings.Contains(err.Error(), "TemplateParam") ||
				strings.Contains(err.Error(), "isv.") {
				t.Logf("预期的API错误（模板需要参数）: %v", err)
			} else {
				t.Errorf("发送简单短信失败: %v", err)
			}
		} else {
			t.Log("简单短信发送成功")
		}
	})

	t.Run("send_template_sms", func(t *testing.T) {
		// 发送模板短信
		templateParam := map[string]string{
			"code": "123456",
		}

		err := SendAliyunSMS(testPhone, templateParam, "")
		if err != nil {
			t.Errorf("发送模板短信失败: %v", err)
		} else {
			t.Log("模板短信发送成功")
		}
	})

	t.Run("send_quota_warning_sms", func(t *testing.T) {
		// 发送额度告警短信
		err := SendQuotaWarningSMS(testPhone, "500.00")
		if err != nil {
			t.Errorf("发送额度告警短信失败: %v", err)
		} else {
			t.Log("额度告警短信发送成功")
		}
	})

	t.Run("test_invalid_phone", func(t *testing.T) {
		// 测试无效手机号
		err := SendAliyunSMS("invalid_phone", nil, "")
		if err == nil {
			t.Error("期望无效手机号发送失败，但成功了")
		} else {
			t.Logf("无效手机号正确返回错误: %v", err)
		}
	})

	t.Run("test_rate_limit", func(t *testing.T) {
		// 测试发送频率（阿里云SMS有频率限制）
		templateParam := map[string]string{
			"code": "789012",
		}

		// 发送多次短信测试频率限制
		for i := 0; i < 3; i++ {
			err := SendAliyunSMS(testPhone, templateParam, "")
			if err != nil {
				if strings.Contains(err.Error(), "isv.BUSINESS_LIMIT_CONTROL") ||
					strings.Contains(err.Error(), "限流") {
					t.Logf("触发频率限制（预期行为）: %v", err)
					return // 频率限制是正常的，测试通过
				} else {
					t.Errorf("第%d次发送失败: %v", i+1, err)
				}
			} else {
				t.Logf("第%d次发送成功", i+1)
				// 发送成功后等待一下再发下一条
				time.Sleep(2 * time.Second)
			}
		}
	})
}

// TestAliyunSMSIntegrationSetup 验证集成测试环境配置
// 此测试在正常测试时也会运行，用于检查环境变量配置
func TestAliyunSMSIntegrationSetup(t *testing.T) {
	t.Run("check_environment_variables", func(t *testing.T) {
		envVars := []string{
			"ALIYUN_SMS_INTEGRATION_TEST",
			"ALIYUN_SMS_ACCESS_KEY_ID",
			"ALIYUN_SMS_ACCESS_KEY_SECRET",
			"ALIYUN_SMS_SIGN_NAME",
			"ALIYUN_SMS_TEMPLATE_CODE",
			"ALIYUN_SMS_TEST_PHONE",
		}

		setVars := []string{}
		missingVars := []string{}

		for _, envVar := range envVars {
			if value := os.Getenv(envVar); value != "" {
				setVars = append(setVars, envVar)
			} else {
				missingVars = append(missingVars, envVar)
			}
		}

		t.Logf("已设置的环境变量: %s", strings.Join(setVars, ", "))
		if len(missingVars) > 0 {
			t.Logf("未设置的环境变量: %s", strings.Join(missingVars, ", "))
		}

		if os.Getenv("ALIYUN_SMS_INTEGRATION_TEST") == "true" {
			t.Log("集成测试已启用，可以运行真实环境测试")
		} else {
			t.Log("集成测试未启用，只会运行单元测试")
		}
	})

	t.Run("validate_test_phone_format", func(t *testing.T) {
		testPhone := os.Getenv("ALIYUN_SMS_TEST_PHONE")
		if testPhone == "" {
			t.Skip("测试手机号未设置")
		}

		// 验证手机号格式（中国大陆手机号）
		if len(testPhone) != 11 || !strings.HasPrefix(testPhone, "1") {
			t.Errorf("测试手机号格式不正确: %s（应为11位以1开头的手机号）", testPhone)
		} else {
			t.Logf("测试手机号格式正确: %s", testPhone)
		}
	})
}
