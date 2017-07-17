package godoSdk

import (
	"context"
	"github.com/digitalocean/godo"
	"golang.org/x/oauth2"
	"os"
	"fmt"
)

type Sdk struct {
	Client *godo.Client
}

func NewSdk() (*Sdk, error) {
	sdk := &Sdk{}
	pat := GetToken()
	if pat == "" {
		return nil, fmt.Errorf("Empty access token. Remember to export DIGITALOCEAN_ACCESS_TOKEN='mytoken'")
	}
	tokenSource := &TokenSource{
		AccessToken: pat,
	}
	oauthClient := oauth2.NewClient(context.Background(), tokenSource)
	client := godo.NewClient(oauthClient)
	sdk.Client = client
	return sdk, nil
}

func GetToken() string {
	return os.Getenv("DIGITALOCEAN_ACCESS_TOKEN")
}

type TokenSource struct {
	AccessToken string
}

func (t *TokenSource) Token() (*oauth2.Token, error) {
	token := &oauth2.Token{
		AccessToken: t.AccessToken,
	}
	return token, nil
}
