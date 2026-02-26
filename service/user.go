package service

import (
	"errors"
	"fmt"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"gorm.io/gorm"
)

// UserService user-related service
type UserService struct{}

// NewUserService creates a new user service
func NewUserService() *UserService {
	return &UserService{}
}

// GetUserByMobile retrieves user by mobile number
func (s *UserService) GetUserByMobile(mobile string) (*model.User, error) {
	var user model.User
	err := model.DB.Where("mobile = ?", mobile).First(&user).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New("用户不存在")
		}
		return nil, err
	}
	return &user, nil
}

// CreateUserWithMobile creates a new user with mobile number
func (s *UserService) CreateUserWithMobile(mobile, clientIp string) (*model.User, error) {
	// Generate username from mobile (can be changed by user later)
	username := "user_" + mobile[len(mobile)-6:]

	user := &model.User{
		Username:    username,
		Mobile:      mobile,
		DisplayName: username,
		Role:        common.RoleCommonUser,
		Status:      common.UserStatusEnabled,
		Group:       "default",
	}

	// Use Insert method which handles quota, aff_code, etc.
	err := user.Insert(0) // inviterId = 0 for SMS auto-register
	if err != nil {
		return nil, fmt.Errorf("创建用户失败: %v", err)
	}

	// Log the auto-registration
	model.RecordLog(user.Id, model.LogTypeSystem,
		fmt.Sprintf("通过手机号 %s 自动注册 (IP: %s)", mobile, clientIp))

	return user, nil
}

// GetOrCreateUserByMobile gets existing user or creates new one
func (s *UserService) GetOrCreateUserByMobile(mobile, clientIp string, autoRegister bool) (*model.User, bool, error) {
	user, err := s.GetUserByMobile(mobile)

	if err == nil {
		// User exists
		return user, false, nil
	}

	// User doesn't exist
	if !autoRegister {
		return nil, false, errors.New("用户不存在,请先注册")
	}

	// Auto-register
	user, err = s.CreateUserWithMobile(mobile, clientIp)
	if err != nil {
		return nil, false, err
	}

	return user, true, nil
}

// CheckMobileAvailable checks if mobile number is available (not bound)
func (s *UserService) CheckMobileAvailable(mobile string) (bool, error) {
	var count int64
	err := model.DB.Model(&model.User{}).Where("mobile = ?", mobile).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count == 0, nil
}

// BindMobileToUser binds mobile number to existing user
func (s *UserService) BindMobileToUser(userId int, mobile string) error {
	// Check if mobile is already bound to another user
	available, err := s.CheckMobileAvailable(mobile)
	if err != nil {
		return err
	}
	if !available {
		return errors.New("该手机号已被其他用户绑定")
	}

	// Update user's mobile
	err = model.DB.Model(&model.User{}).Where("id = ?", userId).Update("mobile", mobile).Error
	if err != nil {
		return fmt.Errorf("绑定失败: %v", err)
	}

	// Log the binding
	model.RecordLog(userId, model.LogTypeSystem,
		fmt.Sprintf("绑定手机号 %s", mobile))

	return nil
}

// UnbindMobile unbinds mobile number from user
func (s *UserService) UnbindMobile(userId int) error {
	// Update user's mobile to empty
	err := model.DB.Model(&model.User{}).Where("id = ?", userId).Update("mobile", "").Error
	if err != nil {
		return fmt.Errorf("解绑失败: %v", err)
	}

	// Log the unbinding
	model.RecordLog(userId, model.LogTypeSystem, "解绑手机号")

	return nil
}

// ResetPasswordByMobile resets user's password by mobile number
func (s *UserService) ResetPasswordByMobile(mobile, newPassword string) error {
	// Get user by mobile
	user, err := s.GetUserByMobile(mobile)
	if err != nil {
		return err
	}

	// Hash password
	hashedPassword, err := common.Password2Hash(newPassword)
	if err != nil {
		return fmt.Errorf("密码加密失败: %v", err)
	}

	// Update password
	err = model.DB.Model(&model.User{}).Where("id = ?", user.Id).Update("password", hashedPassword).Error
	if err != nil {
		return fmt.Errorf("密码重置失败: %v", err)
	}

	// Log password reset
	model.RecordLog(user.Id, model.LogTypeSystem,
		fmt.Sprintf("通过手机号 %s 重置密码", mobile))

	return nil
}
