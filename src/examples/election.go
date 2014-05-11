package main

import (
	"bufio"
	"encoding/xml"
	"fmt"
	"log"
	"message"
	"net"
	"os"
)

var myAddress int
var knownAddresess map[int]bool

func main() {
	port := "224.0.1.60:1888"
	laddr, err := net.ResolveUDPAddr("udp", ":0")
	check(err)
	mcaddr, err := net.ResolveUDPAddr("udp", port)
	check(err)
	conn, err := net.ListenMulticastUDP("udp", nil, mcaddr)
	check(err)
	lconn, err := net.ListenUDP("udp", laddr)
	check(err)
	reader := bufio.NewReader(os.Stdin)

	knownAddresess = make(map[int]bool, 1)
	myAddress = os.Getpid()
	log.Println("My address is ", myAddress)
	msg := message.NewVoteMessage(myAddress)
	mm, _ := xml.Marshal(msg)
	fmt.Println(string(mm))
	go listen(conn)
	for {
		// Sleep 20 seconds to give time to spawn more clients
		// time.Sleep(time.Second * 10)
		txt, _, err := reader.ReadLine()
		b := make([]byte, 256)
		copy(b, txt)
		check(err)
		_, err = lconn.WriteToUDP(b, mcaddr)
		check(err)
	}
}

func listen(conn *net.UDPConn) {
	for {
		b := make([]byte, 256)
		_, _, err := conn.ReadFromUDP(b)
		check(err)
		fmt.Println("read", string(b))
		t, m, err := message.DecodeClientToClientMessage(b)
		if err != nil {
			log.Println("[Client] Can't decode message", string(b))
		}
		switch t {
		case message.VOTE_T:
			address := m.VoteMessage.Number
			fmt.Println("Got this address", address)
			if address != myAddress {
				// Add Adress to list of known address
				knownAddresess[address] = true
				fmt.Println("Known address", knownAddresess)
			}
		case message.COORDINATOR_T:
			Newaddr := m.CoordinatorMessage.Address
			fmt.Println("Coordinator", Newaddr)
		}
		log.Println("[Client] Got", m)
	}
}

func check(err error) {
	if err != nil {
		fmt.Println("error", err)
		os.Exit(1)
	}
}
