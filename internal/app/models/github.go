package models

type GithubConfig struct {
	ClientId string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
}

type GithubTokenResponse struct {
	AccessToken string `json:"access_token"`
	Scope string `json:"scope"`
	TokenType string `json:"token_type"`
}


type GithubUserEmail struct {
	Email string `json:"email"`
	Primary bool `json:"primary"`
	Verified bool `json:"verified"`
}

type GithubUserInfo struct {
	Id int `json:"id"`
}

func NewGithubConfig(clientId, clientSecret string) *GithubConfig {
	return &GithubConfig{
		ClientId: clientId,
		ClientSecret: clientSecret,
	}
}