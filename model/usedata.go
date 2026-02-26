package model

import (
	"fmt"
	"sync"
	"time"

	"github.com/QuantumNous/new-api/common"
	"gorm.io/gorm"
)

// QuotaData 柱状图数据
type QuotaData struct {
	Id        int    `json:"id"`
	UserID    int    `json:"user_id" gorm:"index"`
	Username  string `json:"username" gorm:"index:idx_qdt_model_user_name,priority:2;size:64;default:''"`
	ModelName string `json:"model_name" gorm:"index:idx_qdt_model_user_name,priority:1;size:64;default:''"`
	CreatedAt int64  `json:"created_at" gorm:"bigint;index:idx_qdt_created_at,priority:2"`
	TokenUsed int    `json:"token_used" gorm:"default:0"`
	Count     int    `json:"count" gorm:"default:0"`
	Quota     int    `json:"quota" gorm:"default:0"`
}

func UpdateQuotaData() {
	for {
		if common.DataExportEnabled {
			common.SysLog("正在更新数据看板数据...")
			SaveQuotaDataCache()
		}
		time.Sleep(time.Duration(common.DataExportInterval) * time.Minute)
	}
}

var CacheQuotaData = make(map[string]*QuotaData)
var CacheQuotaDataLock = sync.Mutex{}

func logQuotaDataCache(userId int, username string, modelName string, quota int, createdAt int64, tokenUsed int) {
	key := fmt.Sprintf("%d-%s-%s-%d", userId, username, modelName, createdAt)
	quotaData, ok := CacheQuotaData[key]
	if ok {
		quotaData.Count += 1
		quotaData.Quota += quota
		quotaData.TokenUsed += tokenUsed
	} else {
		quotaData = &QuotaData{
			UserID:    userId,
			Username:  username,
			ModelName: modelName,
			CreatedAt: createdAt,
			Count:     1,
			Quota:     quota,
			TokenUsed: tokenUsed,
		}
	}
	CacheQuotaData[key] = quotaData
}

func LogQuotaData(userId int, username string, modelName string, quota int, createdAt int64, tokenUsed int) {
	// 只精确到小时
	createdAt = createdAt - (createdAt % 3600)

	CacheQuotaDataLock.Lock()
	defer CacheQuotaDataLock.Unlock()
	logQuotaDataCache(userId, username, modelName, quota, createdAt, tokenUsed)
}

func SaveQuotaDataCache() {
	CacheQuotaDataLock.Lock()
	defer CacheQuotaDataLock.Unlock()
	size := len(CacheQuotaData)
	// 如果缓存中有数据，就保存到数据库中
	// 1. 先查询数据库中是否有数据
	// 2. 如果有数据，就更新数据
	// 3. 如果没有数据，就插入数据
	for _, quotaData := range CacheQuotaData {
		quotaDataDB := &QuotaData{}
		DB.Table("quota_data").Where("user_id = ? and username = ? and model_name = ? and created_at = ?",
			quotaData.UserID, quotaData.Username, quotaData.ModelName, quotaData.CreatedAt).First(quotaDataDB)
		if quotaDataDB.Id > 0 {
			//quotaDataDB.Count += quotaData.Count
			//quotaDataDB.Quota += quotaData.Quota
			//DB.Table("quota_data").Save(quotaDataDB)
			increaseQuotaData(quotaData.UserID, quotaData.Username, quotaData.ModelName, quotaData.Count, quotaData.Quota, quotaData.CreatedAt, quotaData.TokenUsed)
		} else {
			DB.Table("quota_data").Create(quotaData)
		}
	}
	CacheQuotaData = make(map[string]*QuotaData)
	common.SysLog(fmt.Sprintf("保存数据看板数据成功，共保存%d条数据", size))
}

func increaseQuotaData(userId int, username string, modelName string, count int, quota int, createdAt int64, tokenUsed int) {
	err := DB.Table("quota_data").Where("user_id = ? and username = ? and model_name = ? and created_at = ?",
		userId, username, modelName, createdAt).Updates(map[string]interface{}{
		"count":      gorm.Expr("count + ?", count),
		"quota":      gorm.Expr("quota + ?", quota),
		"token_used": gorm.Expr("token_used + ?", tokenUsed),
	}).Error
	if err != nil {
		common.SysLog(fmt.Sprintf("increaseQuotaData error: %s", err))
	}
}

func GetQuotaDataByUsername(username string, startTime int64, endTime int64) (quotaData []*QuotaData, err error) {
	var quotaDatas []*QuotaData
	// 从quota_data表中查询数据
	err = DB.Table("quota_data").Where("username = ? and created_at >= ? and created_at <= ?", username, startTime, endTime).Find(&quotaDatas).Error
	return quotaDatas, err
}

func GetQuotaDataByUserId(userId int, startTime int64, endTime int64) (quotaData []*QuotaData, err error) {
	var quotaDatas []*QuotaData
	// 从quota_data表中查询数据
	err = DB.Table("quota_data").Where("user_id = ? and created_at >= ? and created_at <= ?", userId, startTime, endTime).Find(&quotaDatas).Error
	return quotaDatas, err
}

