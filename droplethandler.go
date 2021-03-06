/*
   Copyright 2016-2017 gtalent2@gmail.com

   This Source Code Form is subject to the terms of the Mozilla Public
   License, v. 2.0. If a copy of the MPL was not distributed with this
   file, You can obtain one at http://mozilla.org/MPL/2.0/.
*/

package main

import (
	"context"
	"errors"
	"github.com/digitalocean/godo"
	"golang.org/x/oauth2"
	"log"
	"time"
)

const dropletNS = "dospin-"

type tokenSource struct {
	AccessToken string
}

func (t *tokenSource) Token() (*oauth2.Token, error) {
	token := &oauth2.Token{
		AccessToken: t.AccessToken,
	}
	return token, nil
}

type dropletHandler struct {
	client   *godo.Client
	ctx      context.Context
	settings settings
}

func newDropletHandler(settings settings) *dropletHandler {
	retval := new(dropletHandler)
	retval.settings = settings
	retval.ctx = context.Background()

	// setup DO client
	tokenSource := &tokenSource{settings.APIToken}
	oauthClient := oauth2.NewClient(oauth2.NoContext, tokenSource)
	retval.client = godo.NewClient(oauthClient)

	return retval
}

/*
  Gets the Droplet if it already exists, instantiates it if it does not.
*/
func (me *dropletHandler) Spinup(name string) (string, error) {
	vd := me.settings.Servers[name]
	if droplet, err := me.getDroplet(name); err == nil {
		if vd.UsePublicIP {
			return droplet.PublicIPv4()
		}
		return droplet.PrivateIPv4()
	}
	// create the droplet
	var image godo.DropletCreateImage

	if vd.ImageSlug == "" {
		snapshot, err := me.getSnapshot(name)
		if err != nil {
			return "", err
		}
		image = godo.DropletCreateImage{
			ID: snapshot.ID,
		}
	} else {
		image = godo.DropletCreateImage{
			Slug: vd.ImageSlug,
		}
	}

	// determine droplet size
	var size string
	if vd.InitialSize != "" {
		size = vd.InitialSize
	} else {
		size = vd.Size
	}

	createRequest := &godo.DropletCreateRequest{
		Name:              dropletNS + name,
		Region:            vd.Region,
		Size:              size,
		PrivateNetworking: true,
		SSHKeys:           me.sshKeys(vd.SSHKeys),
		Volumes:           me.volumes(vd.Volumes),
		UserData:          vd.UserData,
		Image:             image,
	}

	log.Println("Spinup: Creating " + name)
	droplet, _, err := me.client.Droplets.Create(me.ctx, createRequest)
	if err != nil {
		log.Println(err)
		if droplet == nil {
			return "", err
		}
	}
	// wait until machine is ready
	for {
		d, _, err := me.client.Droplets.Get(me.ctx, droplet.ID)
		if err != nil {
			log.Println(err)
			return "", err
		} else if d.Status == "active" {
			break
		}
		time.Sleep(500 * time.Millisecond)
	}
	log.Println("Spinup: Created " + name)

	// resize if necessary
	if vd.InitialSize != "" && vd.InitialSize != vd.Size {
		// power off
		me.poweroff(name)

		// resize
		log.Println("Spinup: Resizing " + name)
		action, _, err := me.client.DropletActions.Resize(me.ctx, droplet.ID, vd.Size, false)
		if err != nil || !me.actionWait(action.ID) {
			return "", err
		}
		log.Println("Spinup: Resized " + name)

		// power back on
		log.Println("Spinup: Powering on " + name)
		action, _, err = me.client.DropletActions.PowerOn(me.ctx, droplet.ID)
		if err != nil || !me.actionWait(action.ID) {
			return "", err
		}
		log.Println("Spinup: Powered on " + name)
	}

	// delete the image
	if image.ID > 0 {
		log.Println("Spinup: Deleting image " + name)
		_, err = me.client.Images.Delete(me.ctx, image.ID)
		if err != nil {
			log.Println("Spinup: Could not delete image: ", err)
		} else {
			log.Println("Spinup: Deleted image " + name)
		}
	}

	// get the private IP and return it

	// get new copy of droplet that has IP
	droplet, _, err = me.client.Droplets.Get(me.ctx, droplet.ID)
	if err == nil {
		if vd.UsePublicIP {
			return droplet.PublicIPv4()
		}
		return droplet.PrivateIPv4()
	}
	return "", err
}

