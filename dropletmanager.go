package main

import (
	"errors"
	"github.com/digitalocean/godo"
	"log"
	"time"
)

const DROPLET_NS = "dospin-"

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
		// create the droplet
		image, err := me.getSnapshot(name)
		if err != nil {
			return "", err
		}
		vd := me.settings.VirtualDroplets[name]
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
		// wait until machine is ready
		for {
			d, _, err := me.client.Droplets.Get(droplet.ID)
			if err != nil {
				log.Println(err)
			} else if d.Status == "active" {
				break
			}
			time.Sleep(500 * time.Millisecond)
		}
		if err != nil {
			return "", err
		}
		log.Println("Spinup: Created " + name)

		// get the private IP and return it
		droplet, _, err = me.client.Droplets.Get(droplet.ID)
		if err != nil {
			return "", err
		}
		return droplet.PrivateIPv4()
	}
}

func (me *DropletManager) SpindownMachine(name string) error {
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
				log.Println("Power down of ", name, " failed: ", err)
			}
		}
		log.Println("Spindown: Powered down " + name)
	}

	// snapshot existing droplet
	log.Println("Spindown: Snapshoting " + name)
	action, _, err := me.client.DropletActions.Snapshot(droplet.ID, DROPLET_NS+name)
	if err != nil || !me.actionWait(action.ID) {
		return err
	}
	log.Println("Spindown: Snapshoted " + name)

	// delete droplet
	log.Println("Spindown: Deleting " + name)
	_, err = me.client.Droplets.Delete(droplet.ID)
	if err != nil {
		return err
	}
	log.Println("Spindown: Deleted " + name)

	return err
}

func (me *DropletManager) getDroplet(name string) (godo.Droplet, error) {
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

func (me *DropletManager) getSnapshot(name string) (godo.Image, error) {
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

func (me *DropletManager) actionWait(actionId int) bool {
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
