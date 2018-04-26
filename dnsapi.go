package main

import (
	"errors"
	"strings"
)

type DNSApi struct {
}

func (this DNSApi) Factory(name string) (httpdns DNSBase, err error) {
	switch strings.ToLower(name) {
	case strings.ToLower(DNSPod{}.Type()):
		httpdns = DNSPod{}
		return
	case strings.ToLower(NativeDNS{}.Type()):
		httpdns = NativeDNS{}
		return
	case strings.ToLower(CacheDNS{}.Type()):
		httpdns = CacheDNS{}
		return
	default:
		err = errors.New("type of " + name + " is not exit")
		return
	}
}
