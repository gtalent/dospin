/*
   Copyright 2016 gtalent2@gmail.com

   This Source Code Form is subject to the terms of the Mozilla Public
   License, v. 2.0. If a copy of the MPL was not distributed with this
   file, You can obtain one at http://mozilla.org/MPL/2.0/.
*/
package main

import (
	"flag"
	"log"
)

type cmdOptions struct {
	config string
}

func parseCmdOptions() cmdOptions {
	var o cmdOptions
	flag.StringVar(&o.config, "config", "dospin.json", "Path to the dospin config file")
	flag.Parse()
	return o
}

func testServerManager(settings Settings) {
	dh := NewDropletHandler(settings)
	sm := NewServerManager("minecraft", dh, settings)
	go sm.Serve()

	sm.Spinup()
	sm.Spindown()

	sm.Stop()
	sm.Done()
}

func main() {
	opts := parseCmdOptions()
	log.Println("Loading config:", opts.config)
	settings, err := loadSettings(opts.config)
	if err != nil {
		log.Fatal(err)
	}

	for name, sv := range settings.Servers {
		dh := NewDropletHandler(settings)
		sm := NewServerManager(name, dh, settings)

		// start the ServerManager
		go sm.Serve()

		// assign this ServerManager to all appropriate ports
		for _, port := range sv.Ports {
			log.Println("Setting up port", port)
			setupService(sm, port)
		}
	}

	done := make(chan interface{})
	<-done
}
