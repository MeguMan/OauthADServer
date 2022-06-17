package apiserver

import (
	"OauthADServer/internal/app/cache"
	"OauthADServer/internal/app/helpers"
	"OauthADServer/internal/app/models"
	"OauthADServer/internal/app/storage"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"
)

func (s *server) HandleMailAuth() func(w http.ResponseWriter, r *http.Request) {
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

		http.Redirect(w, r, fmt.Sprintf("https://oauth.mail.ru/login?client_id=%s&response_type=code&scope=userinfo&redirect_uri=http://diplom.com/mail/redirect&state=%s", s.mailCfg.ClientId, oauthCode), http.StatusPermanentRedirect)
	}
}

func (s *server) HandleMailRedirect() func(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	return func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		state := r.URL.Query().Get("state")
		if code != "" && state != ""{
			data := url.Values{}
			data.Set("code", code)
			data.Set("grant_type", "authorization_code")
			data.Set("redirect_uri", "http://diplom.com/mail/redirect")

			urlStr := "https://oauth.mail.ru/token"

			client := &http.Client{}
			req, _ := http.NewRequest("POST", urlStr, strings.NewReader(data.Encode()))
			req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
			req.Header.Add("Authorization", fmt.Sprintf("Basic %s", basicAuth(s.mailCfg.ClientId, s.mailCfg.ClientSecret)))
			res, _ := client.Do(req)

			body, _ := ioutil.ReadAll(res.Body)
			var accessToken models.MailToken
			if err := json.Unmarshal(body, &accessToken); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			urlStr = fmt.Sprintf("https://oauth.mail.ru/userinfo?access_token=%s", accessToken.AccessToken)
			req, _ = http.NewRequest("GET", urlStr, nil)
			res, _ = client.Do(req)

			body, _ = ioutil.ReadAll(res.Body)
			var info models.MailUserInfo
			if err := json.Unmarshal(body, &info); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			val, exists := s.cache.Get(state)
			if !exists {
				http.Error(w, "state not found in cache", http.StatusInternalServerError)
				return
			}

			jwt, err := s.buildJwt(ctx, info.Id, info.Email, storage.ExternalServiceTypeMail)
			if err != nil {
				http.Error(w, "buildJwt", http.StatusInternalServerError)
				return
			}

			err = s.storage.CreateLog(ctx, storage.ExternalServiceTypeMail, storage.LoginStatusOk)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			http.Redirect(w, r, fmt.Sprintf("%s?access_token=%s", val.RedirectUri, jwt.AccessToken), http.StatusPermanentRedirect)
		}
	}
}

