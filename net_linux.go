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

func addPortForward(ruleName, gatewayInt, localIp, targetIp, port string) {
	log.Println("Setting up port", port, "to", targetIp)
	cmdOut, err := exec.Command("iptables", "-A", "PREROUTING", "-t", "nat", "-i", "\""+gatewayInt+"\"", "-p", "tcp", "--dport", port, "-j", "DNAT", "--to", targetIp+":"+port).Output()
	log.Println("iptables", "-A", "PREROUTING,", "-t", "nat", "-i", "\""+gatewayInt+"\"", "-p", "tcp", "--dport", port, "-j", "DNAT", "--to", targetIp+":"+port)
	if err != nil {
		log.Println("iptables error:", err)
	}

	cmdOut, err = exec.Command("iptables", "-A", "FORWARD", "-p", "tcp", "-d", targetIp, "--dport", port, "-j", "ACCEPT").Output()
	log.Println("iptables", "-A", "FORWARD", "-p", "tcp", "-d", targetIp, "--dport", port, "-j", "ACCEPT")
	if err != nil {
		log.Println("iptables error:", err)
	}
}

func rmPortForward(ruleName string) {
	log.Print("Port forwarding not currently implemented for Linux/iptables")
}
