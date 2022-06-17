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

			jwt, err := s.buildJwt(ctx, string(info.Id), primaryEmail, storage.ExternalServiceTypeGithub)
			if err != nil {
				http.Error(w, "buildJwt", http.StatusInternalServerError)
				return
			}

			err = s.storage.CreateLog(ctx, storage.ExternalServiceTypeGithub, storage.LoginStatusOk)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			http.Redirect(w, r, fmt.Sprintf("%s?access_token=%s", val.RedirectUri, jwt.AccessToken), http.StatusPermanentRedirect)
		}
	}
}
