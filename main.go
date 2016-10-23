package main

import (
	"flag"
	"log"
	"net"
	"strings"

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
		if result == nil {
			return
		}
		answer = result.(*dns.Msg)
		if answer == nil {
			return
		}
		putCache(question, answer)
	}
	buffer, err := answer.Pack()
	if err != nil {
		log.Println(err.Error())
		return
	}
	udpConnection.WriteToUDP(buffer, dnsQueryAddress)
}

func init() {
	flag.Parse()
	dnsServer = strings.Split(upstreamDNS, ";")
	newDNSCache()
	if save {
		initDb()
	}
	if len(hostsFilePath) > 0 {
		initHosts(hostsFilePath)
	}
}

func main() {
	server, err := net.ListenUDP("udp", &net.UDPAddr{
		IP:   net.ParseIP(ip),
		Port: port,
	})
	if err != nil {
		log.Fatalln(err.Error())
	}
	log.Printf("Starting server at udp://%s:%d\n", ip, port)
	var buffer []byte
	for {
		buffer = make([]byte, 512)
		_, address, err := server.ReadFromUDP(buffer)
		if err == nil {
			go handle(buffer, address, server)
		}
	}
}
