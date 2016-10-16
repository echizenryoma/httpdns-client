package main

import (
	"fmt"
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
	data, found := dnsCache.Get(questionKey(question))
	if found {
		return data.(dns.Msg), true
	}
	return dns.Msg{}, false
}

func questionKey(question dns.Question) string {
	return fmt.Sprintf("%s|%d|%d", question.Name, question.Qclass, question.Qtype)
}

func putCache(question dns.Question, answer dns.Msg) {
	dnsCache.Set(questionKey(question), answer, 10*time.Second)
}
