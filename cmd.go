package main

import (
	"flag"
	"log"
	"net"

	"github.com/miekg/dns"
)

var (
	confPath = *flag.String("conf", "config.json", "Configure file path")
)

func init() {
	flag.Parse()

	initCache()
	err := initConfig(confPath)
	if err != nil {
		log.Fatalln(err.Error())
	}
}

func main() {
	udpAddr, err := net.ResolveUDPAddr("udp", config.ListenAddr)
	if err != nil {
		log.Fatalln(err.Error())
	}

	server, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		log.Fatalln(err.Error())
	}
	defer server.Close()
	log.Println("Starting server at dns://" + udpAddr.String())

	var buffer []byte
	for {
		buffer = make([]byte, 512)
		_, client, err := server.ReadFromUDP(buffer)
		// log.Println(buffer)
		if err == nil {
			go handle(buffer, client, server)
		}
	}
}

func handle(dnsQueryMsg []byte, client *net.UDPAddr, server *net.UDPConn) {
	var query dns.Msg
	err := query.Unpack(dnsQueryMsg)
	if err != nil {
		log.Println(err.Error())
		return
	}

	// log.Println(&query)

	if len(query.Question) <= 0 {
		return
	}

	cache, err := DNSApi{}.Factory("Cache")
	if err != nil {
		log.Println(err.Error())
		return
	}

	httpdns, err := DNSApi{}.Factory(config.HTTPDNS)
	if err != nil {
		log.Println(err.Error())
		return
	}

	native, err := DNSApi{}.Factory("Native")
	if err != nil {
		log.Println(err.Error())
		return
	}

	for _, question := range query.Question {
		log.Println(GetDNSKey(question))

		var answer dns.Msg

		answer, err = cache.Answer(question)
		if err != nil {
			switch question.Qtype {
			case dns.TypePTR:
				if question.Name == "1.0.0.127.in-addr.arpa." {
					answer, err = LocalArpa(question)
					if err != nil {
						log.Println(err.Error())
						return
					}
				}
				break
			default:
				answer, err = httpdns.Answer(question)
				if err != nil {
					answer, err = native.Answer(question)
					if err != nil {
						log.Println(err.Error())
						return
					}
				}
				break
			}
		}
		AppendDNSCache(question, answer)
		answer.SetReply(&query)
		// log.Println(&answer)
		WriteToUDP(answer, client, server)
	}
}

func WriteToUDP(answer dns.Msg, client *net.UDPAddr, server *net.UDPConn) {
	buffer, err := answer.Pack()
	if err != nil {
		log.Println(err.Error())
		return
	}
	server.WriteToUDP(buffer, client)
}

func LocalArpa(question dns.Question) (answer dns.Msg, err error) {
	// answer.Question = append(answer.Question, question)
	localhost := "http.dns.client."
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
	return
}
