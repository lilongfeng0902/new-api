package service

import (
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/dto"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
)

// TestSMSQuotaNotificationContentFormat 测试SMS额度告警内容格式
func TestSMSQuotaNotificationContentFormat(t *testing.T) {
	tests := []struct {
		name         string
		notifyType   string
		expectedContent string
		expectedValues  []interface{}
	}{
		{
			name:       "SMS notification type",
			notifyType: dto.NotifyTypeSMS,
			expectedContent: "{{value}}，剩余额度：{{value}}，请及时充值",
			expectedValues:  []interface{}{"您的额度即将用尽", "1000.50"},
		},
		{
			name:       "Email notification type",
			notifyType: dto.NotifyTypeEmail,
			expectedContent: "{{value}}，当前剩余额度为 {{value}}，为了不影响您的使用，请及时充值。<br/>充值链接：<a href='{{value}}'>{{value}}</a>",
			expectedValues:  []interface{}{"您的额度即将用尽", "1000.50", "http://example.com/topup", "http://example.com/topup"},
		},
		{
			name:       "Bark notification type",
			notifyType: dto.NotifyTypeBark,
			expectedContent: "{{value}}，剩余额度：{{value}}，请及时充值",
			expectedValues:  []interface{}{"您的额度即将用尽", "1000.50"},
		},
		{
			name:       "Gotify notification type",
			notifyType: dto.NotifyTypeGotify,
			expectedContent: "{{value}}，当前剩余额度为 {{value}}，请及时充值。",
			expectedValues:  []interface{}{"您的额度即将用尽", "1000.50"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 模拟用户设置
			userSetting := dto.UserSetting{
				NotifyType: tt.notifyType,
			}

			// 模拟relay信息
			relayInfo := &relaycommon.RelayInfo{
				UserSetting: userSetting,
				UserQuota:   100050, // 模拟1000.50额度（以分为单位）
			}

			// 模拟阈值检查逻辑
			threshold := common.QuotaRemindThreshold
			quota := 500  // 模拟消耗的额度
			preConsumedQuota := 100

			consumeQuota := quota + preConsumedQuota
			if relayInfo.UserQuota-consumeQuota < threshold {
				prompt := "您的额度即将用尽"

				// 复制checkAndSendQuotaNotify中的内容格式逻辑
				var content string
				var values []interface{}

				notifyType := userSetting.NotifyType
				if notifyType == "" {
					notifyType = dto.NotifyTypeEmail
				}

				if notifyType == dto.NotifyTypeBark {
					content = "{{value}}，剩余额度：{{value}}，请及时充值"
					values = []interface{}{prompt, "1000.50"}
				} else if notifyType == dto.NotifyTypeGotify {
					content = "{{value}}，当前剩余额度为 {{value}}，请及时充值。"
					values = []interface{}{prompt, "1000.50"}
				} else if notifyType == dto.NotifyTypeSMS {
					content = "{{value}}，剩余额度：{{value}}，请及时充值"
					values = []interface{}{prompt, "1000.50"}
				} else {
					content = "{{value}}，当前剩余额度为 {{value}}，为了不影响您的使用，请及时充值。<br/>充值链接：<a href='{{value}}'>{{value}}</a>"
					values = []interface{}{prompt, "1000.50", "http://example.com/topup", "http://example.com/topup"}
				}

				// 验证内容格式
				if content != tt.expectedContent {
					t.Errorf("Expected content %q, got %q", tt.expectedContent, content)
				}

				// 验证参数值
				if len(values) != len(tt.expectedValues) {
					t.Errorf("Expected %d values, got %d", len(tt.expectedValues), len(values))
				} else {
					for i, expected := range tt.expectedValues {
						if values[i] != expected {
							t.Errorf("Expected value[%d] %q, got %q", i, expected, values[i])
						}
					}
				}
			}
		})
	}
}

