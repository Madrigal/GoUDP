package main

import (
	"bufio"
	"encoding/xml"
	"errors"
	"fmt"
	"log"
	"message"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	DEFAULT_ADDR        = "127.0.0.1:1200"
	MAX_USR             = 50000
	MAX_CONN            = 5000
	MILIS_BETWEEN_RETRY = 200
	MAX_RETRY           = 3
	BLOCKED_INITIAL     = 10
)

// Each incoming connection will have a message with whatever they want to send
// and who sent it
type Message struct {
	Content   []byte
	Sender    *net.UDPAddr
	Timestamp time.Time
}

type User struct {
	Alias   string
	Address *net.UDPAddr
	Online  bool
	Blocked []string
}

// Global
// Now just a map of addresses
var users map[string]User
var connections map[string]User

// This really simplifies things, although is kind of ugly
var serverConn *net.UDPConn
var clientConn *net.UDPConn

func init() {
	users = make(map[string]User, MAX_USR)
	connections = make(map[string]User, MAX_CONN)
}

func main() {

	var port string
	fmt.Println(message.NewLogin("Test"))

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
		log.Fatal("error resolving UDP address on ", port, err)
	}

	conn, err := net.ListenUDP("udp", udpAddress)

	if err != nil {
		log.Fatal("error listening on UDP port ", port, err)
	}
	serverConn = conn
	log.Println("Listening on ", udpAddress)
	defer conn.Close()
	go client(port)
	read := handleIncoming()
	for {
		fmt.Println(read)
	}
}

func client(port string) {
	conn, err := net.Dial("udp", port)
	if err != nil {
		// TODO Probably retry
		log.Fatal("Couldn't connect to port", port, err)
	}
	clientConn = conn.(*net.UDPConn)
	go getUserInput()
}

//This is going to be on the main loop and will basically be our user interface
// TODO This address is going to change when we change the server
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

func handleUserInput(line string) {
	line = strings.TrimRight(line, "\n")
	fmt.Println("Hey", line)
	clientConn.Write([]byte(line))
}

func firstRun(rd *bufio.Reader) {
	// Get the user to authenticate with the server
	for {
		fmt.Println("Welcome to the server. Please choose a nickname")
		nickname, err := rd.ReadString('\n')
		if err != nil {
			fmt.Println("Something went wrong reading your input. Try again")
			log.Println("Error on user input", err)
			continue
		}
		sendLogin(nickname)
	}
}

func sendLogin(nickname string) {
	loginMessage := message.NewLogin(nickname)
	// Send message
	byteMessage, _ := xml.Marshal(loginMessage)
	sendToServer(byteMessage)
}

func getUserOption(rd *bufio.Reader) int {
	// Keep asking until we got a correct input
	for {
		fmt.Println("Welcome to the server. Tell us what you want to do: ")
		fmt.Println("1: Send broadcast")
		fmt.Println("2: Send private message")
		fmt.Println("3: Get connected users")
		fmt.Println("4: Publish status to FB")
		fmt.Println("5: Block a user")
		fmt.Println("6: Exit")
		in, err := rd.ReadString('\n')
		if err != nil {
			fmt.Println("Something went wrong reading your input. Try again")
			log.Println("Error on user input", err)
			continue
		}
		opt, err := strconv.Atoi(in)
		if err != nil {
			fmt.Println("Couldn't get a number from what you wrote", err)
			log.Println("Error in atoi", err)
			continue
		}
		// TODO check that it is between our options
		return opt
	}
}

// TODO handle should know what to do when you need to become the server
func handleIncoming() <-chan Message {
	read := listenServer()
	for {
		message := <-read
		if message.Content == nil {
			continue
		}
		fmt.Println("Content", string(message.Content))
		fmt.Println("From address", *message.Sender)
		fmt.Println("In time", message.Timestamp)
		msg := []byte("OK")
		err := sendConfirmation(message.Sender, msg)
		if err != nil {
			// Assume he went offline
			log.Println("Couldn't write message ", string(msg), "to ", message.Sender)
			// TODO just do this if you are server
			disconnectUser(message.Sender)

		}
		// TODO Only servers care about this
		// Discover new connections
		var usr User
		user, ok := isUserConnected(message.Sender)
		if !ok {
			// Means we haven't seen him before
			fmt.Println("Registring new user", message.Sender)
			usr = registerUser(message.Sender)
		} else {
			usr = user
			fmt.Println("User already connected", usr)
		}

	}
}

func sendToServer(msg []byte) error {
	_, err := serverConn.Write(msg)
	if err != nil {
		return err
	}
	// TODO Wait for confirmation
	return nil
}

func isUserConnected(who *net.UDPAddr) (User, bool) {
	val, ok := connections[who.String()]
	return val, ok
}

func registerUser(who *net.UDPAddr) User {
	// TODO Get user login, probably with a login message
	alias := "Pepito"

	// Check that he doesn't exist already
	// _, ok := users[alias]
	usr := User{alias, who, true, make([]string, BLOCKED_INITIAL)}
	users[alias] = usr
	connections[who.String()] = usr
	return usr
}

func amITheServer() bool {
	// TODO
	return true
}

func disconnectUser(who *net.UDPAddr) {
	usr, ok := connections[who.String()]
	if ok {
		// User already known, set as offline
		usr.Online = false
	}
	delete(connections, who.String())
}

// sendConfirmation tries to send a confirmation to the user who
// sent the message. If it doesn't get any confirmation it sends an error
func sendConfirmation(whom *net.UDPAddr, msg []byte) error {
	retriesLeft := MAX_RETRY
	// Send confirmation
	for retriesLeft > 0 {
		_, err := serverConn.WriteTo(msg, whom)
		if err == nil {
			return nil
		}
		time.Sleep(time.Millisecond * MILIS_BETWEEN_RETRY)
		retriesLeft--
	}
	return errors.New("Couldn't send confirmation")
}

// listen handles any incomming connections and writes them to the channel
// It doesn't deal with confirmation nor validation
func listenServer() <-chan Message {
	c := make(chan Message)
	go func() {
		buff := make([]byte, 1024)

		for {
			n, addr, err := serverConn.ReadFromUDP(buff)
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
				// Send to the channel
				c <- m
			}
			if err != nil {
				// If it fails send a nil message
				c <- Message{nil, nil, time.Now()}
				break
			}
		}
	}()
	return c
}
