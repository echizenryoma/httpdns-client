package main

import (
	"errors"
	"strings"
	"time"

	"github.com/miekg/dns"
)

type NativeDNS struct {
}

func (this NativeDNS) Type() string {
	return "Native"
}

func (this NativeDNS) Answer(question dns.Question) (answer dns.Msg, err error) {
	dnsNum := len(config.DNS)

	if dnsNum <= 0 {
		err = errors.New("dns is empty")
		return
	}
	answerChan := make(chan dns.Msg, dnsNum)
	defer close(answerChan)

	for _, dnsServer := range config.DNS {
		if !strings.Contains(dnsServer, ":") {
			dnsServer += ":53"
		}

		go func(dnsServer string) {
			defer func() {
				if recover() != nil {
					return
				}
			}()

			msg := new(dns.Msg)
			msg.SetQuestion(question.Name, question.Qtype)

			dnsClient := new(dns.Client)
			ans, _, err := dnsClient.Exchange(msg, dnsServer)
			if err != nil {
				return
			}
			answerChan <- *ans
		}(dnsServer)
	}

	timer := time.NewTimer(time.Second * 2)
	var ok bool
	select {
	case answer, ok = <-answerChan:
		if ok {
			return
		}
		break
	case <-timer.C:
		err = errors.New("timeout")
		return
	}

	answer.Rcode = dns.RcodeServerFailure
	return
}
