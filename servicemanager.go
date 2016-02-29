/*
   Copyright 2016 gtalent2@gmail.com

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

func setupService(serverManager *ServerManager, port int) {
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
					serverManager.Spinup()

					// close existing connection, not doing anything with it
					conn.Close()
				}
			}
		}
	}()
}
