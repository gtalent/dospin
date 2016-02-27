/*
   Copyright 2016 gtalent2@gmail.com

   This Source Code Form is subject to the terms of the Mozilla Public
   License, v. 2.0. If a copy of the MPL was not distributed with this
   file, You can obtain one at http://mozilla.org/MPL/2.0/.
*/
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
	done       chan interface{}
	usageScore int // spin down server when this reaches 0
	server     ServerHandler
}

func NewServerManager(name string, server ServerHandler, settings Settings) *ServerManager {
	sm := new(ServerManager)

	sm.name = name
	sm.in = make(chan int)
	sm.done = make(chan interface{})
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
	me.done <- 42
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

func (me *ServerManager) Done() {
	<-me.done
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
			log.Println("ServerManager: Got IP for", me.name, ":", ip)
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
