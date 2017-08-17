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
	connActive = iota
	connDisconnected
)

type connStatus struct {
	Status int
	Err    error
}

func portForward(wanConn *net.TCPConn, lanIP, port string, cs chan connStatus) {
	done := make(chan error)
	log.Print("Proxy: Connecting to ", lanIP+":"+port)
	lanConn, err := net.Dial("tcp", lanIP+":"+port)
	if err != nil {
		log.Print("Proxy: LAN dial error: ", err)
		return
	}

	go forwardConn(wanConn, lanConn, done)
	go forwardConn(lanConn, wanConn, done)

	ticker := time.NewTicker(1 * time.Minute)
	for i := 0; i < 2; {
		select {
		case err = <-done:
			if err != nil {
				log.Print("Proxy: ", err)
			}
			i++
		case <-ticker.C:
			cs <- connStatus{Status: connActive}
		}
	}
	log.Print("Proxy: ending connection: ", wanConn.LocalAddr().String())
	ticker.Stop()

	wanConn.Close()
	lanConn.Close()

	cs <- connStatus{Status: connDisconnected, Err: err}
}

func forwardConn(writer, reader net.Conn, done chan error) {
	_, err := io.Copy(writer, reader)
	done <- err
}
