package models

type BitrixConfig struct {
	ClientId string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
}

func NewBitrixConfig(clientId, clientSecret string) *BitrixConfig {
	return &BitrixConfig{
		ClientId: clientId,
		ClientSecret: clientSecret,
	}
}

type BitrixTokenResponse struct {
	AccessToken string `json:"access_token"`
	Expires int64 `json:"expires"`
	ExpiresIn int64 `json:"expires_in"`
	Scope string `json:"scope"`
	Domain string `json:"domain"`
	ServerEndpoint string `json:"server_endpoint"`
	Status string `json:"status"`
	ClientEndpoint string `json:"client_endpoint"`
	MemberId string `json:"member_id"`
	UserId int64 `json:"user_id"`
	RefreshToken string `json:"refresh_token"`
}
