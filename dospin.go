/*
   Copyright 2016-2017 gtalent2@gmail.com

   This Source Code Form is subject to the terms of the Mozilla Public
   License, v. 2.0. If a copy of the MPL was not distributed with this
   file, You can obtain one at http://mozilla.org/MPL/2.0/.
*/
package main

import (
	"flag"
	"log"
	"os"
)

const (
	CMD_SERVE       = "serve"
	CMD_SPINDOWNALL = "spindownall"
)

type cmdOptions struct {
	config  string
	logFile string
	cmd     string
}

func parseCmdOptions() cmdOptions {
	var o cmdOptions
	flag.StringVar(&o.cmd, "cmd", CMD_SERVE, "Mode to run command in ("+CMD_SERVE+","+CMD_SPINDOWNALL+")")
	flag.StringVar(&o.config, "config", "dospin.yaml", "Path to the dospin config file")
	flag.StringVar(&o.logFile, "logFile", "stdout", "Path to the dospin log file")
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

func runServer(opts cmdOptions) int {
	if opts.logFile != "stdout" {
		logFile, err := os.OpenFile(opts.logFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0664)
		if err == nil {
			defer logFile.Close()
			log.SetOutput(logFile)
		} else {
			log.Print("Could not open log file: ", err)

		}
	}
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

	done := make(chan int)
	return <-done
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
