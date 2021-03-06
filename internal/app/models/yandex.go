package models

type YandexConfig struct {
	ClientId string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
}

func NewYandexConfig(clientId, clientSecret string) *YandexConfig {
	return &YandexConfig{
		ClientId: clientId,
		ClientSecret: clientSecret,
	}
}

type TokenResponse struct {
	AccessToken string `json:"access_token"`
}

type YandexUserInfo struct {
	Id string `json:"id"`
	Login string `json:"login"`
	ClientId string `json:"client_id"`
	DisplayName string `json:"display_name"`
	RealName string `json:"real_name"`
	FirstName string `json:"first_name"`
	LastName string `json:"last_name"`
	Sex string `json:"sex"`
	DefaultEmail string `json:"default_email"`
	Emails []string `json:"emails"`
	Birthday string `json:"birthday"`
	DefaultAvatarId string `json:"default_avatar_id"`
	IsAvatarEmpty bool `json:"is_avatar_empty"`
	Psuid string `json:"psuid"`
}
