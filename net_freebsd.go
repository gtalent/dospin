/*
   Copyright 2016-2017 gtalent2@gmail.com

   This Source Code Form is subject to the terms of the Mozilla Public
   License, v. 2.0. If a copy of the MPL was not distributed with this
   file, You can obtain one at http://mozilla.org/MPL/2.0/.
*/
package main

import (
	"bytes"
	"log"
	"os/exec"
	"strconv"
)

func addPortForward(ruleName, ip, port string) {
	pfrule := "\"rdr pass on $dospin_ext_if proto { tcp, udp } from any to any port {" + port + "} -> " + ip + "\""

	in, err := exec.Command("pfctl", "-a", "\"dospin_"+ruleName+"\"", "-f", "-").StdinPipe()
	defer in.Close()
	if err != nil {
		log.Println("Port Forwarding:", err)
	}

	_, err = in.Write([]byte(pfrule))
	if err != nil {
		log.Println("Port Forwarding:", err)
	}
}

func rmPortForward(ruleName string) {
	_, err := exec.Command("pfctl", "-a", "\"dospin_"+ruleName+"\"", "-F", "rules").Output()
	if err != nil {
		log.Println("Port Forwarding:", err)
	}
}

func portUsageCount(ports ...int) int {
	cmd := "sockstat"
	args := []string{"-4c"}
	for _, v := range ports {
		args = append(args, "-p")
		args = append(args, strconv.Itoa(v))
	}
	out, err := exec.Command(cmd, args...).Output()
	if err != nil {
		log.Println("Port Usage Check: Could not run \""+cmd+"\":", err)
	}
	return bytes.Count(out, []byte{'\n'}) - 1
}
