package apiserver

type Config struct {
	YandexClientId string `json:"yandex_client_id"`
	YandexClientSecret string `json:"yandex_client_secret"`
}

func NewConfig() *Config {
	return &Config{}
}