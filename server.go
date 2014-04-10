package main

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"time"
)

const (
	DEFAULT_ADDR = "127.0.0.1:1200"
)

func main() {

	var port string

	if len(os.Args) == 2 {
		// Override default port
		port = os.Args[1]
	} else {
		fmt.Println("You can change the default port passing as argument host:port")
		fmt.Println("e.g. 127.0.0.1:9000")
		port = DEFAULT_ADDR
	}

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
	log.Println("Listening on ", udpAddress)
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
		if message.Content == nil {
			continue
		}
		fmt.Println("Content", string(message.Content))
		fmt.Println("From address", *message.Sender)
		fmt.Println("In time", message.Timestamp)

		sendConfirmation(conn, message.Sender, []byte("OK"))
	}
}

func sendConfirmation(conn *net.UDPConn, whom *net.UDPAddr, msg []byte) error {
	retriesLeft := 3
	// Send confirmation
	for retriesLeft > 0 {
		_, err := conn.WriteTo(msg, whom)
		if err == nil {
			return nil
		}
		time.Sleep(time.Millisecond * 200)
	}
	return errors.New("Couldn't send confirmation")
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
