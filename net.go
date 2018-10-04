package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"syscall"
	"time"
)

const (
	maxdata     int = 131600
	pktlen      int = 188
	pkts        int = 7
	datalen     int = pktlen * pkts
	defaultport int = 1234
)

var data []byte

func main() {
	data = make([]byte, maxdata)
	sourceurl := flag.String("s", "http://82.193.96.217:12999/udp/239.23.0.5:1234", "a string")
	destination := flag.String("d", "239.23.15.102", "a string")
	ttl := flag.Int("t", 64, "ttl of destination packets")
	tos := flag.Int("o", 136, "tos of a packet")
	flag.Parse()

	fmt.Printf("%d %d\n", *ttl, *tos)
	readhttp(*sourceurl, *destination, *ttl, *tos)
}

func readhttp(url string, udpaddr string, ttl int, tos int) {
	// Function will read http input and assume it have valid mpegts stream
	// Then it will write to udp multicast destination
	resp, err := http.Get(url)
	fmt.Printf("connecting to %s will send output to %s\n", url, udpaddr)
	for err != nil {
		log.Printf("Error while connecting to source %v\n", err)
		time.Sleep(10)
		resp, err = http.Get(url)
	}
	fmt.Println("Dialed successful")
	defer resp.Body.Close()
	// Udp Connection open
	madr := net.UDPAddr{IP: net.ParseIP(udpaddr), Port: defaultport}
	udpConn, udpConErr := net.DialUDP("udp", nil, &madr)
	if udpConErr != nil {
		log.Fatalf("udp connection error %v", udpConErr)
	}
	defer udpConn.Close()
	// Setting Socket options and buffer size
	setSockOpts(udpConn, ttl, tos)
	udpConn.SetWriteBuffer(datalen)
	for {
		count, err := io.ReadFull(resp.Body, data)
		if err != nil {
			fmt.Printf("Error while reading source %v\n", err)
		}
		if count < datalen {
			log.Printf("received smal amount of data %d except %d\n", count, len(data))
		}

		//udpConn, udpConErr := net.Dial("udp", "239.23.15.102:1234")
		for i := 0; i+datalen <= count; i += datalen {
			cnt, err := udpConn.Write(data[i : i+datalen])
			if err != nil {
				fmt.Printf("Error while writing to socket %v bytes\n", cnt)
			}
		}
	}
}

func setSockOpts(udpConn *net.UDPConn, params ...int) {
	fmt.Printf("Setting up TTL and TOS for outgoing traffic\n")
	var tos = 136
	var ttl = 64
	if len(params) == 2 {
		ttl = params[0]
		tos = params[1]
	}
	fl, err := udpConn.File()
	if err != nil {
		fmt.Printf("Error getting file object %v\n", err)
	}
	fd := fl.Fd()
	sockerr := syscall.SetsockoptInt(int(fd), syscall.SOL_IP, syscall.IP_MULTICAST_TTL, ttl)
	if sockerr != nil {
		fmt.Printf("Error setting TTL sockopt %v\n", sockerr)
	}
	sockerr = syscall.SetsockoptInt(int(fd), syscall.SOL_IP, syscall.IP_TOS, tos)
	if sockerr != nil {
		fmt.Printf("Error setting IP_TOS sockopt %v\n", sockerr)
	}
}
