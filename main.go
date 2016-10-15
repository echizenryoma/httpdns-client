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

func handle(buf []byte, udpAddress *net.UDPAddr, udpConnection *net.UDPConn) {
	message := new(dns.Msg)
	if err := message.Unpack(buf); err != nil {
		log.Println(err)
		return
	}

	if len(message.Question) <= 0 {
		return
	}

	answer, found := getFromCache(message.Question[0])
	if !found {
		answer = submitQuetion(*message)
		if answer.Answer != nil {
			putCache(message.Question[0].Name, answer)
		}
	}
	buffer, err := answer.Pack()
	if err != nil {
		log.Println(err)
		return
	}
	udpConnection.WriteToUDP(buffer, udpAddress)
}

func submitQuetion(question dns.Msg) dns.Msg {
	answer := make(chan dns.Msg, 1)
	channel <- dnsMessage{question, &answer}
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
		n, addr, err := dnsServer.ReadFromUDP(buffer)
		if err == nil {
			go handle(buffer[0:n], addr, dnsServer)
		}
	}
}
