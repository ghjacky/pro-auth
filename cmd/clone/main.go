package main

import (
	"encoding/json"
	"flag"

	sdk "code.xxxxx.cn/platform/auth/sdk/golang"
	log "github.com/sirupsen/logrus"
)

var clientID = flag.Int64("id", 0, "client id")
var clientSecret = flag.String("secret", "", "client secret")
var authHost = flag.String("host", "", "auth server host")

func main() {
	flag.Parse()

	apiConfig := sdk.APIConfig{
		ClientID:     *clientID,
		ClientSecret: *clientSecret,
		APIHost:      *authHost,
	}

	authClient := sdk.NewApiAuth(&apiConfig)
	res, err := authClient.POST("client/clone", []byte(""))
	if err != nil {
		log.WithError(err).Error("clone failed!")
		return
	}
	client := sdk.Client{}
	err = json.Unmarshal(res, &client)
	if err != nil {
		log.WithError(err).Error("could not get response")
		return
	}
	log.Info("Clone Client succeed, New Client ID: %v", client.Id)
}
