package relay

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"strings"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/dto"
	"github.com/QuantumNous/new-api/logger"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/relay/channel/task/cqtai"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
	"github.com/QuantumNous/new-api/service"
	"github.com/QuantumNous/new-api/setting/ratio_setting"
	"github.com/QuantumNous/new-api/types"

	"github.com/gin-gonic/gin"
)

// validateRequestAndSetAction 验证请求并设置action，避免在adaptor初始化前访问嵌套字段
func validateRequestAndSetAction(c *gin.Context, info *relaycommon.RelayInfo) *dto.TaskError {
	// 确保 TaskRelayInfo 已初始化
	if info.TaskRelayInfo == nil {
		info.TaskRelayInfo = &relaycommon.TaskRelayInfo{}
	}

	// 根据路径设置action
	if strings.Contains(c.Request.URL.Path, "/generator/suno") {
		info.TaskRelayInfo.Action = "MUSIC"
	} else {
		info.TaskRelayInfo.Action = "FETCH"
	}

	// 只对 POST 请求解析 body，GET 请求不需要 body
	if c.Request.Method == http.MethodPost {
		var requestBody map[string]any
		err := common.UnmarshalBodyReusable(c, &requestBody)
		if err != nil {
			return service.TaskErrorWrapperLocal(err, "invalid_request", http.StatusBadRequest)
		}
		c.Set("task_request", requestBody)
	}

	return nil
}

// CqtaiProxyHandler 直接转发 Cqtai API 请求到上游，不经过任务系统
func CqtaiProxyHandler(c *gin.Context, info *relaycommon.RelayInfo) (newAPIError *types.NewAPIError) {
	// 确保信息已正确初始化
	if info == nil {
		logger.LogError(c, "[Cqtai Proxy] relay info is nil")
		return types.NewError(
			fmt.Errorf("relay info is nil"),
			types.ErrorCodeGenRelayInfoFailed,
		)
	}

	info.InitChannelMeta(c)

	// 检查 ChannelMeta 是否正确初始化
	if info.ChannelMeta == nil {
		logger.LogError(c, "[Cqtai Proxy] ChannelMeta is nil after InitChannelMeta")
		return types.NewError(
			fmt.Errorf("channel metadata not initialized"),
			types.ErrorCodeGetChannelFailed,
		)
	}

	// 检查必要的字段
	if info.ChannelBaseUrl == "" {
		logger.LogError(c, "[Cqtai Proxy] ChannelBaseUrl is empty")
		return types.NewError(
			fmt.Errorf("channel base URL is empty"),
			types.ErrorCodeInvalidRequest,
		)
	}

	if info.ApiKey == "" {
		logger.LogError(c, "[Cqtai Proxy] ApiKey is empty")
		return types.NewError(
			fmt.Errorf("API key is empty"),
			types.ErrorCodeChannelInvalidKey,
		)
	}

	// 验证请求（不需要初始化adaptor）
	taskErr := validateRequestAndSetAction(c, info)
	if taskErr != nil {
		return types.NewError(
			taskErr.Error,
			types.ErrorCodeConvertRequestFailed,
			types.ErrOptionWithStatusCode(taskErr.StatusCode),
		)
	}

	// 创建 cqtai adaptor
	adaptor := &cqtai.TaskAdaptor{}
	adaptor.Init(info)
	if taskErr != nil {
		return types.NewError(
			taskErr.Error,
			types.ErrorCodeConvertRequestFailed,
			types.ErrOptionWithStatusCode(taskErr.StatusCode),
		)
	}

	// 构建请求URL
	requestURL, err := adaptor.BuildRequestURL(info)
	if err != nil {
		return types.NewError(err, types.ErrorCodeInvalidRequest)
	}

	// 构建请求体
	var requestBody io.Reader
	if c.Request.Method == http.MethodPost {
		requestBody, err = adaptor.BuildRequestBody(c, info)
		if err != nil {
			return types.NewError(err, types.ErrorCodeBadRequestBody)
		}
	}

	// 记录请求日志（调试模式）
	logger.LogInfo(c, fmt.Sprintf("[Cqtai Proxy] %s %s", c.Request.Method, requestURL))

	// 执行请求
	resp, err := adaptor.DoRequest(c, info, requestBody)
	if err != nil {
		logger.LogError(c, fmt.Sprintf("do request failed: %v", err))
		return types.NewError(err, types.ErrorCodeDoRequestFailed)
	}
	defer resp.Body.Close()

	// 读取响应体
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.LogError(c, fmt.Sprintf("read response body failed: %v", err))
		return types.NewError(err, types.ErrorCodeReadResponseBodyFailed)
	}

	// 复制响应头
	for k, v := range resp.Header {
		if len(v) > 0 {
			c.Writer.Header().Set(k, v[0])
		}
	}
	c.Writer.Header().Set("Content-Type", "application/json")

	// 检查响应状态码
	if resp.StatusCode != http.StatusOK {
		// 尝试解析 cqtai 错误响应
		var cqtaiResp cqtai.CqtaiResponse
		if jsonErr := json.Unmarshal(responseBody, &cqtaiResp); jsonErr == nil && !cqtaiResp.IsSuccess() {
			logger.LogError(c, fmt.Sprintf("cqtai api error: %s", cqtaiResp.Msg))
			err := types.NewError(
				fmt.Errorf("cqtai api error: %s", cqtaiResp.Msg),
				types.ErrorCodeBadResponseStatusCode,
				types.ErrOptionWithStatusCode(resp.StatusCode),
			)
			return err
		}
		logger.LogError(c, fmt.Sprintf("upstream error: %s", string(responseBody)))
		err := types.NewError(
			fmt.Errorf("upstream error: %s", string(responseBody)),
			types.ErrorCodeBadResponseStatusCode,
			types.ErrOptionWithStatusCode(resp.StatusCode),
		)
		return err
	}

	// 直接返回响应
	c.Writer.WriteHeader(resp.StatusCode)
	_, err = io.Copy(c.Writer, bytes.NewBuffer(responseBody))
	if err != nil {
		logger.LogError(c, fmt.Sprintf("failed to write response: %v", err))
		return types.NewError(err, types.ErrorCodeBadResponse)
	}

	// 记录消费（仅对查询接口，使用 suno_fetch 模型）
	if strings.Contains(c.Request.URL.Path, "/v2/sunoinfo") {
		recordCqtaiFetchConsumption(c, info)
	}

	return nil
}

