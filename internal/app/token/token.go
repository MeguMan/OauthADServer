package token

import (
	"OauthADServer/internal/app/storage"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"time"
)

type Manager struct {
	signingKey string
}

type JWT struct {
	AccessToken string `json:"access_token"`
}

func NewManager(signingKey string) *Manager {
	return &Manager{signingKey: signingKey}
}

func (m *Manager) NewJWT(employeeId, externalServiceId string, externalServiceType storage.ExternalServiceType, ttl time.Duration) (*JWT, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"exp": time.Now().Add(ttl).Unix(),
		"employeeId": employeeId,
		"externalServiceId": externalServiceId,
		"externalServiceType": externalServiceType,
	})

	accessToken, err := token.SignedString([]byte(m.signingKey))
	if err != nil {
		return nil, err
	}

	return &JWT{
		AccessToken: accessToken,
	}, nil
}

func (m *Manager) Parse(accessToken string) (*storage.Link, error) {
	token, err := jwt.Parse(accessToken, func(token *jwt.Token) (i interface{}, err error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(m.signingKey), nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("error get user claims from token")
	}

	payload := &storage.Link{
		EmployeeId:            claims["employeeId"].(string),
		ExternalServiceId:     claims["externalServiceId"].(string),
		ExternalServiceTypeId: storage.ExternalServiceType(claims["externalServiceType"].(float64)),
	}

	return payload, nil
}