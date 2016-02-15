package main

import (
	"github.com/digitalocean/godo"
	"golang.org/x/oauth2"
	"log"
)

type TokenSource struct {
	AccessToken string
}

func (t *TokenSource) Token() (*oauth2.Token, error) {
	token := &oauth2.Token{
		AccessToken: t.AccessToken,
	}
	return token, nil
}

func main() {
	settings, err := loadSettings("dospin.json")
	if err != nil {
		log.Fatal(err)
	}
	tokenSource := &TokenSource{settings.Token}
	oauthClient := oauth2.NewClient(oauth2.NoContext, tokenSource)
	client := godo.NewClient(oauthClient)
	dm := NewDropletManager(client, settings)

	_, err = dm.SpinupMachine("minecraft")
	if err != nil {
		log.Println(err)
		return
	}
	//_, err = client.Droplets.Delete(droplet.ID)
	//if err != nil {
	//	log.Println(err)
	//}
}
