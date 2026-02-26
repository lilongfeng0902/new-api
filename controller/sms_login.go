package controller

import (
	"fmt"
	"net/http"
	"regexp"
	"strconv"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/service"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

// Request structs
type SendSmsCodeRequest struct {
	Mobile string `json:"mobile" binding:"required"`
}

type SmsLoginRequest struct {
	Mobile string `json:"mobile" binding:"required"`
	Code   string `json:"code" binding:"required"`
}

type BindMobileRequest struct {
	Mobile string `json:"mobile" binding:"required"`
	Code   string `json:"code" binding:"required"`
}

type ResetPasswordRequest struct {
	Mobile      string `json:"mobile" binding:"required"`
	Code        string `json:"code" binding:"required"`
	NewPassword string `json:"new_password" binding:"required"`
}

// ValidateMobile validates Chinese mobile number format
func ValidateMobile(mobile string) bool {
	// Chinese mobile: starts with 1, followed by 10 digits
	matched, _ := regexp.MatchString(`^1[3-9]\d{9}$`, mobile)
	return matched
}

// ValidatePassword validates password strength
func ValidatePassword(password string) bool {
	// Password must be 8-20 characters
	return len(password) >= 8 && len(password) <= 20
}

// SendSmsCode handles SMS verification code sending for login
func SendSmsCode(c *gin.Context) {
	// Check if SMS login is enabled
	if !getSmsLoginEnabled() {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "短信登录功能未开启",
		})
		return
	}

	var req SendSmsCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "无效的参数",
		})
		return
	}

	// Validate mobile format
	if !ValidateMobile(req.Mobile) {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "请输入有效的11位手机号码",
		})
		return
	}

	// Get client IP
	clientIp := c.ClientIP()

	// Initialize services
	smsCodeConfig := getSmsCodeConfig()
	smsCodeService := service.NewSmsCodeService(smsCodeConfig)
	smsSendService := service.NewSmsSendService()

	// Check sending frequency
	err := smsCodeService.CheckSendFrequency(c.Request.Context(), req.Mobile)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	// Create SMS code record
	smsCode, err := smsCodeService.CreateSmsCode(
		c.Request.Context(),
		req.Mobile,
		model.SmsSceneLogin,
		clientIp,
	)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "验证码生成失败",
		})
		common.SysLog(fmt.Sprintf("Failed to create SMS code: %v", err))
		return
	}

	// Get template code from config
	templateCode := getSmsLoginTemplateCode()
	if templateCode == "" {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "短信模板未配置",
		})
		return
	}

	// Send SMS
	err = smsSendService.SendVerificationCode(
		c.Request.Context(),
		req.Mobile,
		smsCode.Code,
		templateCode,
		nil, // userId is nil for login
	)

	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "短信发送失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "验证码已发送",
		"data": gin.H{
			"expire_minutes": smsCodeConfig.ExpireMinutes,
		},
	})
}

// SmsLogin handles SMS-based login
func SmsLogin(c *gin.Context) {
	// Check if SMS login is enabled
	if !getSmsLoginEnabled() {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "短信登录功能未开启",
		})
		return
	}

	var req SmsLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "无效的参数",
		})
		return
	}

	// Validate mobile format
	if !ValidateMobile(req.Mobile) {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "请输入有效的11位手机号码",
		})
		return
	}

	// Validate code format
	if len(req.Code) < 4 || len(req.Code) > 6 {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "请输入正确的验证码",
		})
		return
	}

	// Get client IP
	clientIp := c.ClientIP()

	// Initialize services
	smsCodeConfig := getSmsCodeConfig()
	smsCodeService := service.NewSmsCodeService(smsCodeConfig)
	userService := service.NewUserService()

	// Validate and use SMS code
	err := smsCodeService.ValidateAndUseSmsCode(
		c.Request.Context(),
		req.Mobile,
		req.Code,
		model.SmsSceneLogin,
		clientIp,
	)

	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	// Get or create user
	autoRegister := getSmsAutoRegisterEnabled()
	user, isNewUser, err := userService.GetOrCreateUserByMobile(req.Mobile, clientIp, autoRegister)

	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	// Check user status
	if user.Status != common.UserStatusEnabled {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "账号已被禁用",
		})
		return
	}

	// Setup login session (reuse existing function)
	setupLogin(user, c)

	// Log the login
	if isNewUser {
		model.RecordLog(user.Id, model.LogTypeSystem,
			fmt.Sprintf("通过短信验证码登录(首次) - 手机号: %s", req.Mobile))
	} else {
		model.RecordLog(user.Id, model.LogTypeSystem,
			fmt.Sprintf("通过短信验证码登录 - 手机号: %s", req.Mobile))
	}
}

