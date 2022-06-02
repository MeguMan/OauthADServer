package apiserver

import (
	"OauthADServer/internal/app/ldap"
	"OauthADServer/internal/app/models"
	"OauthADServer/internal/app/storage"
	"context"
	"fmt"
	"net/http"
)

func Start(cfg *models.GlobalConfig) error {
	ctx := context.Background()

	ldapClient, err := ldap.NewClient(ldap.Settings{
		BaseDn:   cfg.LdapDn,
		Host:     cfg.LdapHost,
		Username: cfg.LdapUsername,
		Password: cfg.LdapPassword,
	})
	if err != nil {
		return fmt.Errorf("ldap.NewClient: %v", err)
	}

	pgStorage, err := storage.NewPgStorage(cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("create pg storage: %v", err)
	}
	defer pgStorage.Close(ctx)
	storageFacade := storage.NewStorageFacade(pgStorage)

	yandexCfg := models.NewYandexConfig(cfg.YandexClientId, cfg.YandexClientSecret)
	googleCfg := models.NewGoogleConfig(cfg.GoogleClientId, cfg.GoogleClientSecret)
	vkCfg := models.NewVkConfig(cfg.VkClientId, cfg.VkClientSecret)
	bitrixCfg := models.NewBitrixConfig(cfg.BitrixClientId, cfg.BitrixClientSecret)
	server := NewServer(yandexCfg, googleCfg, vkCfg, bitrixCfg, ldapClient, storageFacade)

	fmt.Println("server is running")
	return http.ListenAndServe(":8080", server)
}