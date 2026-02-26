package model_setting

import (
	"strings"

	"github.com/QuantumNous/new-api/setting/config"
)

// ChannelPolicy defines a policy for applying settings to specific channels
type ChannelPolicy struct {
	Enabled       bool     `json:"enabled"`
	ChannelIDs    []int    `json:"channel_ids"`
	ChannelTypes  []int    `json:"channel_types"`
	ModelPatterns []string `json:"model_patterns"`
}

// IsChannelEnabled checks if the policy is enabled for the given channel
func (p *ChannelPolicy) IsChannelEnabled(channelID int, channelType int) bool {
	if !p.Enabled {
		return false
	}
	// Check channel IDs
	if len(p.ChannelIDs) > 0 {
		for _, id := range p.ChannelIDs {
			if id == channelID {
				return true
			}
		}
		return false
	}
	// Check channel types
	if len(p.ChannelTypes) > 0 {
		for _, ct := range p.ChannelTypes {
			if ct == channelType {
				return true
			}
		}
		return false
	}
	return true
}

// ChatCompletionsToResponsesPolicy defines when to use Responses API for chat completions
type ChatCompletionsToResponsesPolicy struct {
	ChannelPolicy
}

type GlobalSettings struct {
	PassThroughRequestEnabled        bool                                   `json:"pass_through_request_enabled"`
	ThinkingModelBlacklist           []string                               `json:"thinking_model_blacklist"`
	ChatCompletionsToResponsesPolicy ChatCompletionsToResponsesPolicy      `json:"chat_completions_to_responses_policy"`
}

// 默认配置
var defaultOpenaiSettings = GlobalSettings{
	PassThroughRequestEnabled: false,
	ThinkingModelBlacklist: []string{
		"moonshotai/kimi-k2-thinking",
		"kimi-k2-thinking",
	},
	ChatCompletionsToResponsesPolicy: ChatCompletionsToResponsesPolicy{
		ChannelPolicy: ChannelPolicy{
			Enabled:       false,
			ChannelIDs:    []int{},
			ChannelTypes:  []int{},
			ModelPatterns: []string{},
		},
	},
}

// 全局实例
var globalSettings = defaultOpenaiSettings

func init() {
	// 注册到全局配置管理器
	config.GlobalConfig.Register("global", &globalSettings)
}

func GetGlobalSettings() *GlobalSettings {
	return &globalSettings
}

// ShouldPreserveThinkingSuffix 判断模型是否配置为保留 thinking/-nothinking/-low/-high/-medium 后缀
func ShouldPreserveThinkingSuffix(modelName string) bool {
	target := strings.TrimSpace(modelName)
	if target == "" {
		return false
	}

	for _, entry := range globalSettings.ThinkingModelBlacklist {
		if strings.TrimSpace(entry) == target {
			return true
		}
	}
	return false
}
