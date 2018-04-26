package main

import "github.com/miekg/dns"

type DNSBase interface {
	Type() string

	Answer(dns.Question) (dns.Msg, error)
}
