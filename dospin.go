package main

import (
	"errors"
	"github.com/digitalocean/godo"
	"golang.org/x/oauth2"
	"log"
	"net"
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

	droplet, err := dm.spinupDroplet("minecraft")
	if err != nil {
		log.Println(err)
		return
	}
	_, err = client.Droplets.Delete(droplet.ID)
	if err != nil {
		log.Println(err)
	}
}

// Listens for clients on given ports to spin up droplets for the ports
type ServiceHandler struct {
	dropletManager *DropletManager
	dropletSvcCnt  map[string]int
	/*
	 This should start at 0 and should be incremented any time a cleanup check
	 shows no connections on this port. Once it reaches 5, the port forward
	 should be deleted along with that port in the map
	*/
	portConnStatus map[int]int
}

func NewServiceHandler(dropletManager *DropletManager) *ServiceHandler {
	sh := new(ServiceHandler)
	sh.dropletManager = dropletManager
	return sh
}

func (me *ServiceHandler) setupService(dropletName, port string) {
	addr, err := net.ResolveTCPAddr("tcp", "0.0.0.0:"+port)
	if err != nil {
		log.Print("Could not resolve port and listen address:", err)
		return
	}

	// listen on port
	go func() {
		l, err := net.ListenTCP("tcp", addr)
		if err != nil {
			log.Print("Could not listen for TCP connection:", err)
		} else {
			for {
				conn, err := l.AcceptTCP()
				if err != nil {
					log.Print("Could not accept TCP connection:", err)
				} else {
					// connection accepted
					// spinup droplet
					droplet, err := me.dropletManager.spinupDroplet(dropletName)

					if err == nil {
						// setup port forwarding
						ip, err := droplet.PrivateIPv4()
						if err == nil {
							setupPortForward(ip, port)
						} else {
							log.Print("Could not get private IP address for "+dropletName+": ", err)
						}
						me.dropletSvcCnt[dropletName]++
					} else {
						log.Print("Could not spin up Droplet:", err)
					}

					// close existing connection, not doing anything with it
					conn.Close()
				}
			}
		}
	}()
}

/*
  Periodically checks number of connections to each droplet and deletes
  them when they are no longer needed
*/
func (me *ServiceHandler) cleanup() {
}

type DropletManager struct {
	client   *godo.Client
	settings settings
}

func NewDropletManager(client *godo.Client, settings settings) *DropletManager {
	retval := new(DropletManager)
	retval.client = client
	retval.settings = settings
	return retval
}

/*
  Gets the Droplet if it already exists, instantiates it if it does not.
*/
func (me *DropletManager) spinupDroplet(name string) (*godo.Droplet, error) {
	if droplet, err := me.getDroplet(name); err == nil {
		return &droplet, nil
	} else {
		image, err := me.getSnapshot(name)
		if err != nil {
			return nil, err
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
			return nil, err
		}

		return droplet, nil
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
