package main

import (
	"flag"
	"log"
	"net"
	"os"

	hosts "github.com/janeczku/go-dnsmasq/hostsfile"
)

var (
	hostsFilePath = *flag.String("hosts", "hosts", "Hosts File Path")
)

var hostsFile *hosts.Hostsfile
var useHosts bool

func initHosts(path string) {
	useHosts = false
	if len(path) <= 0 {
		return
	}
	_, err := os.Stat(path)
	if err != nil {
		return
	}
	hostsFile, err = readHostFile(path)
	if err != nil {
		log.Println(err.Error())
		return
	}
	useHosts = true
}

func readHostFile(path string) (*hosts.Hostsfile, error) {
	host, err := hosts.NewHostsfile(path, &hosts.Config{
		Poll:    10,
		Verbose: false,
	})
	if err != nil {
		return nil, err
	}
	return host, nil
}

func getHosts(domain string) (*[]net.IP, error) {
	ips, err := hostsFile.FindHosts(domain)
	if err != nil {
		return nil, err
	}
	return &ips, nil
}
