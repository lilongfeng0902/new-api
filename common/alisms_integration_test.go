//go:build integration
// +build integration

package common

import (
	"os"
	"strings"
	"testing"
)

// TestSendAliyunSMSRealIntegration 集成测试：对接真实的阿里云SMS服务
// 此测试需要真实的阿里云SMS配置，会产生实际的短信费用
//
// 运行方式:
//
//	go test -tags integration -run TestSendAliyunSMSRealIntegration -v ./common
//
// 环境变量配置:
//
//	ALIYUN_SMS_INTEGRATION_TEST=true          # 启用集成测试
//	ALIYUN_SMS_ACCESS_KEY_ID=your_key_id      # AccessKey ID
//	ALIYUN_SMS_ACCESS_KEY_SECRET=your_secret  # AccessKey Secret
//	ALIYUN_SMS_SIGN_NAME=your_sign_name       # 短信签名
//	ALIYUN_SMS_TEMPLATE_CODE=SMS_xxxxxxxxx    # 短信模板CODE
//	ALIYUN_SMS_TEST_PHONE=13800000000        # 测试手机号
//
// 注意:
//  1. 此测试会发送真实短信，请谨慎使用
//  2. 确保账户有足够余额
//  3. 短信签名和模板需要审核通过
//  4. 建议只在staging环境运行
//  5. 正常运行 go test 不会执行此测试（需要 -tags integration）
func TestSendAliyunSMSRealIntegration(t *testing.T) {
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

	// t.Run("send_simple_sms", func(t *testing.T) {
	// 	// 发送简单短信（无模板参数）
	// 	err := SendAliyunSMS(testPhone, nil)
	// 	if err != nil {
	// 		// 检查是否为预期的API错误（比如模板需要参数）
	// 		if strings.Contains(err.Error(), "TemplateParam") ||
	// 			strings.Contains(err.Error(), "isv.") {
	// 			t.Logf("预期的API错误（模板需要参数）: %v", err)
	// 		} else {
	// 			t.Errorf("发送简单短信失败: %v", err)
	// 		}
	// 	} else {
	// 		t.Log("简单短信发送成功")
	// 	}
	// })

	// t.Run("send_template_sms", func(t *testing.T) {
	// 	// 发送模板短信
	// 	templateParam := map[string]string{
	// 		"money": "123456",
	// 	}

	// 	err := SendAliyunSMS(testPhone, templateParam)
	// 	if err != nil {
	// 		t.Errorf("发送模板短信失败: %v", err)
	// 	} else {
	// 		t.Log("模板短信发送成功")
	// 	}
	// })

	t.Run("send_quota_warning_sms", func(t *testing.T) {
		// 发送额度告警短信
		err := SendQuotaWarningSMS(testPhone, "500.00")
		if err != nil {
			t.Errorf("发送额度告警短信失败: %v", err)
		} else {
			t.Log("额度告警短信发送成功")
		}
	})

	// t.Run("test_invalid_phone", func(t *testing.T) {
	// 	// 测试无效手机号
	// 	err := SendAliyunSMS("invalid_phone", nil)
	// 	if err == nil {
	// 		t.Error("期望无效手机号发送失败，但成功了")
	// 	} else {
	// 		t.Logf("无效手机号正确返回错误: %v", err)
	// 	}
	// })

	// t.Run("test_rate_limit", func(t *testing.T) {
	// 	// 测试发送频率（阿里云SMS有频率限制）
	// 	templateParam := map[string]string{
	// 		"code": "789012",
	// 	}

	// 	// 发送多次短信测试频率限制
	// 	for i := 0; i < 3; i++ {
	// 		err := SendAliyunSMS(testPhone, templateParam)
	// 		if err != nil {
	// 			if strings.Contains(err.Error(), "isv.BUSINESS_LIMIT_CONTROL") ||
	// 				strings.Contains(err.Error(), "限流") {
	// 				t.Logf("触发频率限制（预期行为）: %v", err)
	// 				return // 频率限制是正常的，测试通过
	// 			} else {
	// 				t.Errorf("第%d次发送失败: %v", i+1, err)
	// 			}
	// 		} else {
	// 			t.Logf("第%d次发送成功", i+1)
	// 			// 发送成功后等待一下再发下一条
	// 			time.Sleep(2 * time.Second)
	// 		}
	// 	}
	// })
}
