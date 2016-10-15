package main

import "github.com/miekg/dns"

func printDNS(answer dns.Msg) string {
	log := "\n"
	for _, rr := range answer.Answer {
		log += rr.String() + "\n"
	}
	log = log[0 : len(log)-1]
	return log
}
