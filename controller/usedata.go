package controller

import (
	"net/http"
	"strconv"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"

	"github.com/gin-gonic/gin"
)

func GetAllQuotaDates(c *gin.Context) {
	startTimestamp, _ := strconv.ParseInt(c.Query("start_timestamp"), 10, 64)
	endTimestamp, _ := strconv.ParseInt(c.Query("end_timestamp"), 10, 64)
	username := c.Query("username")
	dates, err := model.GetAllQuotaDates(startTimestamp, endTimestamp, username)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    dates,
	})
	return
}

func GetUserQuotaDates(c *gin.Context) {
	userId := c.GetInt("id")
	startTimestamp, _ := strconv.ParseInt(c.Query("start_timestamp"), 10, 64)
	endTimestamp, _ := strconv.ParseInt(c.Query("end_timestamp"), 10, 64)
	// 判断时间跨度是否超过 1 个月
	if endTimestamp-startTimestamp > 2592000 {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "时间跨度不能超过 1 个月",
		})
		return
	}
	dates, err := model.GetQuotaDataByUserId(userId, startTimestamp, endTimestamp)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    dates,
	})
	return
}

// GetPublicQuotaDates 公共数据看板接口 - 未登录用户可访问
func GetPublicQuotaDates(c *gin.Context) {
	startTimestamp, _ := strconv.ParseInt(c.Query("start_timestamp"), 10, 64)
	endTimestamp, _ := strconv.ParseInt(c.Query("end_timestamp"), 10, 64)

	// 如果没有提供时间参数，使用默认的最近7天
	if startTimestamp == 0 || endTimestamp == 0 {
		now := time.Now()
		endTimestamp = now.Unix()
		startTimestamp = now.AddDate(0, 0, -7).Unix()
	}

	// 限制查询时间范围不超过31天，防止性能问题
	if endTimestamp-startTimestamp > 31*24*3600 {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "时间跨度不能超过31天",
		})
		return
	}

	// 获取按时间和模型分组的详细数据（用于图表）
	// username为空时，GetAllQuotaDates会查询所有数据
	quotaData, err := model.GetAllQuotaDates(startTimestamp, endTimestamp, "")
	if err != nil {
		common.ApiError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    quotaData,
	})
	return
}

// GetPublicStats 获取公开统计数据接口 - 未登录用户可访问
func GetPublicStats(c *gin.Context) {
	startTimestamp, _ := strconv.ParseInt(c.Query("start_timestamp"), 10, 64)
	endTimestamp, _ := strconv.ParseInt(c.Query("end_timestamp"), 10, 64)

	// 默认最近24小时
	if startTimestamp == 0 || endTimestamp == 0 {
		now := time.Now()
		endTimestamp = now.Unix()
		startTimestamp = now.AddDate(0, 0, -1).Unix()
	}

	// 限制查询范围(最多31天)
	if endTimestamp-startTimestamp > 31*24*3600 {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "时间跨度不能超过31天",
		})
		return
	}

	// 获取聚合统计
	stats, err := model.GetPublicStats()
	if err != nil {
		common.ApiError(c, err)
		return
	}

	// 获取Top用户
	topUsers, err := model.GetTopUsersByConsumption(startTimestamp, endTimestamp, 10)
	if err != nil {
		common.ApiError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data": gin.H{
			"stats":     stats,
			"top_users": topUsers,
		},
	})
}
