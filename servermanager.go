package main

import "log"

const (
	SERVERMANAGER_SPINUP = iota
	SERVERMANAGER_SPINDOWN
	SERVERMANAGER_STOP
)

type ServerManager struct {
	name       string
	ports      []int
	in         chan int
	out        chan int
	usageScore int // spin down server when this reaches 0
	server     ServerHandler
}

func NewServerManager(name string, server ServerHandler, settings Settings) *ServerManager {
	sm := new(ServerManager)

	sm.in = make(chan int)
	sm.out = make(chan int)
	sm.usageScore = 5
	sm.server = server

	// find the ports associated with this server in settings
	for _, s := range settings.Services {
		if s.LogicalServer == name {
			sm.ports = append(sm.ports, s.Port)
		}
	}

	return sm
}

/*
 Serves channel requests.
*/
func (me *ServerManager) Serve() {
	for running := true; running; {
		select {
		case action := <-me.in:
			running = me.serveAction(action)
		}
	}
}

/*
 Sends the serve loop a spinup message.
*/
func (me *ServerManager) Spinup() {
	me.in <- SERVERMANAGER_SPINUP
}

/*
 Sends the serve loop a spindown message.
*/
func (me *ServerManager) Spindown() {
	me.in <- SERVERMANAGER_SPINDOWN
}

/*
 Sends the serve loop a quit message.
*/
func (me *ServerManager) Stop() {
	me.in <- SERVERMANAGER_STOP
}

func (me *ServerManager) addPortForwards(ip string) {
}

func (me *ServerManager) rmPortForwards() {
}

func (me *ServerManager) serveAction(action int) bool {
	running := true
	switch action {
	case SERVERMANAGER_SPINUP:
		ip, err := me.server.Spinup(me.name)
		if err == nil {
			me.addPortForwards(ip)
		} else {
			log.Println("ServerManager: Could not spin up "+me.name+":", err)
		}
	case SERVERMANAGER_SPINDOWN:
		err := me.server.Spindown(me.name)
		if err == nil {
			me.rmPortForwards()
		} else {
			log.Println("ServerManager: Could not spin down "+me.name+":", err)
		}
	case SERVERMANAGER_STOP:
		running = false
	}
	return running
}
