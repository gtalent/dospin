/*
   Copyright 2016-2017 gtalent2@gmail.com

   This Source Code Form is subject to the terms of the Mozilla Public
   License, v. 2.0. If a copy of the MPL was not distributed with this
   file, You can obtain one at http://mozilla.org/MPL/2.0/.
*/
package main

// Firewall forwarding rule managed by dospin.
type ForwardingRule struct {
	destHost string // Destination IP
	destPort int
	srcPort  int
}

func LoadForwardingRules() []ForwardingRule {
	return nil
}
