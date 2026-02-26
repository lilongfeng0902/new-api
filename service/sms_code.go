package service

import (
	"context"
	"errors"
	"fmt"
	"math/rand"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
)

// SmsCodeService SMS verification code service
type SmsCodeService struct {
	config *SmsCodeConfig
}

// SmsCodeConfig configuration for SMS code service
type SmsCodeConfig struct {
	CodeLength           int // Code length (4 or 6)
	CodeMin              int // Minimum code value (when min=max, use fixed code for testing)
	CodeMax              int // Maximum code value
	ExpireMinutes        int // Expiration time in minutes
	SendFrequencySeconds int // Send frequency limit in seconds
	MaxSendPerDay        int // Maximum sends per day (0 = unlimited)
}

// NewSmsCodeService creates a new SMS code service
func NewSmsCodeService(config *SmsCodeConfig) *SmsCodeService {
	if config == nil {
		config = &SmsCodeConfig{
			CodeLength:           6,
			CodeMin:              100000,
			CodeMax:              999999,
			ExpireMinutes:        5,
			SendFrequencySeconds: 60,
			MaxSendPerDay:        10,
		}
	}
	return &SmsCodeService{config: config}
}

// GenerateCode generates a random numeric code
func (s *SmsCodeService) GenerateCode() string {
	length := s.config.CodeLength
	if length != 4 && length != 6 {
		length = 6 // default to 6 digits
	}

	minCode := s.config.CodeMin
	maxCode := s.config.CodeMax

	// When min == max, use fixed code for testing
	if minCode == maxCode {
		format := fmt.Sprintf("%%0%dd", length)
		return fmt.Sprintf(format, minCode)
	}

	// Ensure min/max are valid for the code length
	maxAllowed := 1
	for i := 0; i < length; i++ {
		maxAllowed *= 10
	}
	maxAllowed-- // e.g., for 6 digits: 999999

	if minCode < 0 {
		minCode = 0
	}
	if maxCode > maxAllowed {
		maxCode = maxAllowed
	}
	if minCode > maxCode {
		minCode, maxCode = maxCode, minCode
	}

	// Generate random code in range [minCode, maxCode]
	code := rand.Intn(maxCode-minCode+1) + minCode
	format := fmt.Sprintf("%%0%dd", length)
	return fmt.Sprintf(format, code)
}

// CheckSendFrequency validates sending frequency limits
func (s *SmsCodeService) CheckSendFrequency(ctx context.Context, mobile string) error {
	var lastCode model.SmsCode
	err := model.DB.Where("mobile = ?", mobile).
		Order("created_at DESC").
		First(&lastCode).Error

	if err == nil {
		// Check 60-second frequency limit
		currentTime := common.GetTimestamp()
		timeSince := currentTime - lastCode.CreatedAt
		if timeSince < int64(s.config.SendFrequencySeconds) {
			remaining := int64(s.config.SendFrequencySeconds) - timeSince
			return fmt.Errorf("发送过于频繁,请 %d 秒后再试", remaining)
		}

		// Check daily limit
		if s.config.MaxSendPerDay > 0 {
			today := common.GetTimestamp() - (common.GetTimestamp() % 86400) // Start of today in seconds
			if lastCode.CreatedAt >= today && lastCode.TodayIndex >= s.config.MaxSendPerDay {
				return errors.New("今日发送次数已达上限")
			}
		}
	}

	return nil
}

// CalculateTodayIndex calculates today's send index for this mobile
func (s *SmsCodeService) CalculateTodayIndex(ctx context.Context, mobile string) int {
	var lastCode model.SmsCode
	err := model.DB.Where("mobile = ?", mobile).
		Order("created_at DESC").
		First(&lastCode).Error

	if err != nil {
		return 1
	}

	today := common.GetTimestamp() - (common.GetTimestamp() % 86400) // Start of today in seconds
	if lastCode.CreatedAt >= today {
		return lastCode.TodayIndex + 1
	}

	return 1
}

// CreateSmsCode creates a new SMS code record
func (s *SmsCodeService) CreateSmsCode(ctx context.Context, mobile string, scene int, clientIp string) (*model.SmsCode, error) {
	code := s.GenerateCode()
	todayIndex := s.CalculateTodayIndex(ctx, mobile)

	smsCode := &model.SmsCode{
		Mobile:     mobile,
		Code:       code,
		Scene:      scene,
		CreateIp:   clientIp,
		TodayIndex: todayIndex,
		Used:       false,
	}

	err := model.DB.Create(smsCode).Error
	if err != nil {
		return nil, err
	}

	return smsCode, nil
}

// ValidateAndUseSmsCode validates and marks code as used
func (s *SmsCodeService) ValidateAndUseSmsCode(ctx context.Context, mobile, code string, scene int, usedIp string) error {
	var smsCode model.SmsCode

	// Find the most recent unused code matching criteria
	err := model.DB.Where("mobile = ? AND code = ? AND scene = ? AND used = ?",
		mobile, code, scene, false).
		Order("created_at DESC").
		First(&smsCode).Error

	if err != nil {
		return errors.New("验证码不存在或已使用")
	}

	// Check expiration
	expireDuration := int64(s.config.ExpireMinutes) * 60 // Convert to seconds
	currentTime := common.GetTimestamp()
	if currentTime-smsCode.CreatedAt > expireDuration {
		return errors.New("验证码已过期")
	}

	// Mark as used
	now := common.GetTimestamp()
	smsCode.Used = true
	smsCode.UsedTime = now
	smsCode.UsedIp = usedIp

	err = model.DB.Save(&smsCode).Error
	if err != nil {
		return fmt.Errorf("标记验证码失败: %v", err)
	}

	return nil
}

// CleanupExpiredCodes removes old SMS codes (run as background task)
func (s *SmsCodeService) CleanupExpiredCodes(ctx context.Context, daysToKeep int) error {
	cutoffTime := common.GetTimestamp() - int64(daysToKeep)*86400 // daysToKeep days ago in seconds

	result := model.DB.Where("created_at < ?", cutoffTime).Delete(&model.SmsCode{})
	if result.Error != nil {
		return result.Error
	}

	common.SysLog(fmt.Sprintf("Cleaned up %d expired SMS codes", result.RowsAffected))
	return nil
}
