package client

import (
	"github.com/spf13/viper"
	"strings"
)

var BaseBotConfig BotConfig

type BotConfig struct {
	Name       string
	OAuthToken string
	Address    string
	Channels   []string
}

func InitializeConfig() {
	BaseBotConfig = BotConfig{
		Name:       viper.GetString("bot.name"),
		OAuthToken: strings.TrimPrefix(viper.GetString("bot.oauth"), "oauth:"),
		Address:    "irc.chat.twitch.tv:6667",
		Channels:   viper.GetStringSlice("bot.channels"),
	}
}
