package apiserver

import (
	"OauthADServer/internal/app/cache"
	"OauthADServer/internal/app/helpers"
	"OauthADServer/internal/app/models"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

func (s *server) HandleOdnoklassnikiAuth() func(w http.ResponseWriter, r *http.Request) {
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

		http.Redirect(w, r, fmt.Sprintf("https://connect.ok.ru/oauth/authorize?client_id=%s&scope=GET_EMAIL&response_type=code&redirect_uri=http://diplom.com/odnoklassniki/redirect&layout=w&state=%s", s.odnklsCfg.ClientId, oauthCode), http.StatusPermanentRedirect)
	}
}

func (s *server) HandleOdnoklassnikiRedirect() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		state := r.URL.Query().Get("state")
		if code != "" && state != ""{
			urlStr := fmt.Sprintf("https://api.ok.ru/oauth/token.do?code=%s&client_id=%s&client_secret=%s&redirect_uri=%s&grant_type=authorization_code", code, s.odnklsCfg.ClientId, s.odnklsCfg.ClientSecret, "http://diplom.com/odnoklassniki/redirect")
			client := &http.Client{}
			req, _ := http.NewRequest("POST", urlStr, nil)
			res, _ := client.Do(req)

			body, _ := ioutil.ReadAll(res.Body)
			var accessToken models.OdnoklassnikiToken
			if err := json.Unmarshal(body, &accessToken); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			urlStr = fmt.Sprintf("http://api.odnoklassniki.ru/fb.do?method=%s&access_token=%s", accessToken.AccessToken)
			req, _ = http.NewRequest("GET", urlStr, nil)
			res, _ = client.Do(req)
			body, _ = ioutil.ReadAll(res.Body)
			var info models.MailUserInfo
			if err := json.Unmarshal(body, &info); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}
	}
}
