package service

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/dto"
)

// TestSendSMSNotify 测试SMS通知发送功能
func TestSendSMSNotify(t *testing.T) {
	// 保存原始配置
	originalEndpoint := common.AliyunSMSEndpoint
	originalAccessKeyId := common.AliyunSMSAccessKeyId
	originalAccessKeySecret := common.AliyunSMSAccessKeySecret
	originalSignName := common.AliyunSMSSignName
	originalTemplateCode := common.AliyunSMSTemplateCode

	// 设置测试配置
	common.AliyunSMSEndpoint = "https://dysmsapi.aliyuncs.com"
	common.AliyunSMSAccessKeyId = "test_key_id"
	common.AliyunSMSAccessKeySecret = "test_key_secret"
	common.AliyunSMSSignName = "test_sign"
	common.AliyunSMSTemplateCode = "SMS_123456"

	// 延迟恢复原始配置
	defer func() {
		common.AliyunSMSEndpoint = originalEndpoint
		common.AliyunSMSAccessKeyId = originalAccessKeyId
		common.AliyunSMSAccessKeySecret = originalAccessKeySecret
		common.AliyunSMSSignName = originalSignName
		common.AliyunSMSTemplateCode = originalTemplateCode
	}()

	tests := []struct {
		name        string
		phoneNumber string
		data        dto.Notify
		expectError bool
		errorMsg    string
		setupMock   func() *httptest.Server
	}{
		{
			name:        "successful sms send",
			phoneNumber: "13800000000",
			data: dto.Notify{
				Type:    "quota_exceed",
				Title:   "额度告警",
				Content: "您的额度即将用尽，剩余额度：1000.50，请及时充值。",
				Values:  []interface{}{"您的额度即将用尽", "1000.50"},
			},
			expectError: false,
			setupMock: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					// 验证请求参数
					query := r.URL.Query()
					if query.Get("PhoneNumbers") != "13800000000" {
						t.Errorf("Expected phone number 13800000000, got %s", query.Get("PhoneNumbers"))
					}

					// 返回成功响应
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusOK)
					w.Write([]byte(`{"RequestId":"test-request-id","Code":"OK","Message":"OK","BizId":"test-biz-id"}`))
				}))
			},
		},
		{
			name:        "sms send with api error",
			phoneNumber: "13800000000",
			data: dto.Notify{
				Type:    "quota_exceed",
				Title:   "额度告警",
				Content: "您的额度即将用尽，剩余额度：500.00，请及时充值。",
				Values:  []interface{}{"您的额度即将用尽", "500.00"},
			},
			expectError: true,
			errorMsg:    "阿里云短信发送失败",
			setupMock: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusOK)
					w.Write([]byte(`{"RequestId":"test-request-id","Code":"isv.INVALID_PARAMETERS","Message":"Invalid phone number","BizId":""}`))
				}))
			},
		},
		{
			name:        "sms send with missing values",
			phoneNumber: "13800000000",
			data: dto.Notify{
				Type:    "quota_exceed",
				Title:   "额度告警",
				Content: "额度告警通知",
				Values:  []interface{}{"额度告警通知"}, // 缺少剩余额度参数
			},
			expectError: false, // 应该成功，但模板参数为空
			setupMock: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusOK)
					w.Write([]byte(`{"RequestId":"test-request-id","Code":"OK","Message":"OK","BizId":"test-biz-id"}`))
				}))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 为每个测试用例设置独立的mock服务器
			server := tt.setupMock()
			defer server.Close()
			common.AliyunSMSEndpoint = server.URL

			err := sendSMSNotify(tt.phoneNumber, tt.data)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got nil")
				} else if !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error containing %q, got %q", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
			}
		})
	}
}

// TestSendSMSNotifyEmptyPhoneNumber 测试空手机号情况
func TestSendSMSNotifyEmptyPhoneNumber(t *testing.T) {
	data := dto.Notify{
		Type:    "quota_exceed",
		Title:   "额度告警",
		Content: "您的额度即将用尽",
		Values:  []interface{}{"您的额度即将用尽"},
	}

	err := sendSMSNotify("", data)
	if err == nil {
		t.Error("Expected error for empty phone number, but got nil")
	}

	expectedError := "failed to send sms to"
	if !strings.Contains(err.Error(), expectedError) {
		t.Errorf("Expected error containing %q, got %q", expectedError, err.Error())
	}
}

