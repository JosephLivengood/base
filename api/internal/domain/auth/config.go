package auth

import (
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type Config struct {
	OAuth2Config  *oauth2.Config
	SecureCookies bool
}

func NewConfig(clientID, clientSecret, redirectURL string, secureCookies bool) *Config {
	return &Config{
		SecureCookies: secureCookies,
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
