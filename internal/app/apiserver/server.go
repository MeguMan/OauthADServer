package apiserver

import (
	"OauthADServer/internal/app/cache"
	"OauthADServer/internal/app/ldap"
	"OauthADServer/internal/app/models"
	"OauthADServer/internal/app/storage"
	"OauthADServer/internal/app/token"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
	"time"
)

type server struct {
	router *mux.Router
	authMiddleware authenticationMiddleware

	yandexCfg *models.YandexConfig
	googleCfg *models.GoogleConfig
	vkCfg *models.VkConfig
	githubCfg *models.GithubConfig
	mailCfg *models.MailConfig
	odnklsCfg *models.OdnoklassnikiConfig
	discCfg *models.DiscordConfig
	fcbCfg *models.FacebookConfig

	ldapStaffClient ldap.Client
	ldapStudClient ldap.Client
	storage storage.Facade
	tokenManager *token.Manager
	cache *cache.Cache
}

func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}

func NewServer(yandexCfg *models.YandexConfig, googleCfg *models.GoogleConfig,vkCfg *models.VkConfig, githubCfg *models.GithubConfig, mailCfg *models.MailConfig, odnklsCfg *models.OdnoklassnikiConfig, discCfg *models.DiscordConfig, fcbCfg *models.FacebookConfig, ldapStaffClient, ldapStudClient ldap.Client, facade storage.Facade, tokenManager *token.Manager, cache *cache.Cache) *server {
	s := &server{
		router: mux.NewRouter(),
		yandexCfg: yandexCfg,
		googleCfg: googleCfg,
		vkCfg: vkCfg,
		githubCfg: githubCfg,
		mailCfg: mailCfg,
		odnklsCfg: odnklsCfg,
		discCfg: discCfg,
		fcbCfg: fcbCfg,
		ldapStaffClient: ldapStaffClient,
		ldapStudClient: ldapStudClient,
		storage: facade,
		tokenManager: tokenManager,
		authMiddleware: authenticationMiddleware{
			tokenManager: tokenManager,
			storage:      facade,
		},
		cache: cache,
	}
	s.configureRouter()
	return s
}

func (s *server) configureRouter() {
	s.router.Path("/yandex/auth").Handler(http.HandlerFunc(s.HandleYandexAuth())).Methods("GET")
	s.router.Path("/yandex/redirect").Handler(http.HandlerFunc(s.HandleYandexRedirect())).Methods("GET")

	s.router.Path("/vk/auth").Handler(http.HandlerFunc(s.HandleVkAuth())).Methods("GET")
	s.router.Path("/vk/redirect").Handler(http.HandlerFunc(s.HandleVkRedirect())).Methods("GET")

	s.router.Path("/google/auth").Handler(http.HandlerFunc(s.HandleGoogleAuth())).Methods("GET")
	s.router.Path("/google/redirect").Handler(http.HandlerFunc(s.HandleGoogleRedirect())).Methods("GET")

	s.router.Path("/github/auth").Handler(http.HandlerFunc(s.HandleGithubAuth())).Methods("GET")
	s.router.Path("/github/redirect").Handler(http.HandlerFunc(s.HandleGithubRedirect())).Methods("GET")

	s.router.Path("/mail/auth").Handler(http.HandlerFunc(s.HandleMailAuth())).Methods("GET")
	s.router.Path("/mail/redirect").Handler(http.HandlerFunc(s.HandleMailRedirect())).Methods("GET")

	//не возвращает почту
	//s.router.Path("/odnoklassniki/auth").Handler(http.HandlerFunc(s.HandleOdnoklassnikiAuth())).Methods("GET")
	//s.router.Path("/odnoklassniki/redirect").Handler(http.HandlerFunc(s.HandleOdnoklassnikiRedirect())).Methods("GET")

	s.router.Path("/discord/auth").Handler(http.HandlerFunc(s.HandleDiscordAuth())).Methods("GET")
	s.router.Path("/discord/redirect").Handler(http.HandlerFunc(s.HandleDiscordRedirect())).Methods("GET")

	//facebook не работает без впна
	//s.router.Path("/facebook/auth").Handler(http.HandlerFunc(s.HandleFacebookAuth())).Methods("GET")
	//s.router.Path("/facebook/redirect").Handler(http.HandlerFunc(s.HandleFacebookRedirect())).Methods("GET")

	//тестовые ручки
	s.router.Path("/ad/test").Handler(http.HandlerFunc(s.HandleTestAd())).Methods("GET")
	s.router.Path("/db/test").Handler(http.HandlerFunc(s.HandleTestDb())).Methods("GET")

	//блок функций закрытых мидлварью
	api := s.router.PathPrefix("/api/v1").Subrouter()
	api.Use(s.authMiddleware.Middleware)
	api.Path("/testAuth").Handler(http.HandlerFunc(s.test())).Methods("GET")
	api.Path("/ad/{username}").Handler(http.HandlerFunc(s.HandleGetUserInfoFromAd())).Methods("POST")
}

func (s *server) HandleTestAd() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		data, err := s.ldapStaffClient.GetUserInfoByUsername("p.s.novikov")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		resp, err := json.Marshal(data)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, string(resp))
	}
}

func (s *server) HandleTestDb() func(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		id, err := s.storage.GetEmployeeId(ctx, "testingexternal", storage.ExternalServiceTypeYandex)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, id)
	}
}

func (s *server) test() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
	}
}

func (s *server) HandleGetUserInfoFromAd() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		vars := mux.Vars(r)
		username := vars["username"]
		if username == "" {
			http.Error(w, "empty username", http.StatusBadRequest)
			return
		}

		data, err := s.ldapStudClient.GetUserInfoByUsername(username)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		resp, err := json.Marshal(data)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, string(resp))
	}
}

func (s *server) buildJwt(ctx context.Context, externalId, email string, serviceType storage.ExternalServiceType) (*token.JWT, error) {
	//employeeId, err := s.storage.GetEmployeeId(ctx, externalId, serviceType)
	//if errors.Is(err, storage.ErrNotFound) {
		//employeeId, err = s.ldapStaffClient.GetEmployeeNumberByEmail("v.e.podolyak@mospolytech.ru") //заментиь на имейл
		//if err != nil {
		//	return nil, fmt.Errorf("GetEmployeeNumberByEmail: %v", err)
		//}
		//employeeId = "testingqwerty" //для теста потом убрать
		//err := s.storage.CreateLink(ctx, storage.Link{
		//	EmployeeId:            employeeId,
		//	ExternalServiceId:     externalId,
		//	ExternalServiceTypeId: serviceType,
		//})
		//if err != nil {
		//	return nil, err
		//}
	//} else if err != nil {
	//	return nil, fmt.Errorf("GetEmployeeNumberByEmail: %v", err)
	//}

	employeeId := "testingqwerty" //для теста потом убрать

	jwt, err := s.tokenManager.NewJWT(employeeId, externalId, serviceType, time.Minute * 60)
	if err != nil {
		return nil, fmt.Errorf("NewJWT: %v", err)
	}

	return jwt, nil
}

func basicAuth(username, password string) string {
	auth := username + ":" + password
	return base64.StdEncoding.EncodeToString([]byte(auth))
}