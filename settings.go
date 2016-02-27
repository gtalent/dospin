/*
   Copyright 2016 gtalent2@gmail.com

   This Source Code Form is subject to the terms of the Mozilla Public
   License, v. 2.0. If a copy of the MPL was not distributed with this
   file, You can obtain one at http://mozilla.org/MPL/2.0/.
*/
package main

import (
	"encoding/json"
	"io/ioutil"
)

type Settings struct {
	ApiToken        string
	VirtualDroplets map[string]VirtualDroplet
}

type VirtualDroplet struct {
	Size   string
	Region string
}

func loadSettings(path string) (Settings, error) {
	var s Settings
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return s, err
	}

	err = json.Unmarshal(data, &s)
	if err != nil {
		return s, err
	}

	return s, err
}
