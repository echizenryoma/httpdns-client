package main

import (
	"fmt"
	"time"

	"github.com/cloudflare/golibs/lrucache"
	"github.com/miekg/dns"
)

var dnsCache *lrucache.LRUCache

func newDNSCache() {
	dnsCache = lrucache.NewLRUCache(uint(128 * 1024 * 1024))
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
	dnsCache.Set(questionKey(question), answer, time.Now().Add(60*time.Second))
}
