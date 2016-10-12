package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net"
	"time"

	"github.com/miekg/dns"
	"github.com/patrickmn/go-cache"
)

var (
	ip   = *flag.String("ip", "127.0.0.1", "DNS bind IP address")
	port = *flag.Int("port", 53, "listen on port number")
)

type resourceRecord struct {
	Name string
	Type uint16
	TTL  uint32
	Data string
}

type resolveData struct {
	Status int
	Answer []resourceRecord
	Now    time.Time
}

type resolveQuery struct {
	Name           string
	Type           uint16
	resolveChannel *chan resolveData
}

var queryChannel = make(chan resolveQuery, 256)
var resolveCache *cache.Cache

func handle(buf []byte, udpAddress *net.UDPAddr, udpConnection *net.UDPConn) {
	dnsMessage := new(dns.Msg)
	if err := dnsMessage.Unpack(buf); err != nil {
		log.Println(err)
		return
	}

	if len(dnsMessage.Question) < 1 {
		return
	}

	var r resolveData
	cid := fmt.Sprintf("%s/%d", dnsMessage.Question[0].Name, dnsMessage.Question[0].Qtype)

	if data, found := resolveCache.Get(cid); found {
		r = data.(resolveData)
	} else {
		r = resolve(dnsMessage.Question[0].Name, dnsMessage.Question[0].Qtype)
		resolveCache.Set(cid, r, 10*time.Second)
	}

	resolveMessage := new(dns.Msg)
	resolveMessage.SetReply(dnsMessage)
	resolveMessage.Compress = true
	if r.Status >= 0 {
		resolveMessage.Rcode = r.Status
		resolveMessage.Truncated = false
		resolveMessage.RecursionDesired = true
		resolveMessage.RecursionAvailable = true
		resolveMessage.AuthenticatedData = false
		resolveMessage.CheckingDisabled = false

		for _, httpDNSRR := range r.Answer {
			resolveMessage.Answer = append(resolveMessage.Answer, convertDNSRR(httpDNSRR))
		}
	} else {
		resolveMessage.Rcode = 2
	}

	buffer, err := resolveMessage.Pack()
	if err != nil {
		log.Fatalln(err)
		return
	}
	udpConnection.WriteToUDP(buffer, udpAddress)
}

func convertDNSRR(ResourceRecord resourceRecord) dns.RR {
	DNSRRHeader := dns.RR_Header{
		Name:   ResourceRecord.Name,
		Rrtype: ResourceRecord.Type,
		Class:  dns.ClassINET,
		Ttl:    ResourceRecord.TTL,
	}
	DNSRR, _ := dns.NewRR(DNSRRHeader.String() + ResourceRecord.Data)
	return DNSRR
}

func resolve(name string, qtype uint16) resolveData {
	query := make(chan resolveData, 1)
	queryChannel <- resolveQuery{name, qtype, &query}
	return <-query
}

func main() {
	flag.Parse()

	if save {
		initDb()
	}

	rand.Seed(time.Now().UnixNano())
	resolveCache = cache.New(24*time.Hour, 60*time.Second)

	laddr := net.UDPAddr{
		IP:   net.ParseIP(ip),
		Port: port,
	}
	dnsServer, err := net.ListenUDP("udp", &laddr)
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
