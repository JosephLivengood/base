package auth

import (
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type Config struct {
	OAuth2Config *oauth2.Config
}

func NewConfig(clientID, clientSecret, redirectURL string) *Config {
	return &Config{
		OAuth2Config: &oauth2.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			RedirectURL:  redirectURL,
			Scopes: []string{
				"https://www.googleapis.com/auth/userinfo.email",
				"https://www.googleapis.com/auth/userinfo.profile",
			},
			Endpoint: google.Endpoint,
		},
	}
}
