package apiserver

import (
	"OauthADServer/internal/app/storage"
	"OauthADServer/internal/app/token"
	"fmt"
	"net/http"
	"strings"
)

type authenticationMiddleware struct {
	tokenManager *token.Manager
	storage storage.Facade
}

func (amw *authenticationMiddleware) Middleware(next http.Handler) http.Handler {
	//ctx := context.Background()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "empty auth header", http.StatusUnauthorized)
			return
		}

		headerParts := strings.Split(authHeader, " ")
		if len(headerParts) != 2 {
			http.Error(w, "invalid auth header", http.StatusUnauthorized)
			return
		}

		authMethod := headerParts[0]
		if authMethod != "Bearer" {
			http.Error(w, "invalid auth method", http.StatusUnauthorized)
			return
		}

		authToken := headerParts[1]
		if authToken == "" {
			http.Error(w, "invalid auth token", http.StatusUnauthorized)
			return
		}

		_, err := amw.tokenManager.Parse(authToken)
		if err != nil {
			msg := fmt.Sprintf("error parsing token %s", authToken)
			http.Error(w, msg, http.StatusUnauthorized)
			return
		}

		//employeeId, err := amw.storage.GetEmployeeId(ctx, payload.ExternalServiceId, payload.ExternalServiceTypeId)
		//if err != nil {
		//	http.Error(w, "employeeId not found", http.StatusUnauthorized)
		//	return
		//}
		//
		//if employeeId != payload.EmployeeId {
		//	http.Error(w, "incorrect employeeId", http.StatusUnauthorized)
		//	return
		//}

		next.ServeHTTP(w, r)
	})
}
