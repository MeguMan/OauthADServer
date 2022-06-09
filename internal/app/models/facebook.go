package models

type FacebookConfig struct {
	ClientId string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
}

type FacebookToken struct {
	AccessToken string `json:"access_token"`
}

func NewFacebookConfig(clientId, clientSecret string) *FacebookConfig {
	return &FacebookConfig{
		ClientId: clientId,
		ClientSecret: clientSecret,
	}
}
type FacebookUserInfo struct {
	Id string `json:"id"`
	Email string `json:"email"`
}
