package main

import (
	"encoding/json"
	"io/ioutil"
)

type Settings struct {
	Token           string
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
