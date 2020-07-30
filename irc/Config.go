package irc

import "github.com/spf13/viper"

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
		OAuthToken: viper.GetString("bot.oauth"),
		Address:    "irc.chat.twitch.tv:6667",
		Channels:   viper.GetStringSlice("bot.channels"),
	}
}
