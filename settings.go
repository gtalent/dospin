/*
   Copyright 2016-2017 gtalent2@gmail.com

   This Source Code Form is subject to the terms of the Mozilla Public
   License, v. 2.0. If a copy of the MPL was not distributed with this
   file, You can obtain one at http://mozilla.org/MPL/2.0/.
*/
package main

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

type Settings struct {
	ApiToken string            `yaml:"api_token"`
	Servers  map[string]Server `yaml:"servers"`
}

type Server struct {
	Ports              []int    `yaml:"ports"`
	ActivityTimeout    string   `yaml:"activity_timeout"`
	UsePublicIP        bool     `yaml:"use_public_ip"`
	InitialSize        string   `yaml:"initial_size"`
	Size               string   `yaml:"size"`
	Region             string   `yaml:"region"`
	UsePersistentImage bool     `yaml:"use_persistent_image"`
	ImageSlug          string   `yaml:"image_slug"`
	UserData           string   `yaml:"user_data"`
	SshKeys            []string `yaml:"ssh_keys"`
	Volumes            []string `yaml:"volumes"`
}

func loadSettings(path string) (Settings, error) {
	var s Settings
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return s, err
	}

	err = yaml.Unmarshal(data, &s)
	if err != nil {
		return s, err
	}

	return s, err
}
