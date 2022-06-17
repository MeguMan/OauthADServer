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
	"time"
)

func (s *server) HandleFacebookAuth() func(w http.ResponseWriter, r *http.Request) {
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

		http.Redirect(w, r, fmt.Sprintf("https://www.facebook.com/dialog/oauth?client_id=%s&redirect_uri=http://diplom.com/facebook/redirect&scope=email&response_type=code&state=%s", s.discCfg.ClientId, oauthCode), http.StatusPermanentRedirect)
	}
}

func (s *server) HandleFacebookRedirect() func(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	return func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		state := r.URL.Query().Get("state")
		if code != "" && state != ""{
			urlStr := fmt.Sprintf("https://graph.facebook.com/oauth/access_token?client_id=%s&client_secret=%s&redirect_uri=%s&code=%s", s.fcbCfg.ClientId, s.fcbCfg.ClientSecret, "http://diplom.com/facebook/redirect", code)

			client := &http.Client{}
			req, _ := http.NewRequest("GET", urlStr, nil)
			//req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
			res, _ := client.Do(req)

			body, _ := ioutil.ReadAll(res.Body)
			var accessToken models.FacebookToken
			if err := json.Unmarshal(body, &accessToken); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			req, _ = http.NewRequest("GET", fmt.Sprintf("https://graph.facebook.com/me?access_token=%s&fields=id,email", accessToken.AccessToken), nil)
			res, _ = client.Do(req)

			body, _ = ioutil.ReadAll(res.Body)
			var info models.FacebookUserInfo
			if err := json.Unmarshal(body, &accessToken); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			val, exists := s.cache.Get(state)
			if !exists {
				http.Error(w, "state not found in cache", http.StatusInternalServerError)
				return
			}

			jwt, err := s.buildJwt(ctx, info.Id, info.Email, storage.ExternalServiceTypeFacebook)
			if err != nil {
				http.Error(w, "buildJwt", http.StatusInternalServerError)
				return
			}

			err = s.storage.CreateLog(ctx, storage.ExternalServiceTypeFacebook, storage.LoginStatusOk)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			http.Redirect(w, r, fmt.Sprintf("%s?access_token=%s", val.RedirectUri, jwt.AccessToken), http.StatusPermanentRedirect)
		}
	}
}

