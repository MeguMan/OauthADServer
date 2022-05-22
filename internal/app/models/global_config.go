package models

type GlobalConfig struct {
	YandexClientId string `json:"yandex_client_id"`
	YandexClientSecret string `json:"yandex_client_secret"`
	GoogleClientId string `json:"google_client_id"`
	GoogleClientSecret string `json:"google_client_secret"`
	LdapHost string `json:"ldap_host"`
	LdapUsername string `json:"ldap_username"`
	LdapPassword string `json:"ldap_password"`
	LdapDn string `json:"ldap_dn"`
}

func NewGlobalConfig() *GlobalConfig {
	return &GlobalConfig{}
}
