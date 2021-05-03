package client

import (
	"github.com/ch629/go-irc-kafka/config"
	"strings"
)

var BaseBotConfig BotConfig

type BotConfig struct {
	Name       string
	OAuthToken string
	Address    string
	Channels   []string
}

func InitializeConfig(config config.Bot) {
	BaseBotConfig = BotConfig{
		Name:       config.Name,
		OAuthToken: strings.TrimPrefix(config.OAuth, "oauth:"),
		Address:    "irc.chat.twitch.tv:6667",
		Channels:   config.Channels,
	}
}
