package main

import (
	"errors"
	"github.com/digitalocean/godo"
)

type DropletManager struct {
	client   *godo.Client
	settings Settings
}

func NewDropletManager(client *godo.Client, settings Settings) *DropletManager {
	retval := new(DropletManager)
	retval.client = client
	retval.settings = settings
	return retval
}

/*
  Gets the Droplet if it already exists, instantiates it if it does not.
*/
func (me *DropletManager) SpinupMachine(name string) (string, error) {
	if droplet, err := me.getDroplet(name); err == nil {
		return droplet.PrivateIPv4()
	} else {
		image, err := me.getSnapshot(name)
		if err != nil {
			return "", err
		}
		vd := me.settings.VirtualDroplets[name]
		createRequest := &godo.DropletCreateRequest{
			Name:   name,
			Region: vd.Region,
			Size:   vd.Size,
			Image: godo.DropletCreateImage{
				ID: image.ID,
			},
		}

		droplet, _, err := me.client.Droplets.Create(createRequest)
		if err != nil {
			return "", err
		}

		return droplet.PrivateIPv4()
	}
}

func (me *DropletManager) getDroplet(name string) (godo.Droplet, error) {
	name = "dospin:" + name
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
	return droplet, errors.New("Could not find droplet")
}

func (me *DropletManager) getSnapshot(name string) (godo.Image, error) {
	name = "dospin:" + name
	page := 0
	perPage := 200
	var image godo.Image
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
			break
		}
	}
	return image, errors.New("Could not find image")
}
