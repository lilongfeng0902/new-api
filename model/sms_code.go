package model

import (
	"gorm.io/gorm"
)

// SmsCode SMS verification code record
type SmsCode struct {
	Id         int            `json:"id" gorm:"primaryKey;autoIncrement"`
	Mobile     string         `json:"mobile" gorm:"type:varchar(16);not null;index:idx_mobile"`
	Code       string         `json:"code" gorm:"type:varchar(10);not null"`
	Scene      int            `json:"scene" gorm:"not null;index:idx_mobile_scene_code,priority:2"`
	CreateIp   string         `json:"create_ip" gorm:"type:varchar(45)"`
	TodayIndex int            `json:"today_index" gorm:"default:1"`
	Used       bool           `json:"used" gorm:"default:false"`
	UsedTime   int64          `json:"used_time" gorm:"bigint"`
	UsedIp     string         `json:"used_ip" gorm:"type:varchar(45)"`
	CreatedAt  int64          `json:"created_at" gorm:"bigint;not null;index"`
	UpdatedAt  int64          `json:"updated_at" gorm:"bigint;not null"`
	DeletedAt  gorm.DeletedAt `json:"-" gorm:"index"`
}

// TableName specifies the table name
func (SmsCode) TableName() string {
	return "sms_codes"
}

// SMS scene constants
const (
	SmsSceneLogin = 1 // Login/Register
	SmsSceneBind  = 2 // Bind mobile number
	SmsSceneReset = 3 // Password reset
)
