package main

import (
	"errors"

	"github.com/miekg/dns"
)

type CacheDNS struct {
}

func (this CacheDNS) Type() string {
	return "Cache"
}

func (this CacheDNS) Answer(question dns.Question) (answer dns.Msg, err error) {
	answer, found := GetFromCache(question)
	if !found {
		err = errors.New("cache not found")
		return
	}
	return
}
