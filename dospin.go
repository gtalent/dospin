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

func main() {
	opts := parseCmdOptions()
	log.Println("Loading config:", opts.config)
	settings, err := loadSettings(opts.config)
	if err != nil {
		log.Fatal(err)
	}

	dh := NewDropletHandler(settings)

	ip, err := dh.SpinupMachine("minecraft")
	if err != nil {
		log.Println("Error:", err)
		return
	}
	log.Println("IP: " + ip)

	if err := dh.SpindownMachine("minecraft"); err != nil {
		log.Println("Error:", err)
		return
	}
}
