package apiserver

import (
	"OauthADServer/internal/app/cache"
	"OauthADServer/internal/app/helpers"
	"OauthADServer/internal/app/ldap"
	"OauthADServer/internal/app/models"
	"OauthADServer/internal/app/storage"
	"OauthADServer/internal/app/token"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gorilla/mux"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type server struct {
	router *mux.Router
	authMiddleware authenticationMiddleware

	yandexCfg *models.YandexConfig
	googleCfg *models.GoogleConfig
	vkCfg *models.VkConfig
	bitrixCfg *models.BitrixConfig
	githubCfg *models.GithubConfig

	ldapStaffClient ldap.Client
	ldapStudClient ldap.Client
	storage storage.Facade
	tokenManager *token.Manager
	cache *cache.Cache
}

func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}

func NewServer(yandexCfg *models.YandexConfig, googleCfg *models.GoogleConfig,vkCfg *models.VkConfig, bitrixCfg *models.BitrixConfig, githubCfg *models.GithubConfig, ldapStaffClient, ldapStudClient ldap.Client, facade storage.Facade, tokenManager *token.Manager, cache *cache.Cache) *server {
	s := &server{
		router: mux.NewRouter(),
		yandexCfg: yandexCfg,
		googleCfg: googleCfg,
		vkCfg: vkCfg,
		bitrixCfg: bitrixCfg,
		githubCfg: githubCfg,
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
	//s.router.Path("/").Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	//	http.ServeFile(w, r, "./public/index.html")
	//})).Methods("GET")
	s.router.Path("/yandex/auth").Handler(http.HandlerFunc(s.HandleYandexAuth())).Methods("GET")
	s.router.Path("/yandex/redirect").Handler(http.HandlerFunc(s.HandleYandexRedirect())).Methods("GET")
	s.router.Path("/vk/auth").Handler(http.HandlerFunc(s.HandleVkAuth())).Methods("GET")
	s.router.Path("/vk/redirect").Handler(http.HandlerFunc(s.HandleVkRedirect())).Methods("GET")
	s.router.Path("/google/auth").Handler(http.HandlerFunc(s.HandleGoogleAuth())).Methods("GET")
	s.router.Path("/google/redirect").Handler(http.HandlerFunc(s.HandleGoogleRedirect())).Methods("GET")

	s.router.Path("/github/auth").Handler(http.HandlerFunc(s.HandleGithubAuth())).Methods("GET")
	s.router.Path("/github/redirect").Handler(http.HandlerFunc(s.HandleGithubRedirect())).Methods("GET")
	//s.router.Path("/bitrix24/redirect").Handler(http.HandlerFunc(s.HandleBitrixRedirect())).Methods("GET")

	//блок функций закрытых мидлварью
	api := s.router.PathPrefix("/api/v1").Subrouter()
	api.Use(s.authMiddleware.Middleware)
	api.Path("/testAuth").Handler(http.HandlerFunc(s.test())).Methods("GET")
	api.Path("/ad/{username}").Handler(http.HandlerFunc(s.HandleGetUserInfoFromAd())).Methods("POST")
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

func (s *server) HandleYandexAuth() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		redirectUri := r.URL.Query().Get("redirect_uri")
		if redirectUri == "" {
			http.Error(w, "empty redirect_uri", http.StatusBadRequest)
			return
		}

		oauthCode := helpers.RandStringBytes(5)
		s.cache.Set(oauthCode, &cache.Value{
			RedirectUri: redirectUri,
		}, time.Second * 600)

		http.Redirect(w, r, fmt.Sprintf("https://oauth.yandex.ru/authorize?response_type=code&display=popup&client_id=%s&state=%s", s.yandexCfg.ClientId, oauthCode), http.StatusPermanentRedirect)
	}
}

