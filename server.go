package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"time"
)

func main() {

	port := "127.0.0.1:1200"

	udpAddress, err := net.ResolveUDPAddr("udp4", port)

	if err != nil {
		log.Println("error resolving UDP address on ", port)
		log.Println(err)
		return
	}

	conn, err := net.ListenUDP("udp", udpAddress)

	if err != nil {
		log.Println("error listening on UDP port ", port)
		log.Println(err)
		return
	}
	log.Println("Got a connection")
	defer conn.Close()
	go getUserInput()
	read := handle(conn)
	for {
		fmt.Println(read)
	}
}

//This is going to be on the main loop and will basically be our user interface
func getUserInput() {
	reader := bufio.NewReader(os.Stdin)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			log.Println("Error reading from stdin")
			continue
		}
		handleUserInput(line)
	}
}

// TODO this should probably get the connection too
func handleUserInput(line string) {
	fmt.Println(strings.TrimRight(line, "\n"))
}

// Each incoming connection will have a message with whatever they want to send
// and who sent it
type Message struct {
	Content   []byte
	Sender    *net.UDPAddr
	Timestamp time.Time
}

// TODO handle should know what to do when you need to become the server
func handle(conn *net.UDPConn) <-chan Message {
	read := listen(conn)
	for {
		message := <-read
		if message.Content != nil {
			fmt.Println("Content", string(message.Content))
			fmt.Println("From address", *message.Sender)
			fmt.Println("In time", message.Timestamp)
		}
	}
}

func listen(conn *net.UDPConn) <-chan Message {
	c := make(chan Message)
	go func() {
		buff := make([]byte, 1024)

		for {
			n, addr, err := conn.ReadFromUDP(buff)
			if n > 0 && addr != nil {
				// Copy the response
				res := make([]byte, n)

				// Trim newline
				if string(buff[n-1]) == "\n" {
					copy(res, buff[:n-1])
				} else {
					copy(res, buff[:n])
				}

				// Create the message
				m := Message{Content: res,
					Sender:    addr,
					Timestamp: time.Now()}
				c <- m
			}
			if err != nil {
				c <- Message{nil, nil, time.Now()}
				break
			}
		}
	}()
	return c
}
