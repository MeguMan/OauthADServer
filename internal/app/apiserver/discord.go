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

func (s *server) HandleDiscordAuth() func(w http.ResponseWriter, r *http.Request) {
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

		http.Redirect(w, r, fmt.Sprintf("https://canary.discord.com/api/oauth2/authorize?client_id=%s&redirect_uri=http://diplom.com/discord/redirect&response_type=code&scope=identify email&state=%s", s.discCfg.ClientId, oauthCode), http.StatusPermanentRedirect)
	}
}

func (s *server) HandleDiscordRedirect() func(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	return func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		state := r.URL.Query().Get("state")
		if code != "" && state != ""{
			data := url.Values{}
			data.Set("client_id", s.discCfg.ClientId)
			data.Set("client_secret", s.discCfg.ClientSecret)
			data.Set("grant_type", "authorization_code")
			data.Set("code", code)
			data.Set("redirect_uri", "http://diplom.com/discord/redirect")


			urlStr := "https://discord.com/api/oauth2/token"

			client := &http.Client{}
			req, _ := http.NewRequest("POST", urlStr, strings.NewReader(data.Encode()))
			req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
			res, _ := client.Do(req)

			body, _ := ioutil.ReadAll(res.Body)
			var accessToken models.DiscordToken
			if err := json.Unmarshal(body, &accessToken); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			req, _ = http.NewRequest("GET", "https://discord.com/api/users/@me", nil)
			req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", accessToken.AccessToken))
			res, _ = client.Do(req)

			body, _ = ioutil.ReadAll(res.Body)
			var info models.DiscordUserInfo
			if err := json.Unmarshal(body, &accessToken); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			val, exists := s.cache.Get(state)
			if !exists {
				http.Error(w, "state not found in cache", http.StatusInternalServerError)
				return
			}

			jwt, err := s.buildJwt(ctx, info.Id, info.Email, storage.ExternalServiceTypeDiscord)
			if err != nil {
				http.Error(w, "buildJwt", http.StatusInternalServerError)
				return
			}

			err = s.storage.CreateLog(ctx, storage.ExternalServiceTypeDiscord, storage.LoginStatusOk)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			http.Redirect(w, r, fmt.Sprintf("%s?access_token=%s", val.RedirectUri, jwt.AccessToken), http.StatusPermanentRedirect)
		}
	}
}
