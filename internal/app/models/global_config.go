package models

type GlobalConfig struct {
	YandexClientId string `json:"yandex_client_id"`
	YandexClientSecret string `json:"yandex_client_secret"`
	GoogleClientId string `json:"google_client_id"`
	GoogleClientSecret string `json:"google_client_secret"`
	VkClientId string `json:"vk_client_id"`
	VkClientSecret string `json:"vk_client_secret"`
	BitrixClientId string `json:"bitrix_client_id"`
	BitrixClientSecret string `json:"bitrix_client_secret"`
	LdapHost string `json:"ldap_host"`
	LdapUsername string `json:"ldap_username"`
	LdapPassword string `json:"ldap_password"`
	LdapDn string `json:"ldap_dn"`
	DatabaseURL string `json:"database_url"`
	JwtSecretKey string `json:"jwt_secret_key"`
}

func NewGlobalConfig() *GlobalConfig {
	return &GlobalConfig{}
}
