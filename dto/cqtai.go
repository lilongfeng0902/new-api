package dto

// CqtaiResponse Cqtai API 通用响应格式
type CqtaiResponse[T any] struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data T      `json:"data"`
}

// IsSuccess 判断是否成功
func (r *CqtaiResponse[T]) IsSuccess() bool {
	return r.Code == 200
}

// CqtaiTaskData Cqtai 查询任务返回的数据结构
type CqtaiTaskData struct {
	ID           string                 `json:"id"`
	TaskType     string                 `json:"taskType"`
	Status       string                 `json:"status"`
	ErrorMsg     string                 `json:"errorMsg"`
	Param        map[string]interface{} `json:"param"`
	Result       []CqtaiTaskResult      `json:"result"`
	CreateTime   string                 `json:"createTime"`
	CompleteTime string                 `json:"completeTime"`
}

// CqtaiTaskResult Cqtai 任务结果
type CqtaiTaskResult struct {
	Tags         []string `json:"tags"`
	Text         string   `json:"text"`
	Title        string   `json:"title"`
	Status       string   `json:"status"`
	ErrorMessage string   `json:"error_message"`
	AudioURL     string   `json:"audio_url,omitempty"`
	VideoURL     string   `json:"video_url,omitempty"`
	ImageURL     string   `json:"image_url,omitempty"`
}
