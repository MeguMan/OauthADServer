package apiserver

import (
	"OauthADServer/internal/app/cache"
	"OauthADServer/internal/app/helpers"
	"OauthADServer/internal/app/models"
	"OauthADServer/internal/app/storage"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

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

			//urlStr := "https://oauth.yandex.ru/token"
			//
			//client := &http.Client{}
			//req, _ := http.NewRequest("POST", urlStr, strings.NewReader(data.Encode()))
			//req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
			//res, _ := client.Do(req)

			//body, _ := ioutil.ReadAll(res.Body)
			body := []byte(`{"access_token":"aoaspdfjaspijmvpaksjpdif"}`)
			var accessToken models.TokenResponse
			if err := json.Unmarshal(body, &accessToken); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			//urlStr = fmt.Sprintf("https://login.yandex.ru/info?format=json&oauth_token=%s", accessToken.AccessToken)
			//req, _ = http.NewRequest("GET", urlStr, nil)
			//res, _ = client.Do(req)

			//body, _ = ioutil.ReadAll(res.Body)
			body = []byte(`{"id":"12345","default_email":"asd@mail.ru"}`)
			var info models.YandexUserInfo
			if err := json.Unmarshal(body, &info); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			//val, _ := s.cache.Get(state)

			//val, exists := s.cache.Get(state)
			//if !exists {
			//	http.Error(w, "state not found in cache", http.StatusInternalServerError)
			//	return
			//}

			_, err := s.buildJwt(ctx, info.Id, info.DefaultEmail, storage.ExternalServiceTypeYandex)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			err = s.storage.CreateLog(ctx, storage.ExternalServiceTypeYandex, storage.LoginStatusOk)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			//http.Redirect(w, r, fmt.Sprintf("%s?access_token=%s", val.RedirectUri, jwt.AccessToken), http.StatusPermanentRedirect)
			w.WriteHeader(http.StatusOK)
		}
	}
}
