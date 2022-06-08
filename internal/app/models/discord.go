package models

type DiscordConfig struct {
	ClientId string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
}

type DiscordToken struct {
	AccessToken string `json:"access_token"`
	TokenType string `json:"token_type"`
}

func NewDiscordConfig(clientId, clientSecret string) *DiscordConfig {
	return &DiscordConfig{
		ClientId: clientId,
		ClientSecret: clientSecret,
	}
}
type DiscordUserInfo struct {
	Id string `json:"id"`
	Email string `json:"email"`
}
