package apiserver

import (
	"net/http"
)

func Start(config *Config) error{
	server := NewServer(config)
	return http.ListenAndServe(":8080", server)
}