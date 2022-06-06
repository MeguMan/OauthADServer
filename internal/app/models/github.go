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

//type GithubUserInfo struct {
//	Id string `json:"id"`
//	Login string `json:"login"`
//	ClientId string `json:"client_id"`
//	DisplayName string `json:"display_name"`
//	RealName string `json:"real_name"`
//	FirstName string `json:"first_name"`
//	LastName string `json:"last_name"`
//	Sex string `json:"sex"`
//	DefaultEmail string `json:"default_email"`
//	Emails []string `json:"emails"`
//	Birthday string `json:"birthday"`
//	DefaultAvatarId string `json:"default_avatar_id"`
//	IsAvatarEmpty bool `json:"is_avatar_empty"`
//	Psuid string `json:"psuid"`
//}
