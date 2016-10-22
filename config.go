package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
)

const (
	tencentPublicDNS = "119.29.29.29"
)

type config struct {
	ListenIP      string   `json:"ip"`
	ListenPort    string   `json:"port"`
	HTTPDNSServer string   `json:"http_dns"`
	DNSServer     []string `json:"dns"`
	Workers       uint16   `json:"workers"`
}

func getConfig(path string) (*config, error) {
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
