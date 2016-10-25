package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net"
	"os"
)

type config struct {
	ListenIP   string   `json:"ip"`
	ListenPort int      `json:"port"`
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

func initConfig(path string) {
	_, err := os.Open(path)
	if err != nil {
		return
	}
	config, err := readConfig(path)
	if err != nil {
		log.Println(err.Error())
		return
	}
	if net.ParseIP(config.ListenIP) != nil {
		ip = config.ListenIP
	}
	if config.ListenPort >= 1 && config.ListenPort <= 65535 {
		port = config.ListenPort
	}
	if len(config.DNS) > 0 {
		dnsServers = config.DNS
	}
	httpDNS = config.HTTPDNS
}