func (s *server) HandleYandexRedirect() func(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	return func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		state := r.URL.Query().Get("state")
		if code != "" && state != ""{
			data := url.Values{}
			data.Set("grant_type", "authorization_code")
			data.Set("code", code)
			data.Set("client_id", s.yandexCfg.ClientId)
			data.Set("client_secret", s.yandexCfg.ClientSecret)

			urlStr := "https://oauth.yandex.ru/token"

			client := &http.Client{}
			req, _ := http.NewRequest("POST", urlStr, strings.NewReader(data.Encode()))
			req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
			res, _ := client.Do(req)

			body, _ := ioutil.ReadAll(res.Body)
			var accessToken models.TokenResponse
			if err := json.Unmarshal(body, &accessToken); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			urlStr = fmt.Sprintf("https://login.yandex.ru/info?format=json&oauth_token=%s", accessToken.AccessToken)
			req, _ = http.NewRequest("GET", urlStr, nil)
			res, _ = client.Do(req)

			body, _ = ioutil.ReadAll(res.Body)
			var info models.YandexUserInfo
			if err := json.Unmarshal(body, &info); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			val, exists := s.cache.Get(state)
			if !exists {
				http.Error(w, "state not found in cache", http.StatusInternalServerError)
				return
			}

			jwt, err := s.buildJwt(ctx, info.Id, info.DefaultEmail, storage.ExternalServiceTypeYandex)
			if err != nil {
				http.Error(w, "buildJwt", http.StatusInternalServerError)
				return
			}

			http.Redirect(w, r, fmt.Sprintf("%s?access_token=%s", val.RedirectUri, jwt.AccessToken), http.StatusPermanentRedirect)
		}
	}
}

func (s *server) HandleGoogleAuth() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		redirectUri := r.URL.Query().Get("redirect_uri")
		if redirectUri == "" {
			http.Error(w, "empty redirect_uri", http.StatusBadRequest)
			return
		}

		oauthCode := helpers.RandStringBytes(5)
		s.cache.Set(oauthCode, &cache.Value{
			RedirectUri: redirectUri,
		}, time.Second * 600)

		http.Redirect(w, r, fmt.Sprintf("https://accounts.google.com/o/oauth2/auth?client_id=%s&redirect_uri=http://localhost:8080/google/redirect&response_type=code&scope=%s&state=%s",
			s.googleCfg.ClientId, "https://www.googleapis.com/auth/userinfo.email%20https://www.googleapis.com/auth/userinfo.profile", oauthCode), http.StatusPermanentRedirect)
	}
}

func (s *server) HandleGoogleRedirect() func(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	return func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		state := r.URL.Query().Get("state")
		if code != "" && state != ""{
			data := url.Values{}
			data.Set("grant_type", "authorization_code")
			data.Set("code", code)
			data.Set("client_id", s.googleCfg.ClientId)
			data.Set("client_secret", s.googleCfg.ClientSecret)
			data.Set("redirect_uri", "http://localhost:8080/google/redirect")

			urlStr := "https://accounts.google.com/o/oauth2/token"

			client := &http.Client{}
			req, _ := http.NewRequest("POST", urlStr, strings.NewReader(data.Encode()))
			req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
			res, _ := client.Do(req)

			body, _ := ioutil.ReadAll(res.Body)
			var accessToken models.TokenResponse
			if err := json.Unmarshal(body, &accessToken); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			urlStr = fmt.Sprintf("https://www.googleapis.com/oauth2/v1/userinfo?alt=json&oauth_token=%s", accessToken.AccessToken)
			req, _ = http.NewRequest("GET", urlStr, nil)
			res, _ = client.Do(req)

			body, _ = ioutil.ReadAll(res.Body)
			var info models.GoogleUserInfo
			if err := json.Unmarshal(body, &info); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			val, exists := s.cache.Get(state)
			if !exists {
				http.Error(w, "state not found in cache", http.StatusInternalServerError)
				return
			}

			jwt, err := s.buildJwt(ctx, info.Id, info.Email, storage.ExternalServiceTypeGoogle)
			if err != nil {
				http.Error(w, fmt.Sprintf("buildJwt: %v", err), http.StatusInternalServerError)
				return
			}

			http.Redirect(w, r, fmt.Sprintf("%s?access_token=%s", val.RedirectUri, jwt.AccessToken), http.StatusPermanentRedirect)
		}
	}
}

