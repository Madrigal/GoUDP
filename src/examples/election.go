package main

import (
	"bufio"
	"encoding/xml"
	"fmt"
	"log"
	"message"
	"net"
	"os"
	"time"
)

var myAddress int
var knownAddresess map[int]bool
var listenMulticast *net.UDPConn
var writeMulticast *net.UDPConn
var multicastAddr *net.UDPAddr

// We will always be listening for new petitions in here.
// Either you or some other process who sends you an address
// will trigger a voting.
// After some time elapses someone will be elected the server

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
	listenMulticast = conn
	writeMulticast = lconn
	multicastAddr = mcaddr
	go listen(listenMulticast)
	go startVoting()
	for {
		// Sleep 20 seconds to give time to spawn more clients
		// time.Sleep(time.Second * 10)
		txt, _, err := reader.ReadLine()
		b := make([]byte, 256)
		copy(b, txt)
		check(err)
		_, err = writeMulticast.WriteToUDP(b, mcaddr)
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

func startVoting() {
	fmt.Println("Start voting!")
	time.Sleep(time.Second * 10)
	// Send you adress to all. In this case it really doesn't matter
	msg := message.NewVoteMessage(myAddress)
	mm, _ := xml.Marshal(msg)
	_, err := writeMulticast.WriteToUDP(mm, multicastAddr)
	check(err)

	// Wait 10 seconds
	c := time.After(time.Second * 10)
	<-c

	// Check if we got any address that is bigger
	for addr, _ := range knownAddresess {
		if addr > myAddress {
			return
		}
	}

	// Become the server if we are the greatest
	msga := message.NewCoordinatorMessage(string(myAddress))
	mmm, _ := xml.Marshal(msga)
	writeMulticast.WriteToUDP(mmm, multicastAddr)
	fmt.Println("NOW I AM BECOME DEATH")
}

func check(err error) {
	if err != nil {
		fmt.Println("error", err)
		os.Exit(1)
	}
}
