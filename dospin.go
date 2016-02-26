/*
   Copyright 2016 gtalent2@gmail.com

   This Source Code Form is subject to the terms of the Mozilla Public
   License, v. 2.0. If a copy of the MPL was not distributed with this
   file, You can obtain one at http://mozilla.org/MPL/2.0/.
*/
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

	ip, err := dm.SpinupMachine("minecraft")
	if err != nil {
		log.Println("Error:", err)
		return
	}
	log.Println("IP: " + ip)

	if err := dm.SpindownMachine("minecraft"); err != nil {
		log.Println("Error:", err)
		return
	}
	//_, err = client.Droplets.Delete(droplet.ID)
	//if err != nil {
	//	log.Println(err)
	//}
}
