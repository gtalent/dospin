/*
   Copyright 2016-2017 gtalent2@gmail.com

   This Source Code Form is subject to the terms of the Mozilla Public
   License, v. 2.0. If a copy of the MPL was not distributed with this
   file, You can obtain one at http://mozilla.org/MPL/2.0/.
*/
package main

import (
	"log"
	"os/exec"
)

func addPortForward(ruleName, localIp, remoteIp, port string) {
	log.Println("Setting up port", port, "to", remoteIp)
	cmdOut, err := exec.Command("iptables", "-t", "nat", "-A", "PREROUTING", "-p", "tcp", "--dport", port, "-j", "DNAT", "--to-destination", remoteIp+":"+port).Output()
	if err != nil {
		log.Println("iptables error:", err)
	}
	log.Println(cmdOut)

	cmdOut, err = exec.Command("iptables", "-t", "nat", "-A", "POSTROUTING", "-p", "tcp", "-d", localIp, "--dport", port, "-j", "SNAT", "--to-source", remoteIp).Output()
	if err != nil {
		log.Println("iptables error:", err)
	}
	log.Println("iptables", "-t", "nat", "-A", "POSTROUTING", "-p", "tcp", "-d", localIp, "--dport", port, "-j", "SNAT", "--to-source", remoteIp)
}

func rmPortForward(ruleName string) {
	log.Print("Port forwarding not currently implemented for Linux/iptables")
}
