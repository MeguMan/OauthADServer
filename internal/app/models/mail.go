package models

type MailConfig struct {
	ClientId string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
}

type MailToken struct {
	AccessToken string `json:"access_token"`
	TokenType string `json:"token_type"`
}

func NewMailConfig(clientId, clientSecret string) *MailConfig {
	return &MailConfig{
		ClientId: clientId,
		ClientSecret: clientSecret,
	}
}
type MailUserInfo struct {
	Id string `json:"id"`
	Email string `json:"email"`
}
