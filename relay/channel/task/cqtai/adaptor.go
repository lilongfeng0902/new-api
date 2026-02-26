package cqtai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/constant"
	"github.com/QuantumNous/new-api/dto"
	"github.com/QuantumNous/new-api/relay/channel"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
	"github.com/QuantumNous/new-api/service"

	"github.com/gin-gonic/gin"
)

// CqtaiResponse Cqtai API 响应格式
type CqtaiResponse struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data any    `json:"data"`
}

// IsSuccess 判断是否成功 (HTTP 200 并且 code 为 200)
func (r *CqtaiResponse) IsSuccess() bool {
	return r.Code == 200
}

type TaskAdaptor struct {
	ChannelType int
}

func (a *TaskAdaptor) ParseTaskResult([]byte) (*relaycommon.TaskInfo, error) {
	return nil, fmt.Errorf("not implement") // todo implement this method if needed
}

func (a *TaskAdaptor) Init(info *relaycommon.RelayInfo) {
	a.ChannelType = info.ChannelType
}

func (a *TaskAdaptor) ValidateRequestAndSetAction(c *gin.Context, info *relaycommon.RelayInfo) (taskErr *dto.TaskError) {
	// 只对 POST 请求解析 body，GET 请求不需要 body
	if c.Request.Method == http.MethodPost {
		var requestBody map[string]any
		err := common.UnmarshalBodyReusable(c, &requestBody)
		if err != nil {
			taskErr = service.TaskErrorWrapperLocal(err, "invalid_request", http.StatusBadRequest)
			return
		}
		c.Set("task_request", requestBody)

		// 根据路径和请求体中的task字段设置action
		if strings.Contains(c.Request.URL.Path, "/generator/suno") {
			// 检查task字段：lyrics或music
			if taskType, ok := requestBody["task"].(string); ok {
				if taskType == "lyrics" {
					info.Action = constant.CqtaiActionLyrics
				} else {
					info.Action = constant.CqtaiActionMusic
				}
			} else {
				info.Action = constant.CqtaiActionMusic
			}
		} else {
			info.Action = constant.CqtaiActionFetch
		}
	} else {
		// GET 请求默认为FETCH
		info.Action = constant.CqtaiActionFetch
	}

	return nil
}

func (a *TaskAdaptor) BuildRequestURL(info *relaycommon.RelayInfo) (string, error) {
	// 路径转换：前端隐藏 cqt 标识，转发到上游时加上
	// 客户端请求 GET /api/v2/sunoinfo?id=123&id=456
	// 转发到: https://api.cqtai.com/api/cqt/v2/sunoinfo?id=123&id=456
	baseURL := info.ChannelBaseUrl
	requestPath := info.RequestURLPath

	// 路径转换：将 /api/generator/suno 和 /api/v2/sunoinfo 转换为带 /api/cqt/ 前缀的路径
	// 注意：RequestURLPath 可能包含 query 参数，如 /api/v2/sunoinfo?id=123
	if strings.HasPrefix(requestPath, "/api/generator/suno") {
		requestPath = strings.Replace(requestPath, "/api/generator/suno", "/api/cqt/generator/suno", 1)
	} else if strings.HasPrefix(requestPath, "/api/v2/sunoinfo") {
		requestPath = strings.Replace(requestPath, "/api/v2/sunoinfo", "/api/cqt/v2/sunoinfo", 1)
	}

	// RequestURLPath 已经包含了query参数
	fullRequestURL := fmt.Sprintf("%s%s", baseURL, requestPath)

	return fullRequestURL, nil
}

func (a *TaskAdaptor) BuildRequestHeader(c *gin.Context, req *http.Request, info *relaycommon.RelayInfo) error {
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", c.Request.Header.Get("Accept"))
	// Cqtai API 使用 Bearer token 认证方式
	req.Header.Set("Authorization", "Bearer "+info.ApiKey)
	return nil
}

func (a *TaskAdaptor) BuildRequestBody(c *gin.Context, info *relaycommon.RelayInfo) (io.Reader, error) {
	// GET 请求不需要 body
	if c.Request.Method == http.MethodGet {
		return nil, nil
	}

	// 直接透传原始请求体，不做任何格式转换
	requestBody, ok := c.Get("task_request")
	if !ok {
		var body map[string]any
		err := common.UnmarshalBodyReusable(c, &body)
		if err != nil {
			return nil, err
		}
		requestBody = body
	}

	data, err := json.Marshal(requestBody)
	if err != nil {
		return nil, err
	}
	return bytes.NewReader(data), nil
}

