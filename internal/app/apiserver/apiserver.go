package apiserver

import (
	"OauthADServer/internal/app/cache"
	"OauthADServer/internal/app/ldap"
	"OauthADServer/internal/app/models"
	"OauthADServer/internal/app/storage"
	"OauthADServer/internal/app/token"
	"context"
	"fmt"
	"net/http"
	"time"
)

func Start(cfg *models.GlobalConfig) error {
	ctx := context.Background()

	ldapStaffClient, err := ldap.NewClient(ldap.Settings{
		BaseDn:   cfg.LdapStaffDn,
		Host:     cfg.LdapStaffHost,
		Username: cfg.LdapUsername,
		Password: cfg.LdapPassword,
	})
	if err != nil {
		return fmt.Errorf("ldap.NewClient staff: %v", err)
	}

	ldapStudClient, err := ldap.NewClient(ldap.Settings{
		BaseDn:   cfg.LdapStudDn,
		Host:     cfg.LdapStudHost,
		Username: cfg.LdapUsername,
		Password: cfg.LdapPassword,
	})
	if err != nil {
		return fmt.Errorf("ldap.NewClient stud: %v", err)
	}

	pgStorage, err := storage.NewPgStorage(cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("create pg storage: %v", err)
	}
	defer pgStorage.Close(ctx)
	storageFacade := storage.NewStorageFacade(pgStorage)

	tokenManager := token.NewManager(cfg.JwtSecretKey)

	cache := cache.New(time.Second * 60, time.Hour * 5)

	yandexCfg := models.NewYandexConfig(cfg.YandexClientId, cfg.YandexClientSecret)
	googleCfg := models.NewGoogleConfig(cfg.GoogleClientId, cfg.GoogleClientSecret)
	vkCfg := models.NewVkConfig(cfg.VkClientId, cfg.VkClientSecret)
	bitrixCfg := models.NewBitrixConfig(cfg.BitrixClientId, cfg.BitrixClientSecret)
	githubCfg := models.NewGithubConfig(cfg.GithubClientId, cfg.GithubClientSecret)
	server := NewServer(yandexCfg, googleCfg, vkCfg, bitrixCfg, githubCfg, ldapStaffClient, ldapStudClient, storageFacade, tokenManager, cache)

	fmt.Println("server is running")
	return http.ListenAndServe(":8080", server)
}