// recordCqtaiFetchConsumption 记录 cqtai 查询接口的消费
func recordCqtaiFetchConsumption(c *gin.Context, info *relaycommon.RelayInfo) {
	modelName := "suno_fetch"
	modelPrice, success := ratio_setting.GetModelPrice(modelName, true)
	if !success || math.IsNaN(modelPrice) || modelPrice <= 0 {
		defaultPrice, ok := ratio_setting.GetDefaultModelPriceMap()[modelName]
		if !ok || math.IsNaN(defaultPrice) || defaultPrice <= 0 {
			modelPrice = 0.01 // 默认价格
		} else {
			modelPrice = defaultPrice
		}
	}

	groupRatio := ratio_setting.GetGroupRatio(info.UsingGroup)
	if math.IsNaN(groupRatio) || groupRatio <= 0 {
		groupRatio = 1.0 // 默认分组倍率
	}

	var ratio float64
	userGroupRatio, hasUserGroupRatio := ratio_setting.GetGroupGroupRatio(info.UserGroup, info.UsingGroup)
	if hasUserGroupRatio && !math.IsNaN(userGroupRatio) && userGroupRatio > 0 {
		ratio = modelPrice * userGroupRatio
	} else {
		ratio = modelPrice * groupRatio
	}

	// 最后的保护性检查
	if math.IsNaN(ratio) || ratio <= 0 {
		common.SysLog(fmt.Sprintf("[Cqtai Fetch] Invalid ratio calculation: modelPrice=%f, groupRatio=%f, userGroupRatio=%f, usingGroup=%s, userGroup=%s",
			modelPrice, groupRatio, userGroupRatio, info.UsingGroup, info.UserGroup))
		return
	}

	quota := int(ratio * common.QuotaPerUnit)
	if quota <= 0 {
		common.SysLog(fmt.Sprintf("[Cqtai Fetch] Invalid quota: %d, ratio=%f", quota, ratio))
		return
	}

	err := service.PostConsumeQuota(info, quota, 0, true)
	if err != nil {
		common.SysLog("error consuming token remain quota: " + err.Error())
		return
	}

	tokenName := c.GetString("token_name")
	action := "FETCH"
	if info.TaskRelayInfo != nil && info.TaskRelayInfo.Action != "" {
		action = info.TaskRelayInfo.Action
	}
	logContent := fmt.Sprintf("操作 %s (查询)", action)

	other := make(map[string]interface{})
	if c != nil && c.Request != nil && c.Request.URL != nil {
		// 精简路径，去掉 /api/cqt 前缀
		requestPath := c.Request.URL.Path
		if strings.HasPrefix(requestPath, "/api/cqt") {
			requestPath = strings.TrimPrefix(requestPath, "/api/cqt")
		}
		other["request_path"] = requestPath
	}
	other["model_name"] = modelName

	// 确保记录的值都是有效的
	displayModelPrice := modelPrice
	if math.IsNaN(displayModelPrice) {
		displayModelPrice = 0.01
	}
	displayRatio := ratio
	if math.IsNaN(displayRatio) {
		displayRatio = 1.0
	}
	displayGroupRatio := groupRatio
	if math.IsNaN(displayGroupRatio) {
		displayGroupRatio = 1.0
	}

	other["model_price"] = displayModelPrice
	other["ratio"] = displayRatio
	other["group_ratio"] = displayGroupRatio
	other["using_group"] = info.UsingGroup
	other["user_group"] = info.UserGroup

	model.RecordConsumeLog(c, info.UserId, model.RecordConsumeLogParams{
		ChannelId: info.ChannelId,
		ModelName: modelName,
		TokenName: tokenName,
		Quota:     quota,
		Content:   logContent,
		TokenId:   info.TokenId,
		Group:     info.UsingGroup,
		Other:     other,
	})

	model.UpdateUserUsedQuotaAndRequestCount(info.UserId, quota)
	model.UpdateChannelUsedQuota(info.ChannelId, quota)
}
