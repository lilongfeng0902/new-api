package service

import (
	"context"
	"fmt"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
)

// SmsSendService SMS sending service
type SmsSendService struct{}

// NewSmsSendService creates a new SMS send service
func NewSmsSendService() *SmsSendService {
	return &SmsSendService{}
}

// SendVerificationCode sends SMS and logs the result
func (s *SmsSendService) SendVerificationCode(ctx context.Context, mobile, code, templateCode string, userId *int) error {
	// Create log entry
	smsLog := &model.SmsLog{
		TemplateCode: templateCode,
		Mobile:       mobile,
		UserId:       userId,
		SendStatus:   model.SmsLogStatusPending,
		SendTime:     common.GetTimestamp(),
	}

	// Template parameters
	templateParam := map[string]string{
		"code": code,
	}

	// Send SMS via Aliyun
	err := common.SendAliyunSMS(mobile, templateParam, templateCode)

	// Update log based on result
	if err != nil {
		smsLog.SendStatus = model.SmsLogStatusFailed
		smsLog.ApiSendCode = "ERROR"
		smsLog.ApiSendMsg = err.Error()
	} else {
		smsLog.SendStatus = model.SmsLogStatusSuccess
		smsLog.ApiSendCode = "OK"
		smsLog.ApiSendMsg = "发送成功"
	}

	// Save log (don't fail if logging fails)
	logErr := model.DB.Create(smsLog).Error
	if logErr != nil {
		common.SysLog(fmt.Sprintf("Failed to save SMS log: %v", logErr))
	}

	return err
}
