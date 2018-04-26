package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
)

type Config struct {
	ListenAddr string   `json:"listen"`
	HTTPDNS    string   `json:"http_dns"`
	DNS        []string `json:"dns"`
}

var config Config

func readConfig(path string) (err error) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	bytes, err := ioutil.ReadAll(f)
	if err != nil {
		return
	}
	err = json.Unmarshal(bytes, &config)
	if err != nil {
		return
	}
	return
}

func initConfig(path string) (err error) {
	_, err = os.Open(path)
	if err != nil {
		return
	}
	err = readConfig(path)
	if err != nil {
		return
	}
	return
}
