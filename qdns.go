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
	server  = *flag.String("server", "119.29.29.29", "Tencent HTTP DNS address")
)

func start() {
	for i := 0; i < workers; i++ {
		go dnsResolve(server)
	}
}

func dnsResolve(server string) {
	var httpTr = http.DefaultTransport.(*http.Transport)
	var httpClient = &http.Client{Timeout: time.Second * 10, Transport: httpTr}

	for query := range queryChannel {
		resolve := resolveData{
			Status: -1,
		}

		url := fmt.Sprintf("http://%s/d?dn=%s", server, query.Name)
		httpGet, _ := http.NewRequest("GET", url, nil)

		response, err := httpClient.Do(httpGet)
		if err == nil {
			buffer, _ := ioutil.ReadAll(response.Body)
			response.Body.Close()

			if response.StatusCode == 200 {
				resolve.Status = 0
				answers := strings.Split(string(buffer), ";")
				if query.Type == dns.TypeA {
					if save {
						insertRecode(query.Name, string(buffer))
					}
					for _, answer := range answers {
						resolve.Answer = append(resolve.Answer, resourceRecord{
							Name: query.Name,
							Type: 1,
							TTL:  255,
							Data: answer,
						})
					}
					resolve.Now = time.Now()
					buffer, _ := json.Marshal(resolve)
					log.Println(string(buffer))
				}
			}
		} else {
			log.Println(err.Error())
		}
		*query.resolveChannel <- resolve
	}
}
