package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/miekg/dns"
)

type DNSPod struct {
}

func (this DNSPod) Ip() string {
	return "119.29.29.29"
}

func (this DNSPod) Type() string {
	return "DNSPod"
}

func (this DNSPod) Answer(question dns.Question) (answer dns.Msg, err error) {
	switch question.Qtype {
	case dns.TypeA:
		break
	case dns.TypeAAAA:
		answer.MsgHdr.Rcode = dns.RcodeNameError
		return
	default:
		answer.MsgHdr.Rcode = dns.RcodeNotImplemented
		err = errors.New("type is not implemented")
		return
	}

	if question.Qtype != dns.TypeA {
		answer.MsgHdr.Rcode = dns.RcodeNotImplemented
		err = errors.New("type is not implemented")
		return
	}

	var httpClient = &http.Client{Timeout: time.Second * 3, Transport: http.DefaultTransport.(*http.Transport)}

	url := fmt.Sprintf("http://%s/d?dn=%s&ttl=1", this.Ip(), question.Name)
	log.Println(url)
	httpGet, _ := http.NewRequest("GET", url, nil)
	response, err := httpClient.Do(httpGet)
	if err != nil {
		answer.MsgHdr.Rcode = dns.RcodeServerFailure
		return
	}
	defer response.Body.Close()

	buffer, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return
	}

	if response.StatusCode != 200 {
		answer.MsgHdr.Rcode = dns.RcodeNotZone
		err = errors.New(response.Status)
		return
	}

	if len(buffer) <= 0 {
		answer.MsgHdr.Rcode = dns.RcodeNameError
		return
	}

	ttl := uint32(10)
	splits := strings.Split(string(buffer), ",")
	if len(splits) > 0 {
		var parseTTL uint64
		parseTTL, err = strconv.ParseUint(splits[1], 10, 32)
		if err != nil {
			return
		}
		ttl = uint32(parseTTL)
	}

	ipList := strings.Split(splits[0], ";")
	if question.Qtype == dns.TypeA && len(ipList) > 0 {
		answer.MsgHdr.Rcode = dns.RcodeSuccess
		for _, ip := range ipList {
			header := dns.RR_Header{
				Name:   question.Name,
				Rrtype: question.Qtype,
				Class:  dns.ClassINET,
				Ttl:    ttl,
			}
			var dnsRR dns.RR
			dnsRR, err = dns.NewRR(header.String() + ip)
			if err != nil {
				return
			}
			// log.Printf("%s|%s\n", question.Name, ip)
			answer.Answer = append(answer.Answer, dnsRR)
		}
		// log.Println(answer)
	}
	// log.Printf("%s|%s\n", question.Name, string(buffer))
	return
}
