/*
   Copyright 2016-2017 gtalent2@gmail.com

   This Source Code Form is subject to the terms of the Mozilla Public
   License, v. 2.0. If a copy of the MPL was not distributed with this
   file, You can obtain one at http://mozilla.org/MPL/2.0/.
*/
package main

import (
	"net"
	"strconv"
	"testing"
)

func TestPortCount(t *testing.T) {
	port := 49214
	// listen on some port that nothing should be using for the sake of the test
	go func() {
		addr, _ := net.ResolveTCPAddr("tcp", "0.0.0.0:"+strconv.Itoa(port))
		net.ListenTCP("tcp", addr)
	}()

	if portUsageCount(49214) != 1 {
		t.Errorf("Port count usage reporting wrong number")
	}
}