func GetAllQuotaDates(startTime int64, endTime int64, username string) (quotaData []*QuotaData, err error) {
	if username != "" {
		return GetQuotaDataByUsername(username, startTime, endTime)
	}
	var quotaDatas []*QuotaData
	// 从quota_data表中查询数据
	// only select model_name, sum(count) as count, sum(quota) as quota, model_name, created_at from quota_data group by model_name, created_at;
	//err = DB.Table("quota_data").Where("created_at >= ? and created_at <= ?", startTime, endTime).Find(&quotaDatas).Error
	err = DB.Table("quota_data").Select("model_name, sum(count) as count, sum(quota) as quota, sum(token_used) as token_used, created_at").Where("created_at >= ? and created_at <= ?", startTime, endTime).Group("model_name, created_at").Find(&quotaDatas).Error
	return quotaDatas, err
}

// PublicStats 公开统计数据结构
type PublicStats struct {
	EnabledModelsCount   int64 `json:"enabled_models_count"`   // 启用的模型数量
	EnabledChannelsCount int64 `json:"enabled_channels_count"` // 启用的渠道(服务商)数量
	ActiveTokensCount    int64 `json:"active_tokens_count"`    // 有效令牌数量
	TodayTokenUsage      int64 `json:"today_token_usage"`      // 今日Token消耗
	TotalReqCount        int64 `json:"total_req_count"`        // 总请求数
	TotalQuota           int64 `json:"total_quota"`            // 总消耗额度
	TotalTokenUsage      int64 `json:"total_token_usage"`      // 总token消耗
	TotalDataCount       int64 `json:"total_data_count"`       // 总数据记录数
}

// GetPublicStats 获取公开统计数据
func GetPublicStats() (*PublicStats, error) {
	stats := &PublicStats{}

	// 查询1: 统计启用的模型数量 (从abilities表, enabled=true, 去重)
	err := DB.Table("abilities").Where("enabled = ?", true).Distinct("model").Count(&stats.EnabledModelsCount).Error
	if err != nil {
		return nil, err
	}

	// 查询2: 统计启用的渠道(服务商)数量 (status=1, deleted_at IS NULL)
	err = DB.Table("channels").Where("status = ?", 1).Count(&stats.EnabledChannelsCount).Error
	if err != nil {
		return nil, err
	}

	// 查询3: 统计有效令牌数量 (deleted_at IS NULL - GORM会自动处理软删除)
	err = DB.Model(&Token{}).Count(&stats.ActiveTokensCount).Error
	if err != nil {
		return nil, err
	}

	// 查询4: 统计今日Token消耗 (从quota_data表,按自然日0:00-23:59)
	now := time.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).Unix()
	todayEnd := todayStart + 24*3600

	var result struct {
		TotalTokens int64
	}
	err = DB.Table("quota_data").
		Select("COALESCE(SUM(token_used), 0) as total_tokens").
		Where("created_at >= ? AND created_at < ?", todayStart, todayEnd).
		Scan(&result).Error
	if err != nil {
		return nil, err
	}
	stats.TodayTokenUsage = result.TotalTokens

	// 查询5: 统计总数据 (从quota_data表)
	var totalResult struct {
		TotalReqCount   int64
		TotalQuota      int64
		TotalTokenUsage int64
		TotalDataCount  int64
	}
	err = DB.Table("quota_data").
		Select("COALESCE(SUM(count), 0) as total_req_count, COALESCE(SUM(quota), 0) as total_quota, COALESCE(SUM(token_used), 0) as total_token_usage, COUNT(*) as total_data_count").
		Scan(&totalResult).Error
	if err != nil {
		return nil, err
	}
	stats.TotalReqCount = totalResult.TotalReqCount
	stats.TotalQuota = totalResult.TotalQuota
	stats.TotalTokenUsage = totalResult.TotalTokenUsage
	stats.TotalDataCount = totalResult.TotalDataCount

	return stats, nil
}

// anonymizeUsername 用户名匿名化处理
// 示例: "john_doe" -> "joh***oe", "alice" -> "a***e", "ab" -> "a***"
func anonymizeUsername(username string) string {
	if username == "" {
		return "匿名用户"
	}

	length := len(username)
	if length <= 2 {
		return string(username[0]) + "***"
	} else if length <= 4 {
		return string(username[0]) + "***" + string(username[length-1])
	} else {
		return username[:3] + "***" + username[length-2:]
	}
}

// GetTopUsersByConsumption 获取消费Top用户列表(用户名匿名化)
func GetTopUsersByConsumption(startTime int64, endTime int64, limit int) ([]map[string]interface{}, error) {
	// 参数验证
	if limit <= 0 {
		limit = 10
	}
	if limit > 50 {
		limit = 50 // 防止滥用
	}

	var results []struct {
		Username     string
		TotalQuota   int64
		TotalTokens  int64
		RequestCount int64
	}

	// 按用户名聚合查询
	err := DB.Table("quota_data").
		Select("username, SUM(quota) as total_quota, SUM(token_used) as total_tokens, SUM(count) as request_count").
		Where("created_at >= ? AND created_at <= ?", startTime, endTime).
		Group("username").
		Order("total_quota DESC").
		Limit(limit).
		Scan(&results).Error

	if err != nil {
		return nil, err
	}

	// 用户名匿名化处理
	anonymizedResults := make([]map[string]interface{}, len(results))
	for i, r := range results {
		anonymized := anonymizeUsername(r.Username)

		anonymizedResults[i] = map[string]interface{}{
			"username":      anonymized,
			"quota":         r.TotalQuota,
			"token_used":    r.TotalTokens,
			"request_count": r.RequestCount,
		}
	}

	return anonymizedResults, nil
}
