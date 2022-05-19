package apiserver

import (
	"OauthADServer/internal/app/ldap"
	"OauthADServer/internal/app/models"
	"fmt"
	"net/http"
)

func Start(cfg *models.GlobalConfig) error{
	ldapClient, err := ldap.NewClient(ldap.Settings{
		Host:     cfg.LdapHost,
		Username: cfg.LdapUsername,
		Password: cfg.LdapPassword,
		BaseDn:   cfg.LdapDn,
	})
	if err != nil {
		return fmt.Errorf("ldap.NewClient: %v", err)
	}

	yandexCfg := models.NewYandexConfig(cfg.YandexClientId, cfg.YandexClientSecret)

	server := NewServer(yandexCfg, ldapClient)

	fmt.Println("server is running")
	return http.ListenAndServe(":8080", server)
}