// SendBindCode handles SMS verification code sending for mobile binding
func SendBindCode(c *gin.Context) {
	// Check if binding is enabled
	if !getSmsBindEnabled() {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "手机号绑定功能未开启",
		})
		return
	}

	// Get current user from session
	session := sessions.Default(c)
	userId := session.Get("id")
	if userId == nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "请先登录",
		})
		return
	}

	var req SendSmsCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "无效的参数",
		})
		return
	}

	// Validate mobile format
	if !ValidateMobile(req.Mobile) {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "请输入有效的11位手机号码",
		})
		return
	}

	// Get client IP
	clientIp := c.ClientIP()

	// Check if current user already has a mobile
	var currentUser model.User
	err := model.DB.First(&currentUser, userId).Error
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "用户不存在",
		})
		return
	}

	if currentUser.Mobile != "" {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "您已绑定手机号，请先解绑",
		})
		return
	}

	// Check if mobile is already bound to another user
	userService := service.NewUserService()
	available, err := userService.CheckMobileAvailable(req.Mobile)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "检查手机号失败",
		})
		return
	}

	if !available {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "该手机号已被其他用户绑定",
		})
		return
	}

	// Initialize services
	smsCodeConfig := getSmsCodeConfig()
	smsCodeService := service.NewSmsCodeService(smsCodeConfig)
	smsSendService := service.NewSmsSendService()

	// Check sending frequency
	err = smsCodeService.CheckSendFrequency(c.Request.Context(), req.Mobile)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	// Create SMS code record
	smsCode, err := smsCodeService.CreateSmsCode(
		c.Request.Context(),
		req.Mobile,
		model.SmsSceneBind,
		clientIp,
	)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "验证码生成失败",
		})
		return
	}

	// Get template code from config
	templateCode := getSmsBindTemplateCode()
	if templateCode == "" {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "短信模板未配置",
		})
		return
	}

	// Send SMS
	userIdInt := userId.(int)
	err = smsSendService.SendVerificationCode(
		c.Request.Context(),
		req.Mobile,
		smsCode.Code,
		templateCode,
		&userIdInt,
	)

	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "短信发送失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "验证码已发送",
		"data": gin.H{
			"expire_minutes": smsCodeConfig.ExpireMinutes,
		},
	})
}

// BindMobile handles mobile number binding
func BindMobile(c *gin.Context) {
	// Check if binding is enabled
	if !getSmsBindEnabled() {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "手机号绑定功能未开启",
		})
		return
	}

	// Get current user from session
	session := sessions.Default(c)
	userId := session.Get("id")
	if userId == nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "请先登录",
		})
		return
	}

	var req BindMobileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "无效的参数",
		})
		return
	}

	// Validate mobile format
	if !ValidateMobile(req.Mobile) {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "请输入有效的11位手机号码",
		})
		return
	}

	// Get client IP
	clientIp := c.ClientIP()

	// Initialize services
	smsCodeConfig := getSmsCodeConfig()
	smsCodeService := service.NewSmsCodeService(smsCodeConfig)
	userService := service.NewUserService()

	// Validate and use SMS code
	err := smsCodeService.ValidateAndUseSmsCode(
		c.Request.Context(),
		req.Mobile,
		req.Code,
		model.SmsSceneBind,
		clientIp,
	)

	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	// Bind mobile to user
	userIdInt := userId.(int)
	err = userService.BindMobileToUser(userIdInt, req.Mobile)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "手机号绑定成功",
		"data": gin.H{
			"mobile": req.Mobile,
		},
	})
}

// UnbindMobile handles mobile number unbinding
func UnbindMobile(c *gin.Context) {
	// Get current user from session
	session := sessions.Default(c)
	userId := session.Get("id")
	if userId == nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "请先登录",
		})
		return
	}

	// Check if user has a mobile bound
	var user model.User
	err := model.DB.First(&user, userId).Error
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "用户不存在",
		})
		return
	}

	if user.Mobile == "" {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "您尚未绑定手机号",
		})
		return
	}

	// Unbind mobile
	userService := service.NewUserService()
	userIdInt := userId.(int)
	err = userService.UnbindMobile(userIdInt)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "手机号解绑成功",
	})
}

