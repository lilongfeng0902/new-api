package model

import (
	"gorm.io/gorm"
)

// SmsLog SMS sending log record
type SmsLog struct {
	Id             int            `json:"id" gorm:"primaryKey;autoIncrement"`
	TemplateCode   string         `json:"template_code" gorm:"type:varchar(50)"`
	Mobile         string         `json:"mobile" gorm:"type:varchar(16);not null;index"`
	UserId         *int           `json:"user_id" gorm:"index"`
	SendStatus     int            `json:"send_status"`
	SendTime       int64          `json:"send_time" gorm:"bigint"`
	ApiSendCode    string         `json:"api_send_code" gorm:"type:varchar(64)"`
	ApiSendMsg     string         `json:"api_send_msg" gorm:"type:varchar(255)"`
	ApiRequestId   string         `json:"api_request_id" gorm:"type:varchar(128)"`
	ApiSerialNo    string         `json:"api_serial_no" gorm:"type:varchar(128)"` // BizId
	CreatedAt      int64          `json:"created_at" gorm:"bigint;not null"`
	UpdatedAt      int64          `json:"updated_at" gorm:"bigint;not null"`
	DeletedAt      gorm.DeletedAt `json:"-" gorm:"index"`
}

// TableName specifies the table name
func (SmsLog) TableName() string {
	return "sms_logs"
}

// SMS log status constants
const (
	SmsLogStatusPending = 0 // Pending
	SmsLogStatusSuccess = 1 // Success
	SmsLogStatusFailed  = 2 // Failed
)
