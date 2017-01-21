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

const (
	CMD_SERVE       = "serve"
	CMD_SPINDOWNALL = "spindownall"
)

type cmdOptions struct {
	config      string
	cmd         string
	varStateDir string
}

func parseCmdOptions() cmdOptions {
	var o cmdOptions
	flag.StringVar(&o.cmd, "cmd", CMD_SERVE, "Mode to run command in ("+CMD_SERVE+","+CMD_SPINDOWNALL+")")
	flag.StringVar(&o.config, "config", "/etc/dospin.json", "Path to the dospin config file")
	flag.StringVar(&o.varStateDir, "varstate", "/var/lib/dospin", "Path to the var state directory")
	flag.Parse()
	return o
}

func spindownAll(opts cmdOptions) {
	// load settings file
	settings, err := loadSettings(opts.config)
	if err != nil {
		log.Fatal(err)
	}

	// spin down servers
	for name, _ := range settings.Servers {
		dh := NewDropletHandler(settings)
		dh.Spindown(name)
	}
}

func runServer(opts cmdOptions) {
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

		// assign this ServerManager all appropriate ports
		for _, port := range sv.Ports {
			log.Println("Setting up port", port)
			sm.setupListener(port)
		}
	}

	done := make(chan interface{})
	<-done
}

func main() {
	opts := parseCmdOptions()
	switch opts.cmd {
	case CMD_SPINDOWNALL:
		spindownAll(opts)
	case CMD_SERVE:
		runServer(opts)
	default:
		println("Invalid cmd: " + opts.cmd)
	}
}
