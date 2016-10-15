package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/miekg/dns"
)

var (
	workers = *flag.Int("workers", 10, "number of independent workers")
	server  = *flag.String("server", tencentPublicDNS, "Tencent HTTP DNS address")
)

func start() {
	for i := 0; i < workers; i++ {
		go getDNS()
	}
}

func getDNS() {
	for query := range channel {
		answer := dns.Msg{}
		switch query.Question.Question[0].Qtype {
		case dns.TypeA:
			answer, _ = getTencentHTTPDNS(query.Question)
			break
		case dns.TypeAAAA:
			answer.Response = false
			answer.Rcode = dns.RcodeServerFailure
			break
		default:
			answer, _ = getClientDNS(query.Question)
		}
		answer.SetReply(&query.Question)
		*query.Answer <- answer
	}
}

func getTencentHTTPDNS(query dns.Msg) (dns.Msg, bool) {
	var httpClient = &http.Client{Timeout: time.Second * 3, Transport: http.DefaultTransport.(*http.Transport)}

	url := fmt.Sprintf("http://%s/d?dn=%s", server, query.Question[0].Name)
	httpGet, _ := http.NewRequest("GET", url, nil)
	response, err := httpClient.Do(httpGet)

	result := true
	answer := dns.Msg{}
	if err == nil {
		buffer, _ := ioutil.ReadAll(response.Body)
		response.Body.Close()
		if response.StatusCode == 200 {
			ipList := strings.Split(string(buffer), ";")
			if query.Question[0].Qtype == dns.TypeA && len(ipList) > 0 {
				answer.Rcode = dns.RcodeSuccess
				if save {
					insertRecode(query.Question[0].Name, string(buffer))
				}
				for _, ip := range ipList {
					header := dns.RR_Header{
						Name:   query.Question[0].Name,
						Rrtype: query.Question[0].Qtype,
						Class:  dns.ClassINET,
						Ttl:    255,
					}
					dnsRR, _ := dns.NewRR(header.String() + ip)
					answer.Answer = append(answer.Answer, dnsRR)
				}
			}
		} else {
			answer.Rcode = dns.RcodeServerFailure
			result = false
		}
	} else {
		result = false
		answer.Rcode = dns.RcodeServerFailure
		log.Println(err.Error())
	}
	return answer, result
}

func getClientDNS(query dns.Msg) (dns.Msg, bool) {
	dnsClient := new(dns.Client)
	message := new(dns.Msg)
	message.SetQuestion(query.Question[0].Name, query.Question[0].Qtype)
	answer, _, err := dnsClient.Exchange(message, server+":53")
	if err != nil {
		log.Println(err.Error())
		return dns.Msg{}, false
	}
	if answer == nil {
		return dns.Msg{}, false
	}
	if answer.Rcode != dns.RcodeSuccess {
		log.Println(err.Error())
		return dns.Msg{}, false
	}
	return *answer, true
}
