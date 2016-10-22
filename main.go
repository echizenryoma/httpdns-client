package main

import (
	"flag"
	"log"
	"math/rand"
	"net"
	"time"

	"github.com/miekg/dns"
)

var (
	ip   = *flag.String("ip", "127.0.0.1", "DNS bind IP address")
	port = *flag.Int("port", 53, "Listen on port")
)

type dnsMessage struct {
	Question dns.Msg
	Answer   *chan dns.Msg
}

var channel = make(chan dnsMessage, 256)

func handle(dnsQueryMsg []byte, dnsQueryAddress *net.UDPAddr, udpConnection *net.UDPConn) {
	message := new(dns.Msg)
	if err := message.Unpack(dnsQueryMsg); err != nil {
		log.Println(err.Error())
		return
	}

	if len(message.Question) <= 0 {
		return
	}
	question := &message.Question[0]
	answer := getFromCache(question)

	if answer == nil {
		answer := resolveDNSQuestion(message)
		if answer.Answer != nil {
			putCache(question, &answer)
		}
	}
	log.Println(getDNSKey(question))
	buffer, err := answer.Pack()
	if err != nil {
		log.Println(err)
		return
	}
	udpConnection.WriteToUDP(buffer, dnsQueryAddress)
}

func resolveDNSQuestion(question *dns.Msg) dns.Msg {
	answer := make(chan dns.Msg, 1)
	channel <- dnsMessage{*question, answer}
	return <-answer
}

func main() {
	flag.Parse()
	newDNSCache()
	if save {
		initDb()
	}
	rand.Seed(time.Now().UnixNano())

	listenUDP := net.UDPAddr{
		IP:   net.ParseIP(ip),
		Port: port,
	}
	dnsServer, err := net.ListenUDP("udp", &listenUDP)
	if err != nil {
		log.Fatalln(err.Error())
	}

	start()
	log.Printf("Starting server at udp://%s:%d\n", ip, port)
	var buffer []byte
	for {
		buffer = make([]byte, 1500)
		_, address, err := dnsServer.ReadFromUDP(buffer)
		if err == nil {
			go handle(buffer, address, dnsServer)
		}
	}
}
