/*
   Copyright 2016 gtalent2@gmail.com

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

func setupPortForward(ip, port string) {
}

func portUsageCount(ports ...int) int {
	cmd := "/usr/bin/sockstat"
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
