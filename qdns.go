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
	upstreamDNS = *flag.String("dns", tencentPublicDNS, "Upstream DNS Server Address")
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
		answer := new(dns.Msg)
		question := query.Question.Question[0]
		switch question.Qtype {
		case dns.TypeA:
			answer = getTencentHTTPDNS(query.Question)
			break
		default:
			answer = getClientDNS(query.Question)
		}
		answer.SetReply(query.Question)
		*query.Answer <- answer
	}
}

func getTencentHTTPDNS(query *dns.Msg) *dns.Msg {
	var httpClient = &http.Client{Timeout: time.Second * 3, Transport: http.DefaultTransport.(*http.Transport)}

	url := fmt.Sprintf("http://%s/d?dn=%s", httpDNS, query.Question[0].Name)
	httpGet, _ := http.NewRequest("GET", url, nil)
	response, err := httpClient.Do(httpGet)

	answer := new(dns.Msg)
	qustion := query.Question[0]
	if err == nil {
		buffer, _ := ioutil.ReadAll(response.Body)
		response.Body.Close()
		if response.StatusCode == 200 {
			ipList := strings.Split(string(buffer), ";")
			if qustion.Qtype == dns.TypeA && len(ipList) > 0 {
				answer.Rcode = dns.RcodeSuccess
				if save {
					insertRecode(qustion.Name, string(buffer))
				}
				for _, ip := range ipList {
					header := dns.RR_Header{
						Name:   qustion.Name,
						Rrtype: qustion.Qtype,
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
		}
	} else {
		answer.Rcode = dns.RcodeServerFailure
		log.Println(err.Error())
	}
	return answer
}

func getClientDNS(query *dns.Msg) *dns.Msg {
	message := new(dns.Msg)
	message.SetQuestion(query.Question[0].Name, query.Question[0].Qtype)
	answer := new(dns.Msg)
	if query.Question[0].Qtype == dns.TypePTR && query.Question[0].Name == "1.0.0.127.in-addr.arpa." {
		localhost := "tencent.http.dns."
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
		return answer
	}

	for _, server := range dnsServer {
		dnsClient := new(dns.Client)
		answer, _, err := dnsClient.Exchange(message, server+":53")
		if err != nil {
			log.Println(err.Error())
			continue
		}
		if answer.Rcode == dns.RcodeSuccess {
			return answer
		}
	}
	answer.Rcode = dns.RcodeServerFailure
	return answer
}
