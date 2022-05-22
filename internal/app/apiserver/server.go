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
	ldapClient ldap.Client
}

func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}

func NewServer(yandexCfg *models.YandexConfig, googleCfg *models.GoogleConfig, ldapClient ldap.Client) *server {
	s := &server{
		router: mux.NewRouter(),
		yandexCfg: yandexCfg,
		googleCfg: googleCfg,
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
				fmt.Printf("Can not unmarshal JSON: %v", err)
			}

			urlStr = fmt.Sprintf("https://login.yandex.ru/info?format=json&oauth_token=%s", accessToken.AccessToken)
			req, _ = http.NewRequest("GET", urlStr, nil)
			res, _ = client.Do(req)

			body, _ = ioutil.ReadAll(res.Body)
			var info models.YandexUserInfo
			if err := json.Unmarshal(body, &info); err != nil {
				fmt.Printf("Can not unmarshal JSON: %v", err)
			}
			fmt.Println(info)
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
				fmt.Printf("Can not unmarshal JSON: %v", err)
			}

			urlStr = fmt.Sprintf("https://www.googleapis.com/oauth2/v1/userinfo?alt=json&oauth_token=%s", accessToken.AccessToken)
			req, _ = http.NewRequest("GET", urlStr, nil)
			res, _ = client.Do(req)

			body, _ = ioutil.ReadAll(res.Body)
			var info models.GoogleUserInfo
			if err := json.Unmarshal(body, &info); err != nil {
				fmt.Printf("Can not unmarshal JSON: %v", err)
			}
			fmt.Println(info)
		}
	}
}
