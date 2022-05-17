package main

import (
	"OauthADServer/internal/app/apiserver"
	"encoding/json"
	"log"
	"os"
)

func main() {
	serverConfig := apiserver.NewConfig()
	configFile, err := os.Open("configs/config.json")
	if err != nil {
		log.Fatal(err)
	}

	jsonParser := json.NewDecoder(configFile)
	if err := jsonParser.Decode(serverConfig); err != nil {
		log.Fatal(err)
	}

	if err := apiserver.Start(serverConfig); err != nil {
		log.Fatal(err)
	}
}