package main

import (
	"fmt"
	"sync"
	"time"

	"github.com/miekg/dns"
	cache "github.com/patrickmn/go-cache"
)

var DNSCache *cache.Cache
var DNSCacheInitOnce sync.Once

func initCache() {
	DNSCacheInitOnce.Do(func() {
		if DNSCache == nil {
			DNSCache = cache.New(15*time.Minute, 60*time.Second)
		}
	})
}

func GetFromCache(question dns.Question) (answer dns.Msg, found bool) {
	key := GetDNSKey(question)
	if len(key) <= 0 {
		found = false
		return
	}
	data, found := DNSCache.Get(key)
	if found {
		answer = data.(dns.Msg)
		return
	}
	return
}

func GetDNSKey(question dns.Question) string {
	return fmt.Sprintf("%s|%d|%d", question.Name, question.Qclass, question.Qtype)
}

func AppendDNSCache(question dns.Question, answer dns.Msg) {
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
		DNSCache.Add(GetDNSKey(question), answer, time.Duration(ttl)*time.Second)
	}
}