// TestSendSMSNotifyContentProcessing 测试内容处理逻辑
func TestSendSMSNotifyContentProcessing(t *testing.T) {
	// 保存原始配置
	originalEndpoint := common.AliyunSMSEndpoint
	originalAccessKeyId := common.AliyunSMSAccessKeyId
	originalAccessKeySecret := common.AliyunSMSAccessKeySecret
	originalSignName := common.AliyunSMSSignName
	originalTemplateCode := common.AliyunSMSTemplateCode

	defer func() {
		common.AliyunSMSEndpoint = originalEndpoint
		common.AliyunSMSAccessKeyId = originalAccessKeyId
		common.AliyunSMSAccessKeySecret = originalAccessKeySecret
		common.AliyunSMSSignName = originalSignName
		common.AliyunSMSTemplateCode = originalTemplateCode
	}()

	// 设置测试配置
	common.AliyunSMSEndpoint = "https://dysmsapi.aliyuncs.com"
	common.AliyunSMSAccessKeyId = "test_key_id"
	common.AliyunSMSAccessKeySecret = "test_key_secret"
	common.AliyunSMSSignName = "test_sign"
	common.AliyunSMSTemplateCode = "SMS_123456"

	// 创建mock服务器来验证模板参数
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		templateParam := query.Get("TemplateParam")

		// 验证模板参数是否包含正确的额度信息
		if !strings.Contains(templateParam, "2000.75") {
			t.Errorf("Expected template param to contain quota value 2000.75, got: %s", templateParam)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"RequestId":"test-request-id","Code":"OK","Message":"OK","BizId":"test-biz-id"}`))
	}))
	defer server.Close()

	common.AliyunSMSEndpoint = server.URL

	data := dto.Notify{
		Type:    "quota_exceed",
		Title:   "额度告警",
		Content: "{{value}}，剩余额度：{{value}}，请及时充值。",
		Values:  []interface{}{"您的额度即将用尽", "2000.75"},
	}

	err := sendSMSNotify("13800000000", data)
	if err != nil {
		t.Errorf("Expected successful content processing, got error: %v", err)
	}
}

// BenchmarkSendSMSNotify 基准测试SMS通知发送性能
func BenchmarkSendSMSNotify(b *testing.B) {
	// 保存原始配置
	originalEndpoint := common.AliyunSMSEndpoint
	originalAccessKeyId := common.AliyunSMSAccessKeyId
	originalAccessKeySecret := common.AliyunSMSAccessKeySecret
	originalSignName := common.AliyunSMSSignName
	originalTemplateCode := common.AliyunSMSTemplateCode

	defer func() {
		common.AliyunSMSEndpoint = originalEndpoint
		common.AliyunSMSAccessKeyId = originalAccessKeyId
		common.AliyunSMSAccessKeySecret = originalAccessKeySecret
		common.AliyunSMSSignName = originalSignName
		common.AliyunSMSTemplateCode = originalTemplateCode
	}()

	// 设置测试配置
	common.AliyunSMSEndpoint = "https://dysmsapi.aliyuncs.com"
	common.AliyunSMSAccessKeyId = "test_key_id"
	common.AliyunSMSAccessKeySecret = "test_key_secret"
	common.AliyunSMSSignName = "test_sign"
	common.AliyunSMSTemplateCode = "SMS_123456"

	// 创建快速响应的mock服务器
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"RequestId":"test-request-id","Code":"OK","Message":"OK","BizId":"test-biz-id"}`))
	}))
	defer server.Close()

	common.AliyunSMSEndpoint = server.URL

	data := dto.Notify{
		Type:    "quota_exceed",
		Title:   "额度告警",
		Content: "您的额度即将用尽，剩余额度：1000.00，请及时充值。",
		Values:  []interface{}{"您的额度即将用尽", "1000.00"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sendSMSNotify("13800000000", data)
	}
}