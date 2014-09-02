package main

import (
	"flag"
	"fmt"
	"github.com/soniah/gosnmp"
	"log"
	"net"
	"sync"
)

var debug = false // global debug mode flag

func main() {
	fast := false
	fastUsage := "Query SNMP OID against every IPs in a subnet"
	flag.BoolVar(&fast, "fast", false, fastUsage)
	flag.BoolVar(&fast, "f", false, fastUsage+" (shorthand)")

	community := "public"
	communityUsage := "SNMP Community"
	flag.StringVar(&community, "community", "public", communityUsage)
	flag.StringVar(&community, "c", "public", communityUsage)

	flag.BoolVar(&debug, "d", false, "enable debug mode")

	flag.Parse()

	gs := *gosnmp.Default
	gs.Community = community

	args := flag.Args()
	if len(args) != 2 {
		log.Fatal("Usage: snmpsweep OID CIDR")
	}
	oid := args[0]
	block := args[1]

	ip, ipnet, err := net.ParseCIDR(block)
	if err != nil {
		log.Fatal(err)
	}

	var wg sync.WaitGroup

	for ip := ip.Mask(ipnet.Mask); ipnet.Contains(ip); incIP(ip) {
		myip := ip.String()
		if fast {
			wg.Add(1)
			go func(myip string) {
				defer wg.Done()
				getCommunity(myip, gs, oid)
			}(ip.String())
		} else {
			getCommunity(myip, gs, oid)
		}
	}
	wg.Wait()
}

func getCommunity(ip string, s gosnmp.GoSNMP, oid string) {
	if debug {
		fmt.Printf("Polling %s\n", ip)
	}

	s.Target = ip
	err := s.Connect()
	if err != nil {
		return
	}
	defer s.Conn.Close()

	oids := []string{oid}
	res, err := s.Get(oids)
	if err != nil {
		return
	}
	for _, variable := range res.Variables {
		if variable.Value != nil {
			fmt.Printf("%s %s\n", ip, variable.Value)
		}
	}
}

func incIP(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}
