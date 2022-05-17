package apiserver

import (
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
	config *Config
}

func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}

func NewServer(config *Config) *server {
	s := &server{
		router: mux.NewRouter(),
		config: config,
	}
	s.configureRouter()
	return s
}

func (s *server) configureRouter() {
	s.router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./public/index.html")
	}).Methods("GET")
	s.router.HandleFunc("/yandex-auth", s.HandleYandexRedirect()).Methods("GET")
}

func (s *server) HandleYandexRedirect() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		if code != "" {
			data := url.Values{}
			data.Set("grant_type", "authorization_code")
			data.Set("code", code)
			data.Set("client_id", s.config.YandexClientId)
			data.Set("client_secret", s.config.YandexClientSecret)

			urlStr := "https://oauth.yandex.ru/token"

			client := &http.Client{}
			req, _ := http.NewRequest("POST", urlStr, strings.NewReader(data.Encode()))
			req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
			res, _ := client.Do(req)

			body, _ := ioutil.ReadAll(res.Body)
			var accessToken models.TokenResponse
			if err := json.Unmarshal(body, &accessToken); err != nil {
				fmt.Println("Can not unmarshal JSON")
			}

			urlStr = fmt.Sprintf("https://login.yandex.ru/info?format=json&oauth_token=%s", accessToken.AccessToken)
			req, _ = http.NewRequest("GET", urlStr, nil)
			res, _ = client.Do(req)

			body, _ = ioutil.ReadAll(res.Body)
			var info models.UserInfo
			if err := json.Unmarshal(body, &info); err != nil {
				fmt.Println("Can not unmarshal JSON")
			}
			fmt.Println(info)
		}
	}
}
