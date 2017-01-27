/*
   Copyright 2016-2017 gtalent2@gmail.com

   This Source Code Form is subject to the terms of the Mozilla Public
   License, v. 2.0. If a copy of the MPL was not distributed with this
   file, You can obtain one at http://mozilla.org/MPL/2.0/.
*/
package main

import (
	"io"
	"log"
	"net"
	"time"
)

const (
	CONN_ACTIVE = iota
	CONN_DISCONNECTED
)

type ConnStatus struct {
	Status int
	Err    error
}

func portForward(wanConn *net.TCPConn, lanIp, port string, connStatus chan ConnStatus) {
	done := make(chan error)
	log.Print("Proxy: Connecting to ", lanIp+":"+port)
	lanConn, err := net.Dial("tcp", lanIp+":"+port)
	if err != nil {
		log.Print("Proxy: LAN dial error:", err)
		return
	}

	go forwardConn(wanConn, lanConn, done)
	go forwardConn(lanConn, wanConn, done)

	ticker := time.NewTicker(time.Minute * 1)
	for i := 0; i < 2; i++ {
		select {
		case err = <-done:
			if err != nil {
				log.Print("Proxy:", err)
			}
		case <-ticker.C:
			connStatus <- ConnStatus{Status: CONN_ACTIVE}
		}
	}
	ticker.Stop()

	wanConn.Close()
	lanConn.Close()

	connStatus <- ConnStatus{Status: CONN_DISCONNECTED, Err: err}
}

func forwardConn(writer, reader net.Conn, done chan error) {
	_, err := io.Copy(writer, reader)
	done <- err
}
