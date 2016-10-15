package main

import (
	"encoding/json"
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
		go getTencentHTTPDNS(server)
	}
}

func getTencentHTTPDNS(server string) {
	var httpClient = &http.Client{Timeout: time.Second * 10, Transport: http.DefaultTransport.(*http.Transport)}

	for query := range channel {
		answer := dns.Msg{
			Compress: true,
		}
		answer.Truncated = false
		answer.RecursionDesired = true
		answer.RecursionAvailable = true
		answer.AuthenticatedData = false
		answer.CheckingDisabled = false

		answer.SetReply(&query.Question)

		url := fmt.Sprintf("http://%s/d?dn=%s", server, query.Question.Question[0].Name)
		httpGet, _ := http.NewRequest("GET", url, nil)
		response, err := httpClient.Do(httpGet)
		if err == nil {
			buffer, _ := ioutil.ReadAll(response.Body)
			response.Body.Close()

			if response.StatusCode == 200 {
				ipList := strings.Split(string(buffer), ";")
				if query.Question.Question[0].Qtype == dns.TypeA && len(ipList) > 0 {
					answer.Rcode = dns.RcodeSuccess
					if save {
						insertRecode(query.Question.Question[0].Name, string(buffer))
					}
					for _, ip := range ipList {
						header := dns.RR_Header{
							Name:   query.Question.Question[0].Name,
							Rrtype: query.Question.Question[0].Qtype,
							Class:  dns.ClassINET,
							Ttl:    255,
						}
						dnsRR, _ := dns.NewRR(header.String() + ip)
						answer.Answer = append(answer.Answer, dnsRR)
					}
					buffer, _ := json.Marshal(answer)
					log.Println(string(buffer))
				}
			} else {
				answer.Rcode = dns.RcodeServerFailure
			}
		} else {
			answer.Rcode = dns.RcodeServerFailure
			log.Println(err.Error())
		}
		*query.Answer <- answer
	}
}
