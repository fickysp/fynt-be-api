package auth

import (
	"fmt"
	"os"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

var GoogleOAuthConfig *oauth2.Config

func getEnv(key string) string {
	value := os.Getenv(key)
	if value != "" {
		return value
	}
	panic(fmt.Sprintf("Environment variable %s empty!", key))
}

func InitOAuth() {
	GoogleOAuthConfig = &oauth2.Config{
		RedirectURL:  getEnv("GOOGLE_REDIRECT_URL"),
		ClientID:     getEnv("GOOGLE_CLIENT_ID"),
		ClientSecret: getEnv("GOOGLE_CLIENT_SECRET"),
		Scopes:       []string{"https://www.googleapis.com/auth/userinfo.email", "https://www.googleapis.com/auth/userinfo.profile"},
		Endpoint:     google.Endpoint,
	}
}
