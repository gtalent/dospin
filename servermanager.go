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
	"time"
)

const (
	servermanagerSpinup = iota
	servermanagerSpindown
	serverManagerStop
)

type serverManagerEvent struct {
	eventType int
	tcpConn   *net.TCPConn
}

/*
ServerHandler is an interface for spinning up and spinning down servers.
*/
type ServerHandler interface {
	// Takes snapshot name, and returns the IP to connect to.
	Spinup(name string) (string, error)
	Spindown(name string) error
}

type serverManager struct {
	name              string
	ports             []int
	in                chan serverManagerEvent
	done              chan int
	connStatus        chan connStatus
	lastKeepAliveTime time.Time
	server            ServerHandler
	activityTimeout   time.Duration
}

func newServerManager(name string, server ServerHandler, settings settings) *serverManager {
	sm := new(serverManager)

	sm.name = name
	sm.ports = settings.Servers[name].Ports
	sm.in = make(chan serverManagerEvent)
	sm.done = make(chan int)
	sm.connStatus = make(chan connStatus)
	sm.server = server
	sm.lastKeepAliveTime = time.Now()

	activityTimeout, err := time.ParseDuration(settings.Servers[sm.name].ActivityTimeout)
	if err != nil { // invalid timeout, default to 5 minutes
		activityTimeout = time.Duration(5 * time.Minute)
	}
	sm.activityTimeout = activityTimeout
	log.Println("serverManager: ", name, " has activity timeout of ", sm.activityTimeout.String())

	return sm
}

/*
 Serves channel requests.
*/
func (me *serverManager) Serve() {
	// TODO: see if server is currently up, and setup port forwarding if so

	ticker := time.NewTicker(1 * time.Minute)

	// event loop
	for running := true; running; {
		select {
		case status := <-me.connStatus:
			if status.Status == connActive {
				me.lastKeepAliveTime = time.Now()
			}
		case action := <-me.in:
			running = me.serveAction(action)
		case <-ticker.C:
			if time.Since(me.lastKeepAliveTime) > me.activityTimeout {
				running = me.serveAction(serverManagerEvent{eventType: servermanagerSpindown})
			}
		}
	}

	ticker.Stop()

	// notify done
	me.done <- 0
}

/*
 Sends the serve loop a spinup message.
*/
func (me *serverManager) Spinup(c *net.TCPConn) {
	me.in <- serverManagerEvent{eventType: servermanagerSpinup, tcpConn: c}
}

/*
 Sends the serve loop a spindown message.
*/
func (me *serverManager) Spindown() {
	me.in <- serverManagerEvent{eventType: servermanagerSpindown}
}

/*
 Sends the serve loop a quit message.
*/
func (me *serverManager) Stop() {
	me.in <- serverManagerEvent{eventType: serverManagerStop}
}

func (me *serverManager) Done() {
	<-me.done
}

func (me *serverManager) setupListener(port int) {
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
				} else { // connection accepted
					// spinup machine
					me.Spinup(conn)
				}
			}
		}
	}()
}

func (me *serverManager) serveAction(event serverManagerEvent) bool {
	running := true
	switch event.eventType {
	case servermanagerSpinup:
		targetIP, err := me.server.Spinup(me.name)
		me.lastKeepAliveTime = time.Now()
		if err == nil {
			log.Println("serverManager: Got IP for", me.name+":", targetIP)
			wanAddr := event.tcpConn.LocalAddr().String()
			_, port, _ := net.SplitHostPort(wanAddr)
			go portForward(event.tcpConn, targetIP, port, me.connStatus)
		} else {
			log.Println("serverManager: Could not spin up "+me.name+":", err)
		}
	case servermanagerSpindown:
		err := me.server.Spindown(me.name)
		if err != nil {
			log.Println("serverManager: Could not spin down "+me.name+":", err)
		}
	case serverManagerStop:
		running = false
	}
	return running
}
