package operations

import (
	"github.com/ch629/go-irc-kafka/config"
	"strings"
)

var (
	// TODO: Temp until message handling rewrite
	botConfig   BotConfig
	kafkaConfig config.Kafka
)

type BotConfig struct {
	Name       string
	OAuthToken string
	Channels   []string
}

func InitializeConfig(c config.Config) {
	botConfig = BotConfig{
		Name:       c.Bot.Name,
		OAuthToken: strings.TrimPrefix(c.Bot.OAuth, "oauth:"),
		Channels:   c.Bot.Channels,
	}

	kafkaConfig = c.Kafka
}
