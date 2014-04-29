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

type InternalMessage struct {
	Type      message.Type
	Content   *message.UserPackage
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

// Brain rant
// We need to get several channels
// Server:
// 1. Read incoming messages
// 2. Validate them (e.g. correct structure, that they are not repeated)
// 3. Map them to an action
// 4. Execute that action
// Client:
// 1. Get what the user wants to do
// 2. Map it to an action (e.g. create a new login message)
// 3. Send the message to the server, wait for confirmation and resend if neccesary
//    (all-in-one)
// Additional stuff:
// 1. When a certain treshold of not delivered messages is reached the client should start
//    a "select the server" mechanism, which will result in getting a server address. If
//    this address is the same as yours (not sure about this) become the server
// 2. Synchronization of clocks.

func init() {
	users = make(map[string]User, MAX_USR)
	connections = make(map[string]User, MAX_CONN)
}

func main() {
	// Get port where we are going to listen
	var port string
	if len(os.Args) == 2 {
		// Override default port if given
		port = os.Args[1]
	} else {
		fmt.Println("You can change the default port passing as argument host:port")
		fmt.Println("e.g. 127.0.0.1:9000")
		port = DEFAULT_ADDR
	}

	// Create server connection
	conn := initServer(port)
	serverConn = conn
	defer conn.Close()

	// Create client connection on same port
	go client(port)

	// Handle incoming messages and loop forever
	read := handleIncoming()
	for {
		fmt.Println(read)
	}
}

func initServer(port string) *net.UDPConn {
	udpAddress, err := net.ResolveUDPAddr("udp4", port)
	if err != nil {
		log.Fatal("error resolving UDP address on ", port, err)
	}
	conn, err := net.ListenUDP("udp", udpAddress)
	if err != nil {
		log.Fatal("error listening on UDP port ", port, err)
	}
	log.Println("Listening on ", udpAddress)
	return conn
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
		m := <-read
		if m.Content == nil {
			continue
		}
		fmt.Println("Content", string(m.Content))
		fmt.Println("From address", *m.Sender)
		fmt.Println("In time", m.Timestamp)
		msg := []byte("OK")
		err := sendMessage(m.Sender, msg)
		if err != nil {
			// Assume he went offline
			log.Println("Couldn't write message ", string(msg), "to ", m.Sender)
			disconnectUser(m.Sender)

		}
		// Convert to internal message
		t, p, err := message.DecodeUserMessage(m.Content)
		if err != nil {
			fmt.Println("Error reading XML. Please check it")
			sendError(m.Sender, "Error reading XML. Please check it")
			continue
		}
		internalM := InternalMessage{
			Type:      t,
			Content:   p,
			Sender:    m.Sender,
			Timestamp: m.Timestamp,
		}
		// Dispatch
		switch internalM.Type {
		case message.UNKNOWN_T:
			sendError(internalM.Sender, "Couldn't match type to any know type")

		case message.LOGIN_T:
			loginHandler(internalM)

		case message.BROAD_T:
			broadcastHandler(internalM)

		case message.DM_T:
			fmt.Println("<<Type", message.DM_T)
		case message.GET_CONN_T:
			fmt.Println("<<Type", message.GET_CONN_T)
		case message.EXIT_T:
			fmt.Println("<<Type", message.EXIT_T)
		}
		fmt.Println(internalM)

	}
}

func sendError(who *net.UDPAddr, msg string) {
	fmt.Println("Sending error to ", who.String())
	fmt.Println("Message", msg)
	errMsg := message.NewErrorMessage(msg)
	fmt.Println(errMsg)
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

//// Server get message types
func loginHandler(m InternalMessage) {
	usr, ok := isUserConnected(m.Sender)
	if ok {
		// User is already connected, don't do anything
		// Maybe this could be an error, but it seems to much
		fmt.Println("User already connected", usr)
	} else {
		// Means we haven't seen him before
		fmt.Println("Registring new user", m.Sender)
		err := registerUser(m.Sender, m.Content.Login)
		if err != nil {
			sendError(m.Sender, err.Error())
		}
	}
}

func broadcastHandler(m InternalMessage) {
	// Create a broadcastMessage
	alias, err := getUserAlias(m.Sender)
	if err != nil {
		sendError(m.Sender, "Fail to send broadcast, reason"+err.Error())
	}
	msg := message.NewSBroadcast(alias, m.Content.UMessage.Message)
	fmt.Println(msg)
	sendBroadcast(&msg)
}

func sendBroadcast(broadcastMessage *message.SMessage) {
	// TODO check if it's for every registered user or every connected user
	m, err := xml.Marshal(broadcastMessage)
	if err != nil {
		// server error
		log.Println("Server error sending broadcast", err.Error())
		return
	}
	for _, usr := range connections {
		if usr.Alias == broadcastMessage.From {
			continue
		}
		log.Println("sending data", broadcastMessage, "to user", usr.Alias)
		sendMessage(usr.Address, m)
	}
}

// It's not found because we are using a different thing
func getUserAlias(who *net.UDPAddr) (string, error) {
	usr, ok := connections[who.String()]
	if !ok {
		return "", errors.New("Your user wasn't found. Please login first")
	}
	return usr.Alias, nil
}

func registerUser(who *net.UDPAddr, loginMessage *message.Login) error {

	alias := loginMessage.Nickname
	// Check that he doesn't exist already
	_, isAlreadyRegistered := users[alias]
	if isAlreadyRegistered {
		return errors.New("Nickname already registered, please choose a different one")
	}
	usr := User{alias, who, true, make([]string, BLOCKED_INITIAL)}
	users[usr.Alias] = usr
	connections[who.String()] = usr
	return nil
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

// sendMessage tries to send a confirmation to the user who
// sent the message. If it doesn't get any confirmation it sends an error
func sendMessage(whom *net.UDPAddr, msg []byte) error {
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
