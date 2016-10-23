package qdns

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	co "github.com/magicdawn/go-co"
	"github.com/miekg/dns"
)

const (
	tencentPublicDNS = "119.29.29.29"
)

var (
	httpDNS     = *flag.String("httpdns", tencentPublicDNS, "Tencent HTTP DNS address")
	upstreamDNS = *flag.String("dns", tencentPublicDNS, "Upstream DNS Server Address")
)

var dnsServer []string

func resolveAsync(query *dns.Msg) *co.Task {
	return co.Async(func() interface{} {
		if query == nil {
			return nil
		}
		answer := new(dns.Msg)
		question := &query.Question[0]
		switch question.Qtype {
		case dns.TypeA:
			answer = getTencentHTTPDNS(question)
			break
		default:
			answer = getClientDNS(question)
		}
		answer.SetReply(query)
		// buffer, _ := json.Marshal(answer)
		// log.Println(string(buffer))
		return answer
	})
}

func getTencentHTTPDNS(question *dns.Question) *dns.Msg {
	if question == nil {
		return nil
	}

	var httpClient = &http.Client{Timeout: time.Second * 3, Transport: http.DefaultTransport.(*http.Transport)}

	url := fmt.Sprintf("http://%s/d?dn=%s", httpDNS, question.Name)
	httpGet, _ := http.NewRequest("GET", url, nil)
	response, err := httpClient.Do(httpGet)

	answer := new(dns.Msg)
	buffer, _ := ioutil.ReadAll(response.Body)
	response.Body.Close()
	if err != nil {
		answer.Rcode = dns.RcodeServerFailure
		log.Println(err.Error())
		return answer
	}

	if response.StatusCode != 200 || len(buffer) <= 0 {
		answer.Rcode = dns.RcodeNotZone
		return answer
	}

	ipList := strings.Split(string(buffer), ";")
	if question.Qtype == dns.TypeA && len(ipList) > 0 {
		answer.Rcode = dns.RcodeSuccess
		if save {
			co.Await(insertRecodeAsync(question.Name, string(buffer)))
		}
		for _, ip := range ipList {
			header := dns.RR_Header{
				Name:   question.Name,
				Rrtype: question.Qtype,
				Class:  dns.ClassINET,
				Ttl:    10,
			}
			dnsRR, _ := dns.NewRR(header.String() + ip)
			answer.Answer = append(answer.Answer, dnsRR)
		}
	}
	log.Printf("%s|%s\n", question.Name, string(buffer))
	return answer
}

func getClientDNS(question *dns.Question) *dns.Msg {
	if question == nil {
		return nil
	}

	message := new(dns.Msg)
	message.SetQuestion(question.Name, question.Qtype)
	answer := new(dns.Msg)
	if question.Qtype == dns.TypePTR && question.Name == "1.0.0.127.in-addr.arpa." {
		localhost := "tencent.http.dns."
		answer.Question = append(answer.Question, *question)
		header := dns.RR_Header{
			Name:     question.Name,
			Rrtype:   question.Qtype,
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
