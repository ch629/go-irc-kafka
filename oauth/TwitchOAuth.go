package oauth

import (
	"context"
	"github.com/spf13/viper"
	"golang.org/x/oauth2/clientcredentials"
	"golang.org/x/oauth2/twitch"
)

var oauth2Config *clientcredentials.Config

func initializeConfig() {
	if oauth2Config == nil {
		oauth2Config = &clientcredentials.Config{
			ClientID:     viper.GetString("oauth.id"),
			ClientSecret: viper.GetString("oauth.secret"),
			TokenURL:     twitch.Endpoint.TokenURL,
		}
	}
}

// TODO: This gets the token for an application, validate how to get it for a user
func GetToken() (string, error) {
	initializeConfig()
	token, err := oauth2Config.Token(context.Background())
	if err != nil {
		return "", err
	}

	return token.AccessToken, nil
}