func (me *dropletHandler) Spindown(name string) error {
	droplet, err := me.getDroplet(name)
	if err != nil {
		// droplet not existing is not an error
		return nil
	}

	// power off
	err = me.poweroff(name)
	if err != nil {
		return err
	}

	// snapshot existing droplet
	if me.settings.Servers[name].UsePersistentImage {
		log.Println("Spindown: Creating image " + name)
		action, _, err := me.client.DropletActions.Snapshot(me.ctx, droplet.ID, dropletNS+name)
		if err != nil || !me.actionWait(action.ID) {
			return err
		}
		log.Println("Spindown: Created image " + name)
	}

	// delete droplet
	log.Println("Spindown: Deleting droplet " + name)
	_, err = me.client.Droplets.Delete(me.ctx, droplet.ID)
	if err != nil {
		return err
	}
	log.Println("Spindown: Deleted droplet " + name)

	return err
}

func (me *dropletHandler) poweroff(name string) error {
	droplet, err := me.getDroplet(name)
	if err != nil {
		return err
	}
	if droplet.Status != "off" {
		log.Println("Powering down " + name)
		// wait until machine is off
		for {
			droplet, err = me.getDroplet(name)
			if err != nil {
				log.Println("Power down of", name, "failed:", err)
				if droplet.ID < 1 {
					return err
				}
			} else if droplet.Status == "off" {
				break
			}
			time.Sleep(100 * time.Millisecond)
			_, _, err = me.client.DropletActions.Shutdown(me.ctx, droplet.ID)
			if err != nil {
				log.Println("Power down of", name, "failed:", err)
			}
		}
		log.Println("Powered down", name)
	}
	return err
}

func (me *dropletHandler) getDroplet(name string) (godo.Droplet, error) {
	name = dropletNS + name
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
		images, _, err := me.client.Droplets.List(me.ctx, opt)
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

func (me *dropletHandler) getSnapshot(name string) (godo.Image, error) {
	name = dropletNS + name
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
		images, _, err := me.client.Images.ListUser(me.ctx, opt)
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

func (me *dropletHandler) actionWait(actionID int) bool {
	for {
		a, _, err := me.client.Actions.Get(me.ctx, actionID)
		if err != nil {
			log.Println("Action retrieval failed: ", err)
		} else if a.Status == "completed" {
			return true
		} else if a.Status == "errored" {
			log.Println("Action failed: ", a.Type, " on ", a.ResourceID)
			return false
		}
		time.Sleep(1000 * time.Millisecond)
	}
}

func (me *dropletHandler) sshKeys(keyNames []string) []godo.DropletCreateSSHKey {
	// build key map
	page := 0
	perPage := 200
	keyMap := make(map[string]string)
	for {
		page++
		opt := &godo.ListOptions{
			Page:    page,
			PerPage: perPage,
		}
		keys, _, err := me.client.Keys.List(me.ctx, opt)
		if err != nil {
			break
		}

		for _, v := range keys {
			keyMap[v.Name] = v.Fingerprint
		}

		// check next page?
		if len(keys) < perPage {
			break
		}
	}

	// build output key list
	var out []godo.DropletCreateSSHKey
	for _, kn := range keyNames {
		fp := keyMap[kn]
		out = append(out, godo.DropletCreateSSHKey{Fingerprint: fp})
	}
	return out
}

func (me *dropletHandler) volumes(names []string) []godo.DropletCreateVolume {
	var out []godo.DropletCreateVolume
	for _, name := range names {
		out = append(out, godo.DropletCreateVolume{Name: name})
	}
	return out
}
