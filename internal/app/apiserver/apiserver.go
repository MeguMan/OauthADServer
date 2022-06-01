package apiserver

import (
	"OauthADServer/internal/app/ldap"
	"OauthADServer/internal/app/models"
	"fmt"
	"net/http"
)

func Start(cfg *models.GlobalConfig) error{
	ldapClient, err := ldap.NewClient(ldap.Settings{
		BaseDn:   cfg.LdapDn,
		Host:     cfg.LdapHost,
		Username: cfg.LdapUsername,
		Password: cfg.LdapPassword,
	})
	if err != nil {
		return fmt.Errorf("ldap.NewClient: %v", err)
	}

	a, err := ldapClient.GetEmployeeNumberByEmail("p.novikov@mami.ru")
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(a)

	yandexCfg := models.NewYandexConfig(cfg.YandexClientId, cfg.YandexClientSecret)
	googleCfg := models.NewGoogleConfig(cfg.GoogleClientId, cfg.GoogleClientSecret)
	vkCfg := models.NewVkConfig(cfg.VkClientId, cfg.VkClientSecret)
	bitrixCfg := models.NewBitrixConfig(cfg.BitrixClientId, cfg.BitrixClientSecret)
	server := NewServer(yandexCfg, googleCfg, vkCfg, bitrixCfg, ldapClient)

	fmt.Println("server is running")
	return http.ListenAndServe(":8080", server)
}