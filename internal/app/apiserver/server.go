package apiserver

import (
	"OauthADServer/internal/app/ldap"
	"OauthADServer/internal/app/models"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

type server struct {
	router *mux.Router
	yandexCfg *models.YandexConfig
	googleCfg *models.GoogleConfig
	vkCfg *models.VkConfig
	bitrixCfg *models.BitrixConfig
	ldapClient ldap.Client
}

func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}

func NewServer(yandexCfg *models.YandexConfig, googleCfg *models.GoogleConfig,vkCfg *models.VkConfig, bitrixCfg *models.BitrixConfig, ldapClient ldap.Client) *server {
	s := &server{
		router: mux.NewRouter(),
		yandexCfg: yandexCfg,
		googleCfg: googleCfg,
		vkCfg: vkCfg,
		bitrixCfg: bitrixCfg,
		ldapClient: ldapClient,
	}
	s.configureRouter()
	return s
}

func (s *server) configureRouter() {
	s.router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./public/index.html")
	}).Methods("GET")
	s.router.HandleFunc("/yandex/redirect", s.HandleYandexRedirect()).Methods("GET")
	s.router.HandleFunc("/google/redirect", s.HandleGoogleRedirect()).Methods("GET")
	s.router.HandleFunc("/vk/redirect", s.HandleVkRedirect()).Methods("GET")
	s.router.HandleFunc("/bitrix24/redirect", s.HandleBitrixRedirect()).Methods("GET")
	s.router.HandleFunc("/ad/{username}", s.HandleGetUserInfoFromAd()).Methods("POST")
}

func (s *server) HandleGetUserInfoFromAd() func(w http.ResponseWriter, r *http.Request) {
	type ReqBody struct {
		Login string `json:"login"`
		Password string `json:"password"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		vars := mux.Vars(r)
		username := vars["username"]
		if username == "" {
			http.Error(w, "empty username", http.StatusBadRequest)
			return
		}

		req := ReqBody{}
		b, _ := ioutil.ReadAll(r.Body)
		err := json.Unmarshal(b, &req)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		data, err := s.ldapClient.GetUserInfoByUsername(username)
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

func (s *server) HandleYandexRedirect() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		if code != "" {
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

			id, err := s.ldapClient.GetEmployeeNumberByEmail("p.novikov@mami.ru")
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			fmt.Println(id)
		}
	}
}

func (s *server) HandleGoogleRedirect() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		if code != "" {
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
			fmt.Println(info)
		}
	}
}

func (s *server) HandleVkRedirect() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		if code != "" {
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

			urlStr = fmt.Sprintf("https://api.vk.com/method/users.get?v=5.81&uids=%s&access_token=%s&fields=photo_big", accessTokenResponse.UserId, accessTokenResponse.AccessToken)
			req, _ = http.NewRequest("GET", urlStr, nil)
			res, _ = client.Do(req)

			body, _ = ioutil.ReadAll(res.Body)
			var info models.VkUsersGetResponse
			if err := json.Unmarshal(body, &info); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			fmt.Println(info.Response[0])
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
