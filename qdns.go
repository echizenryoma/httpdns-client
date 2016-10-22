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
	workers     = *flag.Int("workers", 10, "number of independent workers")
	httpDNS     = *flag.String("httpdns", tencentPublicDNS, "Tencent HTTP DNS address")
	upstreamDNS = *flag.String("httpdns", tencentPublicDNS, "Upstream DNS Server Address")
)

var dnsServer []string

func start() {
	dnsServer = strings.Split(upstreamDNS, ";")
	for i := 0; i < workers; i++ {
		go getDNS()
	}
}

func getDNS() {
	for query := range channel {
		answer := dns.Msg{}
		switch query.Question.Question[0].Qtype {
		case dns.TypeA:
			answer, _ = getTencentHTTPDNS(&query.Question)
			break
		default:
			answer, _ = getClientDNS(&query.Question)
		}
		*answer.SetReply(&query.Question)
		*query.Answer <- answer
	}
}

func getTencentHTTPDNS(query *dns.Msg) (*dns.Msg, bool) {
	var httpClient = &http.Client{Timeout: time.Second * 3, Transport: http.DefaultTransport.(*http.Transport)}

	url := fmt.Sprintf("http://%s/d?dn=%s", httpDNS, query.Question[0].Name)
	httpGet, _ := http.NewRequest("GET", url, nil)
	response, err := httpClient.Do(httpGet)

	result := true
	answer := new(dns.Msg)
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
						Ttl:    10,
					}
					dnsRR, _ := dns.NewRR(header.String() + ip)
					answer.Answer = append(answer.Answer, dnsRR)
				}
			}
			log.Printf("%s|%s\n", query.Question[0].Name, string(buffer))
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

func getClientDNS(query *dns.Msg) (*dns.Msg, bool) {
	message := new(dns.Msg)
	message.SetQuestion(query.Question[0].Name, query.Question[0].Qtype)
	if query.Question[0].Qtype == dns.TypePTR && query.Question[0].Name == "1.0.0.127.in-addr.arpa." {
		localhost := "tencent.http.dns."
		answer := new(dns.Msg)
		answer.Question = append(answer.Question, query.Question[0])
		header := dns.RR_Header{
			Name:     query.Question[0].Name,
			Rrtype:   query.Question[0].Qtype,
			Class:    1,
			Ttl:      10080,
			Rdlength: uint16(len(localhost)),
		}
		rr := new(dns.PTR)
		rr.Hdr = header
		rr.Ptr = localhost
		answer.Answer = append(answer.Answer, rr)
		// buffer, _ := json.Marshal(answer)
		// log.Println(string(buffer))
		return answer, true
	}

	for _, server := range dnsServer {
		dnsClient := new(dns.Client)
		answer, _, err := dnsClient.Exchange(message, server+":53")
		if err != nil {
			log.Println(err.Error())
			continue
		}
		if answer.Rcode == dns.RcodeSuccess {
			return answer, true
		}
	}
	return nil, false
}
