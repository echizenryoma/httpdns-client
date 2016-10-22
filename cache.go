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

func getFromCache(question *dns.Question) *dns.Msg {
	if question.Qtype != dns.TypeA {
		return nil
	}
	data, found := dnsCache.Get(getDNSKey(question))
	answer := data.(dns.Msg)
	if found {
		return &answer
	}
	return nil
}

func getDNSKey(question *dns.Question) string {
	return fmt.Sprintf("%s|%d|%d", question.Name, question.Qclass, question.Qtype)
}

func putCache(question *dns.Question, answer *dns.Msg) {
	// buffer, _ := json.Marshal(answer)
	// log.Println(string(buffer))
	if answer.Rcode == dns.RcodeSuccess {
		ttl := uint32(10)
		for _, header := range answer.Answer {
			rr := *(header.Header())
			if ttl < rr.Ttl {
				ttl = rr.Ttl
			}
		}

		for _, header := range answer.Ns {
			rr := *(header.Header())
			if ttl < rr.Ttl {
				ttl = rr.Ttl
			}
		}

		for _, header := range answer.Extra {
			rr := *(header.Header())
			if ttl < rr.Ttl {
				ttl = rr.Ttl
			}
		}
		dnsCache.Set(getDNSKey(question), *answer, time.Duration(ttl)*time.Second)
	}
}
