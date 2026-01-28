package main

import (
	"fmt"
	"log"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
)

func main() {
	// ===== 重要：请修改这里的API Key =====
	apiKey := "sk-6d1cefce7ccc41c8b31d06d3659be36efzrWxMh9WVLk8ZYb" // 请替换为你的实际Cqtai API Key
	// ====================================

	fmt.Println("🚀 开始添加 Cqtai 渠道...")

	// 初始化数据库
	err := model.InitDB()
	if err != nil {
		log.Fatal("❌ 初始化数据库失败:", err)
	}
	defer model.CloseDB()

	// 准备渠道数据
	baseURL := "https://api.cqtai.com"
	models := "cqtai_music,cqtai_lyrics"
	channelName := "Cqtai"

	channel := &model.Channel{
		Type:        58, // ChannelTypeCqtai
		Name:        channelName,
		Key:         apiKey,
		BaseURL:     &baseURL,
		Models:      models,
		Status:      1, // 1=启用
		Weight:      common.GetPointer[uint](0),
		Priority:    common.GetPointer[int64](0),
		CreatedTime: common.GetTimestamp(),
		Group:       "default",
		AutoBan:     common.GetPointer[int](1),
	}

	// 插入渠道
	err = channel.Insert()
	if err != nil {
		log.Fatal("❌ 插入渠道失败:", err)
	}

	fmt.Printf("✅ 成功添加 Cqtai 渠道，ID: %d\n", channel.Id)
	fmt.Printf("   渠道名称: %s\n", channel.Name)
	fmt.Printf("   Base URL: %s\n", *channel.BaseURL)
	fmt.Printf("   支持模型: %s\n", channel.Models)
	fmt.Printf("   状态: %d (1=启用)\n", channel.Status)

	// 验证abilities
	var abilities []model.Ability
	err = model.DB.Where("channel_id = ?", channel.Id).Find(&abilities).Error
	if err != nil {
		log.Printf("⚠️  查询abilities失败: %v", err)
	} else {
		fmt.Println("\n📋 已添加的模型能力:")
		for _, ability := range abilities {
			status := "禁用"
			if ability.Enabled {
				status = "启用"
			}
			fmt.Printf("   - %s (%s)\n", ability.Model, status)
		}
	}

	fmt.Println("\n✅ Cqtai 渠道添加完成!")
	fmt.Println("\n📝 后续步骤:")
	fmt.Println("1. 重启 new-api 服务以加载新渠道")
	fmt.Println("2. 在管理后台 -> 渠道管理 中查看和测试渠道")
	fmt.Println("3. 确保 Token 中启用了 cqtai_music 和 cqtai_lyrics 模型")
	fmt.Println("4. 使用 ./debug_cqtai.sh 脚本测试渠道功能")
}
