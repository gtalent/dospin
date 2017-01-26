/*
   Copyright 2016-2017 gtalent2@gmail.com

   This Source Code Form is subject to the terms of the Mozilla Public
   License, v. 2.0. If a copy of the MPL was not distributed with this
   file, You can obtain one at http://mozilla.org/MPL/2.0/.
*/
package main

import (
	"log"
	"net"
	"strconv"
)

const (
	SERVERMANAGER_SPINUP = iota
	SERVERMANAGER_SPINDOWN
	SERVERMANAGER_STOP
)

type serverManagerEvent struct {
	eventType int
	tcpConn   *net.TCPConn
}

type ServerHandler interface {
	// Takes snapshot name, and returns the IP to connect to.
	Spinup(name string) (string, error)
	Spindown(name string) error
}

type ServerManager struct {
	name       string
	ports      []int
	gatewayInt string
	in         chan serverManagerEvent
	done       chan interface{}
	usageScore int // spin down server when this reaches 0
	server     ServerHandler
}

func NewServerManager(name string, server ServerHandler, settings Settings) *ServerManager {
	sm := new(ServerManager)

	sm.name = name
	sm.ports = settings.Servers[name].Ports
	sm.gatewayInt = settings.GatewayInterface
	sm.in = make(chan serverManagerEvent)
	sm.done = make(chan interface{})
	sm.usageScore = 5
	sm.server = server

	return sm
}

/*
 Serves channel requests.
*/
func (me *ServerManager) Serve() {
	// TODO: see if server is currently up, and setup port forwarding if so

	// event loop
	for running := true; running; {
		select {
		case action := <-me.in:
			running = me.serveAction(action)
		}
	}

	// notify done
	me.done <- 42
}

/*
 Sends the serve loop a spinup message.
*/
func (me *ServerManager) Spinup(c *net.TCPConn) {
	me.in <- serverManagerEvent{eventType: SERVERMANAGER_SPINUP, tcpConn: c}
}

/*
 Sends the serve loop a spindown message.
*/
func (me *ServerManager) Spindown(c *net.TCPConn) {
	me.in <- serverManagerEvent{eventType: SERVERMANAGER_SPINDOWN}
}

/*
 Sends the serve loop a quit message.
*/
func (me *ServerManager) Stop() {
	me.in <- serverManagerEvent{eventType: SERVERMANAGER_STOP}
}

func (me *ServerManager) Done() {
	<-me.done
}

func (me *ServerManager) addPortForwards(localIp, remoteIp string) {
	log.Println("Ports:", me.ports)
	for _, p := range me.ports {
		port := strconv.Itoa(p)
		addPortForward(me.name, me.gatewayInt, localIp, remoteIp, port)
	}
}

func (me *ServerManager) rmPortForwards() {
	rmPortForward(me.name)
}

func (me *ServerManager) setupListener(port int) {
	portStr := strconv.Itoa(port)
	addr, err := net.ResolveTCPAddr("tcp", "0.0.0.0:"+portStr)
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

					// spinup machine
					me.Spinup(conn)

					// close existing connection, not doing anything with it
					conn.Close()
				}
			}
		}
	}()
}

func (me *ServerManager) serveAction(event serverManagerEvent) bool {
	running := true
	switch event.eventType {
	case SERVERMANAGER_SPINUP:
		targetIp, err := me.server.Spinup(me.name)
		if err == nil {
			log.Println("ServerManager: Got IP for", me.name+":", targetIp)
			localIp := event.tcpConn.LocalAddr().String()
			me.addPortForwards(localIp, targetIp)
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
