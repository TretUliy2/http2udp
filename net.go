package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"syscall"
)

const (
	bufsize  int = 131600
	pktsisze int = 188
	datalen  int = pktsisze * 7
)

func main() {
	source := flag.String("s", "http://10.71.0.190:12999/udp/239.23.0.5:1234",
		"Source http url")
	udp := flag.String("d", "239.23.15.102", "destination udp address")
	ttl := flag.Int("t", 64, "ttl value int default 64")
	tos := flag.Int("p", 136, "tos value in dec, int default 136")
	help := flag.Bool("h", false, "help")

	flag.Parse()
	if *help {
		usage()
	}

	log.Printf("%s -> %s\n", *source, *udp)
	readhttp(*source, *udp, *ttl, *tos)
}

func usage() {
	fmt.Printf("%s -s <source http address> -d <destination multicast address>"+
		"-h help [-t ttl] [-p tos in dec default 136]\n", os.Args[0])
	os.Exit(1)
}

func readhttp(httpaddr string, udpaddr string, ttl int, tos int) {
	resp, err := http.Get(httpaddr)
	maddr := net.UDPAddr{IP: net.ParseIP(udpaddr), Port: 1234}
	udpConn, udpConErr := net.DialUDP("udp", nil, &maddr)

	if err != nil {
		log.Fatalf("Error while connecting to source %v", err)
	}
	defer resp.Body.Close()
	if udpConErr != nil {
		log.Fatalf("udp connection error %v", udpConErr)
	}
	// Setting TOS and TTL for connection
	// ipv4conn := ipv4.NewPacketConn(udpConn)
	// ipv4conn.SetTOS(0xc0)
	// ipv4conn.SetMulticastTTL(64)

	// Buffer without it error in
	udpConn.SetWriteBuffer(datalen)
	defer udpConn.Close()
	log.Printf("Dialed successful %s\n", httpaddr)
	udpsockopt(udpConn, ttl, tos)
	data := make([]byte, bufsize)

	for {
		count, err := io.ReadFull(resp.Body, data)
		if err != nil {
			log.Printf("Error while reading data %v", err)
			break
		}
		for i := 0; i+datalen <= count; i += datalen {
			udpConn.Write(data[i : i+datalen])
		}
	}
}

func udpsockopt(udpConn *net.UDPConn, ttl int, tos int) {
	file, err := udpConn.File()
	fd := file.Fd()

	err = syscall.SetsockoptInt(int(fd), syscall.SOL_IP, syscall.IP_MULTICAST_TTL, ttl)
	if err != nil {
		log.Printf("Error setting TTL sockopt %v\n", err)
	}
	err = syscall.SetsockoptInt(int(fd), syscall.SOL_IP, syscall.IP_TOS, tos)
	if err != nil {
		log.Printf("Error setting TOS sockopt %v", err)
	}
}