func (s *server) HandleVkAuth() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		redirectUri := r.URL.Query().Get("redirect_uri")
		if redirectUri == "" {
			http.Error(w, "empty redirect_uri", http.StatusBadRequest)
			return
		}

		oauthCode := helpers.RandStringBytes(5)
		s.cache.Set(oauthCode, &cache.Value{
			RedirectUri: redirectUri,
		}, time.Second * 600)

		http.Redirect(w, r, fmt.Sprintf("https://oauth.vk.com/authorize?client_id=%s&redirect_uri=http://localhost:8080/vk/redirect&response_type=code&scope=email&state=%s", s.vkCfg.ClientId, oauthCode), http.StatusPermanentRedirect)
	}
}

func (s *server) HandleVkRedirect() func(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	return func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		state := r.URL.Query().Get("state")
		if code != "" && state != ""{
			urlStr := fmt.Sprintf("https://oauth.vk.com/access_token?client_id=%s&client_secret=%s&redirect_uri=%s&code=%s", s.vkCfg.ClientId, s.vkCfg.ClientSecret, "http://localhost:8080/vk/redirect", code)

			client := &http.Client{}
			req, _ := http.NewRequest("GET", urlStr, nil)
			res, _ := client.Do(req)

			body, _ := ioutil.ReadAll(res.Body)
			var accessTokenResponse models.VkTokenResponse
			if err := json.Unmarshal(body, &accessTokenResponse); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			//urlStr = fmt.Sprintf("https://api.vk.com/method/users.get?v=5.81&uids=%s&access_token=%s&fields=photo_big", accessTokenResponse.UserId, accessTokenResponse.AccessToken)
			//req, _ = http.NewRequest("GET", urlStr, nil)
			//res, _ = client.Do(req)

			//body, _ = ioutil.ReadAll(res.Body)
			//var info models.VkUsersGetResponse
			//if err := json.Unmarshal(body, &info); err != nil {
			//	http.Error(w, err.Error(), http.StatusInternalServerError)
			//	return
			//}

			val, exists := s.cache.Get(state)
			if !exists {
				http.Error(w, "state not found in cache", http.StatusInternalServerError)
				return
			}

			jwt, err := s.buildJwt(ctx, string(accessTokenResponse.UserId), accessTokenResponse.Email, storage.ExternalServiceTypeVkontakte)
			if err != nil {
				http.Error(w, fmt.Sprintf("buildJwt: %v", err), http.StatusInternalServerError)
				return
			}

			http.Redirect(w, r, fmt.Sprintf("%s?access_token=%s", val.RedirectUri, jwt.AccessToken), http.StatusPermanentRedirect)
		}
	}
}

func (s *server) HandleGithubAuth() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		redirectUri := r.URL.Query().Get("redirect_uri")
		if redirectUri == "" {
			http.Error(w, "empty redirect_uri", http.StatusBadRequest)
			return
		}

		oauthCode := helpers.RandStringBytes(5)
		s.cache.Set(oauthCode, &cache.Value{
			RedirectUri: redirectUri,
		}, time.Second * 600)

		http.Redirect(w, r, fmt.Sprintf("https://github.com/login/oauth/authorize?client_id=%s&state=%s&scope=user", s.githubCfg.ClientId, oauthCode), http.StatusPermanentRedirect)
	}
}

