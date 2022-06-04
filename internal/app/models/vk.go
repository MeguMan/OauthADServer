package models

type VkConfig struct {
	ClientId string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
}

func NewVkConfig(clientId, clientSecret string) *VkConfig {
	return &VkConfig{
		ClientId: clientId,
		ClientSecret: clientSecret,
	}
}

type VkTokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn int64 `json:"expires_in"`
	UserId int64 `json:"user_id"`
	Email string `json:"email"`
}

type VkUsersGetResponse struct {
	Response []*VkUserInfo `json:"response"`
}

type VkUserInfo struct {
	Id int64 `json:"id"`
	PhotoBig string `json:"photo_big"`
	FirstName string `json:"first_name"`
	LastName string `json:"last_name"`
}
