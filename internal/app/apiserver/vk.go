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

		http.Redirect(w, r, fmt.Sprintf("https://oauth.vk.com/authorize?client_id=%s&redirect_uri=http://diplom.com/vk/redirect&response_type=code&scope=email&state=%s", s.vkCfg.ClientId, oauthCode), http.StatusPermanentRedirect)
	}
}

func (s *server) HandleVkRedirect() func(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	return func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		state := r.URL.Query().Get("state")
		if code != "" && state != ""{
			urlStr := fmt.Sprintf("https://oauth.vk.com/access_token?client_id=%s&client_secret=%s&redirect_uri=%s&code=%s", s.vkCfg.ClientId, s.vkCfg.ClientSecret, "http://diplom.com/vk/redirect", code)

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

			err = s.storage.CreateLog(ctx, storage.ExternalServiceTypeVkontakte, storage.LoginStatusOk)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			http.Redirect(w, r, fmt.Sprintf("%s?access_token=%s", val.RedirectUri, jwt.AccessToken), http.StatusPermanentRedirect)
		}
	}
}