func (s *server) HandleGithubRedirect() func(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	return func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		state := r.URL.Query().Get("state")
		if code != "" && state != ""{
			data := url.Values{}
			data.Set("code", code)
			data.Set("client_id", s.githubCfg.ClientId)
			data.Set("client_secret", s.githubCfg.ClientSecret)

			urlStr := "https://github.com/login/oauth/access_token"

			client := &http.Client{}
			req, _ := http.NewRequest("POST", urlStr, strings.NewReader(data.Encode()))
			req.Header.Add("Accept", "application/json")
			res, _ := client.Do(req)

			body, _ := ioutil.ReadAll(res.Body)
			var accessToken models.GithubTokenResponse
			if err := json.Unmarshal(body, &accessToken); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			urlStr = "https://api.github.com/user"
			req, _ = http.NewRequest("GET", urlStr, nil)
			req.Header.Add("Authorization", fmt.Sprintf("token %s", accessToken.AccessToken))
			req.Header.Add("Accept", "application/vnd.github.v3+json")
			res, _ = client.Do(req)
			body, _ = ioutil.ReadAll(res.Body)
			var info models.GithubUserInfo
			if err := json.Unmarshal(body, &info); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			urlStr = "https://api.github.com/user/public_emails"
			req, _ = http.NewRequest("GET", urlStr, nil)
			req.Header.Add("Authorization", fmt.Sprintf("token %s", accessToken.AccessToken))
			req.Header.Add("Accept", "application/vnd.github.v3+json")
			res, _ = client.Do(req)
			body, _ = ioutil.ReadAll(res.Body)
			var emails []models.GithubUserEmail
			if err := json.Unmarshal(body, &emails); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			var primaryEmail string
			for _, email := range emails {
				if email.Primary == true {
					primaryEmail = email.Email
					break
				}
			}

			val, exists := s.cache.Get(state)
			if !exists {
				http.Error(w, "state not found in cache", http.StatusInternalServerError)
				return
			}

			jwt, err := s.buildJwt(ctx, string(info.Id), primaryEmail, storage.ExternalServiceTypeYandex)
			if err != nil {
				http.Error(w, "buildJwt", http.StatusInternalServerError)
				return
			}

			http.Redirect(w, r, fmt.Sprintf("%s?access_token=%s", val.RedirectUri, jwt.AccessToken), http.StatusPermanentRedirect)
		}
	}
}

func (s *server) HandleBitrixRedirect() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		if code != "" {
			urlStr := fmt.Sprintf("https://b24-ep42ak.bitrix24.ru/oauth/token/?client_id=%s&grant_type=authorization_code&client_secret=%s&redirect_uri=%s&code=%s&scope=task,crm", s.bitrixCfg.ClientId, s.bitrixCfg.ClientSecret, "http://localhost:8080/bitrix24/redirect", code)

			client := &http.Client{}
			req, _ := http.NewRequest("GET", urlStr, nil)
			res, _ := client.Do(req)

			body, _ := ioutil.ReadAll(res.Body)
			var accessTokenResponse models.BitrixTokenResponse
			if err := json.Unmarshal(body, &accessTokenResponse); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			fmt.Println(accessTokenResponse)
		}
	}
}

func (s *server) buildJwt(ctx context.Context, externalId, email string, serviceType storage.ExternalServiceType) (*token.JWT, error) {
	employeeId, err := s.storage.GetEmployeeId(ctx, externalId, serviceType)
	if errors.Is(err, storage.ErrNotFound) {
		employeeId, err = s.ldapStaffClient.GetEmployeeNumberByEmail("v.e.podolyak@mospolytech.ru") //заментиь на имейл
		if err != nil {
			return nil, fmt.Errorf("GetEmployeeNumberByEmail: %v", err)
		}
		employeeId = "testingqwerty" //для теста потом убрать
		//err := s.storage.CreateLink(ctx, storage.Link{
		//	EmployeeId:            employeeId,
		//	ExternalServiceId:     externalId,
		//	ExternalServiceTypeId: serviceType,
		//})
		//if err != nil {
		//	return nil, err
		//}
	} else if err != nil {
		return nil, fmt.Errorf("GetEmployeeNumberByEmail: %v", err)
	}

	employeeId = "testingqwerty" //для теста потом убрать

	jwt, err := s.tokenManager.NewJWT(employeeId, externalId, serviceType, time.Minute * 60)
	if err != nil {
		return nil, fmt.Errorf("NewJWT: %v", err)
	}

	return jwt, nil
}