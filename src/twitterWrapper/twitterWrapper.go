package twitterWrapper

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/garyburd/go-oauth/oauth"
	"github.com/xiam/twitter"
	"io/ioutil"
	"os"
)

type TwitterConnfig struct {
	App struct {
		User string `json:"user"`
		Pass string `json:"pass"`
	} `json:"app"`
	User struct {
		Token  string `json:"token"`
		Secret string `json:"secret"`
	} `json:"user"`
}

func ReadConfig() (*TwitterConnfig, error) {
	path := os.Getenv("GO_PROJECT_ROOT")
	if path == "" {
		return nil, errors.New("No configuration available")
	}
	path = path + "/config/twitter_config.json"
	file, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var config TwitterConnfig
	json.Unmarshal(file, &config)
	fmt.Println(config)
	return &config, nil
}

func WriteConfig(conf *TwitterConnfig) error {
	path := os.Getenv("GO_PROJECT_ROOT")
	if path == "" {
		return errors.New("No configuration available")
	}
	path = path + "/config/twitter_config.json"
	b, err := json.MarshalIndent(conf, " ", "  ")
	if err != nil {
		return err
	}
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	file.Write(b)
	file.Close()
	return nil
}

func NewClient() (*twitter.Client, error) {
	conf, err := ReadConfig()
	if err != nil {
		return nil, err
	}
	client := twitter.New(&oauth.Credentials{
		conf.App.User,
		conf.App.Pass,
	})
	client.SetAuth(&oauth.Credentials{
		conf.User.Token,
		conf.User.Secret,
	})
	_, err = client.VerifyCredentials(nil)
	if err == nil {
		return client, nil
	}
	// else User is not setup, let's fix that!
	err = client.Setup()
	if err != nil {
		return nil, err
	}
	// Next we would want to write the credentials
	userToken := client.Auth.Token
	userSecret := client.Auth.Secret
	conf.User.Token = userToken
	conf.User.Secret = userSecret
	WriteConfig(conf)
	return client, nil
}
