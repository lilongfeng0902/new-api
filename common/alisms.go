package common

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
)

type AliyunSMSRequest struct {
	PhoneNumbers  string            `json:"PhoneNumbers"`
	SignName      string            `json:"SignName"`
	TemplateCode  string            `json:"TemplateCode"`
	TemplateParam map[string]string `json:"TemplateParam,omitempty"`
}

type AliyunSMSResponse struct {
	RequestId string `json:"RequestId"`
	Code      string `json:"Code"`
	Message   string `json:"Message"`
	BizId     string `json:"BizId,omitempty"`
}

func percentEncode(str string) string {
	str = url.QueryEscape(str)
	str = strings.ReplaceAll(str, "+", "%20")
	str = strings.ReplaceAll(str, "*", "%2A")
	str = strings.ReplaceAll(str, "%7E", "~")
	return str
}

func signStringToString(method, canonicalizedQueryString, accessKeySecret string) string {
	stringToSign := method + "&" + percentEncode("/") + "&" + percentEncode(canonicalizedQueryString)
	h := hmac.New(sha1.New, []byte(accessKeySecret+"&"))
	h.Write([]byte(stringToSign))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

func buildCanonicalizedQueryString(params map[string]string) string {
	var keys []string
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var queryString string
	for _, k := range keys {
		if params[k] != "" {
			queryString += "&" + percentEncode(k) + "=" + percentEncode(params[k])
		}
	}
	if len(queryString) > 0 {
		queryString = queryString[1:] // Remove the leading '&'
	}
	return queryString
}

func SendAliyunSMS(phoneNumber string, templateParam map[string]string, templateCode string) error {
	if AliyunSMSEndpoint == "" || AliyunSMSAccessKeyId == "" || AliyunSMSAccessKeySecret == "" {
		return fmt.Errorf("阿里云短信配置未完成")
	}

	if AliyunSMSSignName == "" || AliyunSMSTemplateCode == "" {
		return fmt.Errorf("阿里云短信签名或模板未配置")
	}

	// 使用提供的模板代码或默认的告警模板代码
	actualTemplateCode := AliyunSMSTemplateCode
	if templateCode != "" {
		actualTemplateCode = templateCode
	}

	// 构建请求参数
	params := map[string]string{
		"AccessKeyId":      AliyunSMSAccessKeyId,
		"Action":           "SendSms",
		"Format":           "JSON",
		"PhoneNumbers":     phoneNumber,
		"RegionId":         "cn-hangzhou",
		"SignName":         AliyunSMSSignName,
		"SignatureMethod":  "HMAC-SHA1",
		"SignatureNonce":   uuid.New().String(),
		"SignatureVersion": "1.0",
		"TemplateCode":     actualTemplateCode,
		"Timestamp":        time.Now().UTC().Format("2006-01-02T15:04:05Z"),
		"Version":          "2017-05-25",
	}

	// 添加模板参数
	if len(templateParam) > 0 {
		templateParamJSON, err := json.Marshal(templateParam)
		if err != nil {
			return fmt.Errorf("序列化模板参数失败: %v", err)
		}
		params["TemplateParam"] = string(templateParamJSON)
	}

	// 构建规范化的查询字符串
	canonicalizedQueryString := buildCanonicalizedQueryString(params)

	// 生成签名
	signature := signStringToString("GET", canonicalizedQueryString, AliyunSMSAccessKeySecret)
	params["Signature"] = signature

	// 构建最终的查询字符串
	var finalParams []string
	for k, v := range params {
		finalParams = append(finalParams, percentEncode(k)+"="+percentEncode(v))
	}
	sort.Strings(finalParams)
	queryString := strings.Join(finalParams, "&")

	// 发送请求
	requestURL := AliyunSMSEndpoint + "/?" + queryString

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	resp, err := client.Get(requestURL)
	if err != nil {
		return fmt.Errorf("发送阿里云短信请求失败: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("读取响应失败: %v", err)
	}

	var smsResp AliyunSMSResponse
	if err := json.Unmarshal(body, &smsResp); err != nil {
		return fmt.Errorf("解析响应失败: %v", err)
	}

	if smsResp.Code != "OK" {
		return fmt.Errorf("阿里云短信发送失败: %s - %s", smsResp.Code, smsResp.Message)
	}

	return nil
}

// SendQuotaWarningSMS 发送额度告警短信
func SendQuotaWarningSMS(phoneNumber string, remainingQuota string) error {
	templateParam := map[string]string{
		"quote": remainingQuota,
	}
	return SendAliyunSMS(phoneNumber, templateParam, AliyunSMSTemplateCode)
}
