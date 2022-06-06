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
	GithubClientId string `json:"github_client_id"`
	GithubClientSecret string `json:"github_client_secret"`
	LdapStaffHost string `json:"ldap_staff_host"`
	LdapStudHost string `json:"ldap_stud_host"`
	LdapUsername string `json:"ldap_username"`
	LdapPassword string `json:"ldap_password"`
	LdapStaffDn string `json:"ldap_staff_dn"`
	LdapStudDn string `json:"ldap_stud_dn"`
	DatabaseURL string `json:"database_url"`
	JwtSecretKey string `json:"jwt_secret_key"`
}

func NewGlobalConfig() *GlobalConfig {
	return &GlobalConfig{}
}
