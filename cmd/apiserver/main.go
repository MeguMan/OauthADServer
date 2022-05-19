package main

import (
	"OauthADServer/internal/app/apiserver"
	"OauthADServer/internal/app/models"
	"encoding/json"
	"fmt"
	"os"
)

func main() {
	configFile, err := os.Open("configs/config.json")
	if err != nil {
		fmt.Println(err)
	}

	cfg := models.NewGlobalConfig()
	jsonParser := json.NewDecoder(configFile)
	if err := jsonParser.Decode(cfg); err != nil {
		fmt.Println(err)
	}

	if err := apiserver.Start(cfg); err != nil {
		fmt.Println(err)
	}
}