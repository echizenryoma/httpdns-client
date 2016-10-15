package main

import (
	"time"

	"github.com/miekg/dns"
	cache "github.com/patrickmn/go-cache"
)

var dnsCache *cache.Cache

func newDNSCache() {
	dnsCache = cache.New(24*time.Hour, 60*time.Second)
}

func getFromCache(question dns.Question) (dns.Msg, bool) {
	if question.Qtype != dns.TypeA {
		return dns.Msg{}, false
	}
	data, found := dnsCache.Get(question.Name)
	if found {
		return data.(dns.Msg), true
	}
	return dns.Msg{}, false
}