// SendResetCode handles SMS verification code sending for password reset
func SendResetCode(c *gin.Context) {
	// Check if reset is enabled
	if !getSmsResetEnabled() {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "短信重置密码功能未开启",
		})
		return
	}

	var req SendSmsCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "无效的参数",
		})
		return
	}

	// Validate mobile format
	if !ValidateMobile(req.Mobile) {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "请输入有效的11位手机号码",
		})
		return
	}

	// Get client IP
	clientIp := c.ClientIP()

	// Check if mobile exists (bound to a user)
	userService := service.NewUserService()
	user, err := userService.GetUserByMobile(req.Mobile)
	if err != nil {
		// For security, always return success even if mobile doesn't exist
		// to prevent mobile number enumeration attacks
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "验证码已发送",
			"data": gin.H{
				"expire_minutes": 5,
			},
		})
		// Log the attempt for security monitoring
		common.SysLog(fmt.Sprintf("Password reset attempt for non-existent mobile: %s from IP: %s", req.Mobile, clientIp))
		return
	}

	// Initialize services
	smsCodeConfig := getSmsCodeConfig()
	smsCodeService := service.NewSmsCodeService(smsCodeConfig)
	smsSendService := service.NewSmsSendService()

	// Check sending frequency
	err = smsCodeService.CheckSendFrequency(c.Request.Context(), req.Mobile)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	// Create SMS code record
	smsCode, err := smsCodeService.CreateSmsCode(
		c.Request.Context(),
		req.Mobile,
		model.SmsSceneReset,
		clientIp,
	)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "验证码生成失败",
		})
		return
	}

	// Get template code from config
	templateCode := getSmsResetTemplateCode()
	if templateCode == "" {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "短信模板未配置",
		})
		return
	}

	// Send SMS
	err = smsSendService.SendVerificationCode(
		c.Request.Context(),
		req.Mobile,
		smsCode.Code,
		templateCode,
		&user.Id,
	)

	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "短信发送失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "验证码已发送",
		"data": gin.H{
			"expire_minutes": smsCodeConfig.ExpireMinutes,
		},
	})
}

// SmsResetPassword handles password reset via SMS
func SmsResetPassword(c *gin.Context) {
	// Check if reset is enabled
	if !getSmsResetEnabled() {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "短信重置密码功能未开启",
		})
		return
	}

	var req ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "无效的参数",
		})
		return
	}

	// Validate mobile format
	if !ValidateMobile(req.Mobile) {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "请输入有效的11位手机号码",
		})
		return
	}

	// Validate password strength
	if !ValidatePassword(req.NewPassword) {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "密码长度必须为8-20位",
		})
		return
	}

	// Get client IP
	clientIp := c.ClientIP()

	// Initialize services
	smsCodeConfig := getSmsCodeConfig()
	smsCodeService := service.NewSmsCodeService(smsCodeConfig)
	userService := service.NewUserService()

	// Validate and use SMS code
	err := smsCodeService.ValidateAndUseSmsCode(
		c.Request.Context(),
		req.Mobile,
		req.Code,
		model.SmsSceneReset,
		clientIp,
	)

	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	// Reset password
	err = userService.ResetPasswordByMobile(req.Mobile, req.NewPassword)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "密码重置成功，请重新登录",
	})
}

// Helper functions to get config values
func getSmsLoginEnabled() bool {
	value := common.OptionMap["SmsLoginEnabled"]
	return value == "true"
}

func getSmsAutoRegisterEnabled() bool {
	value := common.OptionMap["SmsAutoRegisterEnabled"]
	// Default to true
	return value == "true" || value == ""
}

func getSmsBindEnabled() bool {
	value := common.OptionMap["SmsBindEnabled"]
	// Default to true
	return value == "true" || value == ""
}

func getSmsResetEnabled() bool {
	value := common.OptionMap["SmsResetEnabled"]
	// Default to true
	return value == "true" || value == ""
}

func getSmsLoginTemplateCode() string {
	return common.OptionMap["SmsLoginTemplateCode"]
}

func getSmsBindTemplateCode() string {
	return common.OptionMap["SmsBindTemplateCode"]
}

func getSmsResetTemplateCode() string {
	return common.OptionMap["SmsResetTemplateCode"]
}

func getSmsCodeConfig() *service.SmsCodeConfig {
	config := &service.SmsCodeConfig{
		CodeLength:           6,
		CodeMin:              100000,
		CodeMax:              999999,
		ExpireMinutes:        5,
		SendFrequencySeconds: 60,
		MaxSendPerDay:        10,
	}

	// Load from options
	if val := common.OptionMap["SmsCodeLength"]; val != "" {
		if length, err := strconv.Atoi(val); err == nil {
			config.CodeLength = length
		}
	}
	if val := common.OptionMap["SmsCodeMin"]; val != "" {
		if min, err := strconv.Atoi(val); err == nil {
			config.CodeMin = min
		}
	}
	if val := common.OptionMap["SmsCodeMax"]; val != "" {
		if max, err := strconv.Atoi(val); err == nil {
			config.CodeMax = max
		}
	}
	if val := common.OptionMap["SmsCodeExpireMinutes"]; val != "" {
		if minutes, err := strconv.Atoi(val); err == nil {
			config.ExpireMinutes = minutes
		}
	}
	if val := common.OptionMap["SmsSendFrequencySeconds"]; val != "" {
		if seconds, err := strconv.Atoi(val); err == nil {
			config.SendFrequencySeconds = seconds
		}
	}
	if val := common.OptionMap["SmsMaxSendPerDay"]; val != "" {
		if max, err := strconv.Atoi(val); err == nil {
			config.MaxSendPerDay = max
		}
	}

	return config
}
