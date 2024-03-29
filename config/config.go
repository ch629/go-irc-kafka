package config

import (
	"io"
	"sync"

	"github.com/spf13/afero"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

type (
	Config struct {
		Bot   Bot
		Kafka Kafka
		Irc   Irc
	}
	Bot struct {
		Name     string
		OAuth    string
		Channels []string
	}
	Kafka struct {
		Brokers []string
		Topic   string
	}
	Irc struct {
		Address string
	}
)

var (
	config = Config{
		Bot: Bot{
			Name:     "",
			OAuth:    "",
			Channels: []string{},
		},
		Kafka: Kafka{
			Brokers: []string{"localhost:9092"},
			Topic:   "",
		},
		Irc: Irc{
			Address: "irc.chat.twitch.tv:6667",
		},
	}

	configInit sync.Once
)

func LoadConfig(fs afero.Fs) (Config, error) {
	var err error
	configInit.Do(func() {
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
		viper.AddConfigPath(".")

		if err = viper.ReadInConfig(); err != nil {
			if _, ok := err.(viper.ConfigFileNotFoundError); ok {
				// Config not found
				var f io.WriteCloser
				if f, err = fs.Create("config.yaml"); err != nil {
					return
				}
				defer f.Close()
				if err = yaml.NewEncoder(f).Encode(config); err != nil {
					return
				}
			} else {
				return
			}
		}
		err = viper.Unmarshal(&config)
	})

	return config, err
}
