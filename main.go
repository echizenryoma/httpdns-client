package main

import (
	"flag"
	"log"
	"math/rand"
	"net"
	"time"

	co "github.com/magicdawn/go-co"
	"github.com/miekg/dns"
)

var (
	ip   = *flag.String("ip", "127.0.0.1", "DNS bind IP address")
	port = *flag.Int("port", 53, "Listen on port")
)

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
	log.Println(getDNSKey(question))
	answer := getFromCache(question)
	if answer == nil {
		result, _ := co.Await(resolveAsync(message))
		answer = result.(*dns.Msg)
		if answer != nil {
			putCache(question, answer)
		}
	}
	buffer, err := answer.Pack()
	if err != nil {
		log.Println(err.Error())
		return
	}
	udpConnection.WriteToUDP(buffer, dnsQueryAddress)
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
