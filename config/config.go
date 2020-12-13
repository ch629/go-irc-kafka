package config

import (
	"github.com/spf13/viper"
)

func LoadConfig() {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config not found
			setupDefaults()

			if err = viper.WriteConfigAs("config.yaml"); err != nil {
				panic(err)
			}
		} else {
			panic(err)
		}
	}
}

func setupDefaults() {
	viper.SetDefault("bot.name", "")
	viper.SetDefault("bot.oauth", "")
	viper.SetDefault("bot.channels", []string{})

	viper.SetDefault("kafka.brokers", []string{"localhost:9092"})
	viper.SetDefault("kafka.topic", "example-topic")

}