func (a *TaskAdaptor) DoRequest(c *gin.Context, info *relaycommon.RelayInfo, requestBody io.Reader) (*http.Response, error) {
	// 添加调试日志
	if common.DebugEnabled {
		common.SysLog(fmt.Sprintf("[Cqtai Debug] DoRequest: method=%s, url=%s, hasBody=%v",
			c.Request.Method, c.Request.URL.Path, requestBody != nil))
	}
	return channel.DoTaskApiRequest(a, c, info, requestBody)
}

func (a *TaskAdaptor) DoResponse(c *gin.Context, resp *http.Response, info *relaycommon.RelayInfo) (taskID string, taskData []byte, taskErr *dto.TaskError) {
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		taskErr = service.TaskErrorWrapper(err, "read_response_body_failed", http.StatusInternalServerError)
		return
	}

	// 解析 Cqtai API 响应格式
	var cqtaiResponse CqtaiResponse
	err = json.Unmarshal(responseBody, &cqtaiResponse)
	if err != nil {
		taskErr = service.TaskErrorWrapper(err, "unmarshal_response_body_failed", http.StatusInternalServerError)
		return
	}

	// 检查响应是否成功
	if !cqtaiResponse.IsSuccess() {
		taskErr = service.TaskErrorWrapper(
			fmt.Errorf("%s", cqtaiResponse.Msg),
			"fail_to_fetch_task",
			http.StatusInternalServerError,
		)
		return
	}

	// 复制响应头
	for k, v := range resp.Header {
		c.Writer.Header().Set(k, v[0])
	}
	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.WriteHeader(resp.StatusCode)

	// 返回原始响应体给客户端
	_, err = io.Copy(c.Writer, bytes.NewBuffer(responseBody))
	if err != nil {
		taskErr = service.TaskErrorWrapper(err, "copy_response_body_failed", http.StatusInternalServerError)
		return
	}

	// 提取 task_id（如果 data 是字符串类型）
	if dataStr, ok := cqtaiResponse.Data.(string); ok {
		return dataStr, nil, nil
	}

	// 如果 data 不是字符串，返回 JSON 编码的 data
	dataBytes, _ := json.Marshal(cqtaiResponse.Data)
	return string(dataBytes), nil, nil
}

func (a *TaskAdaptor) GetModelList() []string {
	return ModelList
}

func (a *TaskAdaptor) GetChannelName() string {
	return ChannelName
}

func (a *TaskAdaptor) FetchTask(baseUrl, key string, body map[string]any, proxy string) (*http.Response, error) {
	// Cqtai的查询接口是 GET 类型：/api/cqt/v2/sunoinfo
	// 参数：id (单个参数，可以多次使用查询多个任务)
	// 示例：/api/cqt/v2/sunoinfo?id=f5aaff05ad134cc3b7b4a51944420edb
	requestUrl := fmt.Sprintf("%s/api/cqt/v2/sunoinfo", baseUrl)

	// 将body中的ids转换为query参数（使用id参数名）
	if idsInterface, ok := body["ids"]; ok {
		if ids, ok := idsInterface.([]string); ok && len(ids) > 0 {
			// 使用url.Values构建query参数，每个id作为单独的参数
			params := url.Values{}
			for _, id := range ids {
				params.Add("id", id)
			}
			requestUrl += "?" + params.Encode()
		}
	}

	req, err := http.NewRequest("GET", requestUrl, nil)
	if err != nil {
		common.SysLog(fmt.Sprintf("Get Task error: %v", err))
		return nil, err
	}

	// 设置超时时间
	timeout := time.Second * 15
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	// 使用带有超时的 context 创建新的请求
	req = req.WithContext(ctx)
	// Cqtai API 使用 Bearer token 认证方式
	req.Header.Set("Authorization", "Bearer "+key)
	client, err := service.GetHttpClientWithProxy(proxy)
	if err != nil {
		return nil, fmt.Errorf("new proxy http client failed: %w", err)
	}
	return client.Do(req)
}
