package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
)

func listen(conn *net.UDPConn) {
	fmt.Println("On listener")
	var buf []byte = make([]byte, 1500)
	for {
		n, address, err := conn.ReadFromUDP(buf)

		if err != nil {
			log.Println("error reading data from connection")
			log.Println(err)
			return
		}

		if address != nil {

			log.Println("got message from ", address, " with n = ", n)

			if n > 0 {
				log.Println("from address", address, "got message:", string(buf[0:n]), n)
			}
		}
	}
}

func main() {

	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "Usage:%s host:port", os.Args[0])
		os.Exit(1)
	}

	service := os.Args[1]
	log.Println("Connecting to server at ", service)
	conn, err := net.Dial("udp", service)
	if err != nil {
		log.Println("Could not resolve udp address or connect to it  on ", service)
		log.Println(err)
		return
	}
	go listen(conn.(*net.UDPConn))
	log.Println("Connected to server at ", service)
	defer conn.Close()

	reader := bufio.NewReader(os.Stdin)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			log.Println("Error reading from stdin")
			continue
		}

		n, err := conn.Write([]byte(line))
		if err != nil {
			log.Println("error writing data to server", service)
			log.Println(err)
			return
		}

		if n > 0 {
			log.Println("Wrote ", n, " bytes to server at ", service)
		}
	}

}