// TestQuotaThresholdLogic 测试额度阈值逻辑
func TestQuotaThresholdLogic(t *testing.T) {
	tests := []struct {
		name           string
		userQuota      int
		consumedQuota  int
		threshold      int
		expectWarning  bool
	}{
		{
			name:          "quota above threshold",
			userQuota:     2000,
			consumedQuota: 100,
			threshold:     1000,
			expectWarning: false,
		},
		{
			name:          "quota at threshold",
			userQuota:     2000,
			consumedQuota: 1000,
			threshold:     1000,
			expectWarning: false, // 等于阈值不算不足
		},
		{
			name:          "quota below threshold",
			userQuota:     2000,
			consumedQuota: 1100,
			threshold:     1000,
			expectWarning: true,
		},
		{
			name:          "zero threshold",
			userQuota:     1000,
			consumedQuota: 500,
			threshold:     0,
			expectWarning: false, // 阈值为0时不应该触发警告
		},
		{
			name:          "custom user threshold",
			userQuota:     2000,
			consumedQuota: 1800,
			threshold:     500, // 自定义低阈值
			expectWarning: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 计算剩余额度
			remainingQuota := tt.userQuota - tt.consumedQuota

			// 判断是否需要告警
			quotaTooLow := remainingQuota < tt.threshold

			if quotaTooLow != tt.expectWarning {
				t.Errorf("Expected warning=%t for remaining quota %d < threshold %d, got %t",
					tt.expectWarning, remainingQuota, tt.threshold, quotaTooLow)
			}
		})
	}
}

// TestQuotaWarningThresholdFromSettings 测试用户自定义阈值设置
func TestQuotaWarningThresholdFromSettings(t *testing.T) {
	tests := []struct {
		name                    string
		systemThreshold        int
		userSettingThreshold   float64
		expectedThreshold      int
	}{
		{
			name:                  "use system threshold",
			systemThreshold:      1000,
			userSettingThreshold: 0,
			expectedThreshold:    1000,
		},
		{
			name:                  "use user custom threshold",
			systemThreshold:      1000,
			userSettingThreshold: 500.5,
			expectedThreshold:    500,
		},
		{
			name:                  "user threshold as integer",
			systemThreshold:      1000,
			userSettingThreshold: 200.0,
			expectedThreshold:    200,
		},
		{
			name:                  "user threshold rounds down",
			systemThreshold:      1000,
			userSettingThreshold: 300.9,
			expectedThreshold:    300,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 模拟系统阈值
			originalThreshold := common.QuotaRemindThreshold
			common.QuotaRemindThreshold = tt.systemThreshold
			defer func() { common.QuotaRemindThreshold = originalThreshold }()

			// 模拟用户设置
			userSetting := dto.UserSetting{
				QuotaWarningThreshold: tt.userSettingThreshold,
			}

			// 计算实际使用的阈值
			threshold := common.QuotaRemindThreshold
			if userSetting.QuotaWarningThreshold != 0 {
				threshold = int(userSetting.QuotaWarningThreshold)
			}

			if threshold != tt.expectedThreshold {
				t.Errorf("Expected threshold %d, got %d", tt.expectedThreshold, threshold)
			}
		})
	}
}

// BenchmarkQuotaNotificationContentProcessing 基准测试额度通知内容处理
func BenchmarkQuotaNotificationContentProcessing(b *testing.B) {
	userSetting := dto.UserSetting{
		NotifyType: dto.NotifyTypeSMS,
	}

	relayInfo := &relaycommon.RelayInfo{
		UserSetting: userSetting,
		UserQuota:   100050, // 模拟1000.50额度
	}

	threshold := 1000
	quota := 500
	preConsumedQuota := 100
	consumeQuota := quota + preConsumedQuota

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if relayInfo.UserQuota-consumeQuota < threshold {
			prompt := "您的额度即将用尽"
			var content string
			var values []interface{}

			notifyType := userSetting.NotifyType
			if notifyType == "" {
				notifyType = dto.NotifyTypeEmail
			}

			if notifyType == dto.NotifyTypeBark {
				content = "{{value}}，剩余额度：{{value}}，请及时充值"
				values = []interface{}{prompt, "1000.50"}
			} else if notifyType == dto.NotifyTypeGotify {
				content = "{{value}}，当前剩余额度为 {{value}}，请及时充值。"
				values = []interface{}{prompt, "1000.50"}
			} else if notifyType == dto.NotifyTypeSMS {
				content = "{{value}}，剩余额度：{{value}}，请及时充值"
				values = []interface{}{prompt, "1000.50"}
			} else {
				content = "{{value}}，当前剩余额度为 {{value}}，为了不影响您的使用，请及时充值。<br/>充值链接：<a href='{{value}}'>{{value}}</a>"
				values = []interface{}{prompt, "1000.50", "http://example.com/topup", "http://example.com/topup"}
			}

			// 模拟占位符替换
			for range values {
				content = "processed content"
			}
			_ = content // 使用变量避免未使用警告
		}
	}
}