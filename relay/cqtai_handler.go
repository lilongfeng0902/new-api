package relay

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/dto"
	"github.com/QuantumNous/new-api/logger"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
	"github.com/QuantumNous/new-api/relay/channel/task/cqtai"
	"github.com/QuantumNous/new-api/service"
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
		return types.NewErrorWithStatusCode(
			fmt.Errorf("relay info is nil"),
			"",
			http.StatusInternalServerError,
		)
	}

	logger.LogInfo(c, "[Cqtai Proxy] calling InitChannelMeta")
	info.InitChannelMeta(c)

	logger.LogInfo(c, fmt.Sprintf("[Cqtai Proxy] after InitChannelMeta - ChannelMeta nil: %v", info.ChannelMeta == nil))

	// 检查 ChannelMeta 是否正确初始化
	if info.ChannelMeta == nil {
		logger.LogError(c, "[Cqtai Proxy] ChannelMeta is nil after InitChannelMeta")
		return types.NewErrorWithStatusCode(
			fmt.Errorf("channel metadata not initialized"),
			"",
			http.StatusInternalServerError,
		)
	}

	logger.LogInfo(c, fmt.Sprintf("[Cqtai Proxy] ChannelBaseUrl: %q, ApiKey: %q, ChannelType: %d",
		info.ChannelBaseUrl, info.ApiKey, info.ChannelType))

	// 检查必要的字段
	if info.ChannelBaseUrl == "" {
		logger.LogError(c, "[Cqtai Proxy] ChannelBaseUrl is empty")
		return types.NewErrorWithStatusCode(
			fmt.Errorf("channel base URL is empty"),
			"",
			http.StatusInternalServerError,
		)
	}

	if info.ApiKey == "" {
		logger.LogError(c, "[Cqtai Proxy] ApiKey is empty")
		return types.NewErrorWithStatusCode(
			fmt.Errorf("API key is empty"),
			"",
			http.StatusInternalServerError,
		)
	}

	// 验证请求（不需要初始化adaptor）
	taskErr := validateRequestAndSetAction(c, info)
	if taskErr != nil {
		return types.NewErrorWithStatusCode(
			taskErr.Error,
			"",
			taskErr.StatusCode,
		)
	}

	// 创建 cqtai adaptor
	adaptor := &cqtai.TaskAdaptor{}
	adaptor.Init(info)
	if taskErr != nil {
		return types.NewErrorWithStatusCode(
			taskErr.Error,
			"",
			taskErr.StatusCode,
		)
	}

	// 构建请求URL
	requestURL, err := adaptor.BuildRequestURL(info)
	if err != nil {
		return types.NewErrorWithStatusCode(err, "", http.StatusInternalServerError)
	}

	// 构建请求体
	var requestBody io.Reader
	if c.Request.Method == http.MethodPost {
		requestBody, err = adaptor.BuildRequestBody(c, info)
		if err != nil {
			return types.NewErrorWithStatusCode(err, "", http.StatusBadRequest)
		}
	}

	// 记录请求日志（调试模式）
	logger.LogInfo(c, fmt.Sprintf("[Cqtai Proxy] %s %s", c.Request.Method, requestURL))

	// 执行请求
	resp, err := adaptor.DoRequest(c, info, requestBody)
	if err != nil {
		logger.LogError(c, fmt.Sprintf("do request failed: %v", err))
		return types.NewErrorWithStatusCode(err, "", http.StatusInternalServerError)
	}
	defer resp.Body.Close()

	// 读取响应体
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.LogError(c, fmt.Sprintf("read response body failed: %v", err))
		return types.NewErrorWithStatusCode(err, "", http.StatusInternalServerError)
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
			return types.NewErrorWithStatusCode(
				fmt.Errorf("cqtai api error: %s", cqtaiResp.Msg),
				"",
				resp.StatusCode,
			)
		}
		logger.LogError(c, fmt.Sprintf("upstream error: %s", string(responseBody)))
		return types.NewErrorWithStatusCode(
			fmt.Errorf("upstream error: %s", string(responseBody)),
			"",
			resp.StatusCode,
		)
	}

	// 直接返回响应
	c.Writer.WriteHeader(resp.StatusCode)
	_, err = io.Copy(c.Writer, bytes.NewBuffer(responseBody))
	if err != nil {
		logger.LogError(c, fmt.Sprintf("failed to write response: %v", err))
		return types.NewErrorWithStatusCode(err, "", http.StatusInternalServerError)
	}

	return nil
}
