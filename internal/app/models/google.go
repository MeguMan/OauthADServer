package models

type GoogleConfig struct {
	ClientId string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
}

func NewGoogleConfig(clientId, clientSecret string) *GoogleConfig {
	return &GoogleConfig{
		ClientId: clientId,
		ClientSecret: clientSecret,
	}
}

type GoogleUserInfo struct {
	Id string `json:"id"`
	Email string `json:"email"`
	VerifiedEmail bool `json:"verified_email"`
	Name string `json:"name"`
	GivenName string `json:"given_name"`
	FamilyName string `json:"family_name"`
	Picture string `json:"picture"`
	Locale string `json:"locale"`
}
