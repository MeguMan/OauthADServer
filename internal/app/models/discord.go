package models

type OdnoklassnikiConfig struct {
	ClientId string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
}

type OdnoklassnikiToken struct {
	AccessToken string `json:"access_token"`
	TokenType string `json:"token_type"`
}

func NewOdnoklassnikiConfig(clientId, clientSecret string) *OdnoklassnikiConfig {
	return &OdnoklassnikiConfig{
		ClientId: clientId,
		ClientSecret: clientSecret,
	}
}
type OdnoklassnikiUserInfo struct {
	Id string `json:"id"`
	Email string `json:"email"`
}
