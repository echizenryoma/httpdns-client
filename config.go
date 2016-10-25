package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
)

type config struct {
	ListenIP   string   `json:"ip"`
	ListenPort string   `json:"port"`
	HTTPDNS    string   `json:"http_dns"`
	DNS        []string `json:"dns"`
	Hosts      string   `json:"hosts"`
}

func readConfig(path string) (*config, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	jsonBytes, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}
	var config config
	if err := json.Unmarshal(jsonBytes, &config); err != nil {
		return nil, err
	}
	return &config, nil
}
