/*
   Copyright 2016 gtalent2@gmail.com

   This Source Code Form is subject to the terms of the Mozilla Public
   License, v. 2.0. If a copy of the MPL was not distributed with this
   file, You can obtain one at http://mozilla.org/MPL/2.0/.
*/
package main

import (
	"errors"
	"github.com/digitalocean/godo"
	"golang.org/x/oauth2"
	"log"
	"time"
)

const DROPLET_NS = "dospin-"

type tokenSource struct {
	AccessToken string
}

func (t *tokenSource) Token() (*oauth2.Token, error) {
	token := &oauth2.Token{
		AccessToken: t.AccessToken,
	}
	return token, nil
}

type DropletHandler struct {
	client   *godo.Client
	settings Settings
}

func NewDropletHandler(settings Settings) *DropletHandler {
	retval := new(DropletHandler)
	retval.settings = settings

	// setup DO client
	tokenSource := &tokenSource{settings.ApiToken}
	oauthClient := oauth2.NewClient(oauth2.NoContext, tokenSource)
	retval.client = godo.NewClient(oauthClient)

	return retval
}

/*
  Gets the Droplet if it already exists, instantiates it if it does not.
*/
func (me *DropletHandler) Spinup(name string) (string, error) {
	if droplet, err := me.getDroplet(name); err == nil {
		return droplet.PrivateIPv4()
	} else {
		// create the droplet
		image, err := me.getSnapshot(name)
		if err != nil {
			return "", err
		}
		vd := me.settings.Servers[name]
		createRequest := &godo.DropletCreateRequest{
			Name:              DROPLET_NS + name,
			Region:            vd.Region,
			Size:              vd.Size,
			PrivateNetworking: true,
			Image: godo.DropletCreateImage{
				ID: image.ID,
			},
		}

		log.Println("Spinup: Creating " + name)
		droplet, _, err := me.client.Droplets.Create(createRequest)
		if err != nil {
			log.Println(err)
			if droplet == nil {
				return "", err
			}
		}
		// wait until machine is ready
		for {
			d, _, err := me.client.Droplets.Get(droplet.ID)
			if err != nil {
				log.Println(err)
				return "", err
			} else if d.Status == "active" {
				break
			}
			time.Sleep(500 * time.Millisecond)
		}
		log.Println("Spinup: Created " + name)

		// delete the image
		log.Println("Spinup: Deleting image " + name)
		_, err = me.client.Images.Delete(image.ID)
		if err != nil {
			log.Println("Spinup: Could not delete image: ", err)
		}
		log.Println("Spinup: Deleted image " + name)

		// get the private IP and return it
		droplet, _, err = me.client.Droplets.Get(droplet.ID)
		if err != nil {
			return "", err
		}
		return droplet.PrivateIPv4()
	}
}

func (me *DropletHandler) Spindown(name string) error {
	droplet, err := me.getDroplet(name)
	if err != nil {
		return err
	}

	// power off
	if droplet.Status != "off" {
		log.Println("Spindown: Powering down " + name)
		// wait until machine is off
		for {
			droplet, err = me.getDroplet(name)
			if err != nil {
				log.Println(err)
			} else if droplet.Status == "off" {
				break
			}
			time.Sleep(100 * time.Millisecond)
			_, _, err = me.client.DropletActions.Shutdown(droplet.ID)
			if err != nil {
				log.Println("Spindown: Power down of ", name, " failed: ", err)
			}
		}
		log.Println("Spindown: Powered down " + name)
	}

	// snapshot existing droplet
	log.Println("Spindown: Creating image " + name)
	action, _, err := me.client.DropletActions.Snapshot(droplet.ID, DROPLET_NS+name)
	if err != nil || !me.actionWait(action.ID) {
		return err
	}
	log.Println("Spindown: Creating imaged " + name)

	// delete droplet
	log.Println("Spindown: Deleting droplet " + name)
	_, err = me.client.Droplets.Delete(droplet.ID)
	if err != nil {
		return err
	}
	log.Println("Spindown: Deleted droplet " + name)

	return err
}

func (me *DropletHandler) getDroplet(name string) (godo.Droplet, error) {
	name = DROPLET_NS + name
	page := 0
	perPage := 200
	var droplet godo.Droplet
	for {
		page++
		// get list of droplets
		opt := &godo.ListOptions{
			Page:    page,
			PerPage: perPage,
		}
		images, _, err := me.client.Droplets.List(opt)
		if err != nil {
			break
		}
		// find droplet
		for _, a := range images {
			if a.Name == name {
				return a, nil
			}
		}
		// check next page?
		if len(images) < perPage {
			break
		}
	}
	return droplet, errors.New("Could not find droplet: " + name)
}

func (me *DropletHandler) getSnapshot(name string) (godo.Image, error) {
	name = DROPLET_NS + name
	page := 0
	perPage := 200
	var image godo.Image
	var err error
	for {
		page++

		// get list of images
		opt := &godo.ListOptions{
			Page:    page,
			PerPage: perPage,
		}
		images, _, err := me.client.Images.ListUser(opt)
		if err != nil {
			break
		}

		// find image
		for _, a := range images {
			if a.Name == name {
				return a, nil
			}
		}

		// check next page?
		if len(images) < perPage {
			err = errors.New("Could not find image: " + name)
			break
		}
	}
	return image, err
}

func (me *DropletHandler) actionWait(actionId int) bool {
	for {
		a, _, err := me.client.Actions.Get(actionId)
		if err != nil {
			log.Println("Action failed: ", err)
			return false
		} else if a.Status == "completed" {
			return true
		} else if a.Status == "errored" {
			log.Println("Action failed: ", a.Type, " on ", a.ResourceID)
			return false
		}
		time.Sleep(1000 * time.Millisecond)
	}
}
