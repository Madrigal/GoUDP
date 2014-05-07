package main

import (
	"bufio"
	"encoding/xml"
	"errors"
	"flag"
	"fmt"
	"log"
	"message"
	"net"
	"os"
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
	Pending [][]byte
}

// Global
// Now just a map of addresses

// Map of aliases
var users map[string]*User

// Map of addresses
var connections map[string]*User

// This really simplifies things, although is kind of ugly
var serverConn *net.UDPConn
var clientConn *net.UDPConn
var GlobalPort string

// A petition to start the server. For now just holds the port to connect to
type ServerPetition struct {
	Port string
}

var myAlias string

var startServer chan ServerPetition
var stopServer chan ServerPetition
var sendingChannel chan []byte

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

// To check if a user is connected don't just reserve logins. Use the login message to map a user to a connection,
// but a user can connect from multiple addresses. Is just as a login without passwords ;)
func init() {
	users = make(map[string]*User, MAX_USR)
	connections = make(map[string]*User, MAX_CONN)
	startServer = make(chan ServerPetition, 1)
	stopServer = make(chan ServerPetition, 1)
	sendingChannel = make(chan []byte)
}

func main() {
	portPtr := flag.String("port", DEFAULT_ADDR, "port to bind to")
	serverPtr := flag.Bool("s", false, "Wheter this instance should become the server")
	flag.Parse()

	shouldBeServer := *serverPtr
	port := *portPtr
	GlobalPort = port
	confirmationChan := make(chan []byte)
	go serverControl()
	go sendDataToServer(sendingChannel, confirmationChan)
	if shouldBeServer {
		s := ServerPetition{port}
		startServer <- s
	}

	// Always create a client
	client(port, confirmationChan)
}

func serverControl() {
	for {
		select {
		case b := <-startServer:
			conn := initServer(b.Port)
			serverConn = conn
			defer conn.Close()

			// Handle incoming messages and loop forever
			go handleIncoming()

		case <-stopServer:
			killServer()
		}
	}
}

func killServer() {
	var retries int32
	var err error
	// Sanity check
	if serverConn == nil {
		log.Println("Trying to stop a non started server!")
		return
	}
	for retries = 3; retries > 0; retries-- {
		err = serverConn.Close()
		if err == nil {
			break
		}
		time.Sleep(500 * time.Millisecond)
	}
	if retries == 0 {
		panic("Couldn't stop the server, reason" + err.Error())
	}
}

func initServer(port string) *net.UDPConn {
	log.Println("Starting server")
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

func client(port string, confirmation chan []byte) {
	log.Println("Starting client")
	retries := 3
	var err error
	var con net.Conn
	var conn *net.UDPConn
	for ; retries > 0; retries-- {
		con, err = net.Dial("udp", port)
		conn = con.(*net.UDPConn)
		if err == nil {
			break
		} else {
			// Give some time to the server to setup
			time.Sleep(500 * time.Millisecond)
			log.Println("Failing because", err)
		}
	}

	if retries == 0 {
		// If we couldn't create a client we are useless
		log.Fatal("Couldn't connect to port", port, err)
	}
	clientConn = conn
	c := make(chan []byte)
	go listenClient(clientConn, c)
	go handleClient(c, confirmation)
	getUserInput()
}

//This is going to be on the main loop and will basically be our user interface
// TODO This address is going to change when we change the server
func getUserInput() {
	fmt.Println("Put your input!")
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

func sendXmlToServer(xmlMessage interface{}) {
	bytes, err := xml.Marshal(xmlMessage)
	if err != nil {
		fmt.Println("Error marshaling", err)
	}
	sendingChannel <- bytes
}

func sendDataToServer(sending chan []byte, confirmation chan []byte) {
	for {
		bytes := <-sending
		fmt.Println("From send data to server ", string(bytes))
		clientConn.Write(bytes)
		<-confirmation
		fmt.Println("Got confirmation")
	}
}

func handleUserInput(line string) {
	line = strings.TrimRight(line, "\n")
	arr := strings.Split(line, " ")
	l := arr[0]
	length := len(arr)
	switch {
	case l == "/nick":
		if length <= 1 {
			fmt.Println("Missing arguments")
			return
		}
		nick := arr[1]
		m := message.NewLogin(nick)
		// You can do this but the server will
		// reject you if there is an error
		myAlias = nick
		sendXmlToServer(m)

	case l == "/names":
		m := message.NewUGetConnected()
		sendXmlToServer(m)

	case l == "/msg":
		if length <= 2 {
			fmt.Println("Missing arguments")
			return
		}
		to := arr[1]
		msg := strings.Join(arr[2:length], " ")
		m := message.NewDirectMessage(to, msg)
		sendXmlToServer(m)

	case l == "/send":
		if length <= 3 {
			fmt.Println("Missing arguments")
			return
		}
		to := arr[1]
		filename := arr[2]
		// TODO
		fmt.Println("/SEND", "to", to, "filename", filename)

	case l == "/block":
		if length <= 1 {
			fmt.Println("Missing arguments")
			return
		}
		who := arr[1]
		m := message.NewBlock(myAlias, who)
		sendXmlToServer(m)

	case l == "/fb":
		if length <= 2 {
			fmt.Println("Missing arguments")
			return
		}
		message := strings.Join(arr[1:length], " ")
		// TODO
		fmt.Println("/FB", message)

	case l == "/quit":
		m := message.NewExit()
		sendXmlToServer(m)

	case l == "/admin":
		if length != 2 {
			fmt.Println("Incorrect arguments")
			return
		}
		action := arr[1]
		switch {
		case action == "start":
			s := ServerPetition{GlobalPort}
			startServer <- s
		case action == "stop":
			s := ServerPetition{}
			stopServer <- s
		default:
			fmt.Println("Unkwon admin action")
		}

	default:
		m := message.NewBroadcast(line)
		sendXmlToServer(m)
	}

	fmt.Println("You wrote", line)

	// clientConn.Write([]byte(line))
}

func blockUser(blocker string, blocked string) {
	// Get both users
	I, ok := users[blocker]
	if !ok {
		fmt.Errorf("Couldn't get the current alias for blocking", blocker)
	}
	_, ok = users[blocked]
	if !ok {
		fmt.Errorf("You can't block a user that is not registered!, you tried to block", blocked)
	}

	I.Blocked = append(I.Blocked, blocked)
}

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
			fmt.Println("Got", string(m.Content))
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
			directMessageHandler(internalM)

		case message.GET_CONN_T:
			getConnectedHandler(internalM)

		case message.BLOCK_T:
			blockHandler(internalM)

		case message.EXIT_T:
			exitHandler(internalM)

		}
	}
}

func sendError(who *net.UDPAddr, msg string) {
	fmt.Println("Sending error to ", who.String())
	fmt.Println("Message", msg)
	errMsg := message.NewErrorMessage(msg)
	m, err := xml.Marshal(errMsg)
	if err != nil {
		// server error
		log.Println("Server error sending broadcast", err.Error())
		return
	}
	// Don't buffer error messages
	sendMessage(who, m)
}

func isUserConnected(who *net.UDPAddr) (*User, bool) {
	val, ok := connections[who.String()]
	return val, ok
}

/// Handlers
func loginHandler(m InternalMessage) {
	usr, ok := isUserConnected(m.Sender)
	if ok {
		// User is already connected, meanning he already picked an alias
		// don't do anything.
		// Maybe this could be an error, but it seems to much
		log.Println("User already connected", usr)
	} else {
		// Means we haven't seen him before
		fmt.Println("Registring new connection", m.Sender)
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

func directMessageHandler(m InternalMessage) {
	dm := m.Content.UMessage
	// Get the alias of the sender
	alias, err := getUserAlias(m.Sender)
	if err != nil {
		sendError(m.Sender, "Fail to send broadcast, reason"+err.Error())
	}
	// Create new message
	msg := message.NewSDirectMessage(alias, dm.Message)

	// Get a reference to the user we are sending the message
	reciever, ok := users[dm.To]
	if !ok {
		sendError(m.Sender, "The user"+dm.To+"Doesn't exist!")
	}

	// send it!
	mm, err := xml.Marshal(msg)
	if err != nil {
		log.Println("Error marshaling dm, reason", err.Error())
	}
	sendMessaeToUserCheckBlocked(reciever, alias, mm)
	// sendMessageToUser(reciever, mm)
}

func getConnectedHandler(m InternalMessage) {
	// Get all the alias of connected users
	// TODO This is probably very expensive, maybe should keep a cache of this
	// but maybe is not worthy
	connectedUsers := make([]string, len(connections))
	i := 0
	for _, usr := range connections {
		connectedUsers[i] = usr.Alias
		i++
	}

	// Make the response
	msg := message.NewSGetConnected(connectedUsers)

	// Prepare the message to be sent
	mm, err := xml.Marshal(msg)
	if err != nil {
		log.Println("Error marshaling getConnected, reason", err.Error())
	}

	// Get reference to the user who sent this
	usr, ok := connections[m.Sender.String()]
	if !ok {
		log.Println("Fuck!!!")
	}

	// Send it!
	sendMessageToUser(usr, mm)
}

func blockHandler(m InternalMessage) {
	block := m.Content.Block
	blockUser(block.Blocker, block.Blocked)
}

func exitHandler(m InternalMessage) {
	// I guess that's it
	disconnectUser(m.Sender)
}

func sendBroadcast(broadcastMessage *message.SMessage) {
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
		// TODO Check blocked
		sendMessaeToUserCheckBlocked(usr, broadcastMessage.From, m)
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

// registerUser assumes that a user already was already chec
func registerUser(who *net.UDPAddr, loginMessage *message.Login) error {
	alias := loginMessage.Nickname
	// Check that he doesn't exist already
	var usr *User
	usr, isAlreadyRegistered := users[alias]
	if isAlreadyRegistered {
		if usr.Online {
			// That login is already used, choose a different one
			sendError(who, "Login already taken, choose a different one")
			return errors.New("Login already taken")
		}

		// Update to new status
		usr.Address = who
		usr.Online = true
		sendPendingMessages(usr)

	} else {
		// Create a new user
		usr = &User{alias,
			who,
			true,
			make([]string, BLOCKED_INITIAL),
			make([][]byte, 100),
		}
		users[usr.Alias] = usr
	}
	connections[who.String()] = usr
	fmt.Println("Connections", connections)
	return nil
}

func disconnectUser(who *net.UDPAddr) {
	usr, ok := connections[who.String()]
	if ok {
		// User already known, set as offline
		usr.Online = false
	}
	delete(connections, who.String())
}

func sendPendingMessages(usr *User) {
	pending := usr.Pending
	for _, message := range pending {
		sendMessageToUser(usr, message)
	}
}

func sendMessaeToUserCheckBlocked(to *User, sender string, msg []byte) error {
	// Iterate and check if the user is blocked
	for _, alias := range to.Blocked {
		if alias == sender {
			return nil
		}
	}
	return sendMessageToUser(to, msg)
}

func sendMessageToUser(usr *User, msg []byte) error {
	// See if the user is connected
	if usr.Online {
		// If he is, try to send message
		err := sendMessage(usr.Address, msg)
		if err == nil {
			return nil
		}
	}
	// If user is not currently connected or sent failed try to save it for later
	err := saveMessageForLater(usr, msg)
	if err != nil {
		return err
	}
	return nil
}

func saveMessageForLater(usr *User, msg []byte) error {
	usr.Pending = append(usr.Pending, msg)
	return nil
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

func handleClient(c <-chan []byte, confirmation chan<- []byte) {
	fmt.Println("In handle client")
	for {
		b := <-c
		fmt.Println("From handle client", string(b))
		if message.IsConfirmation(b) {
			confirmation <- b
			continue
		}
		fmt.Println("Decoding user message")
		t, m, err := message.DecodeServerMessage(b)
		if err != nil {
			fmt.Println("Error reading XML from server")
			fmt.Println("Got", string(b))
			continue
		}
		switch t {
		case message.ERROR_T:
			fmt.Println("Error from server:", m.Error.Message)
		case message.DM_T:
			msg := m.Direct
			fmt.Println("Message from ", msg.From, ": ", msg.Message)
		case message.BROAD_T:
			msg := m.Direct
			fmt.Println("Broadcast from ", msg.From, ": ", msg.Message)

		case message.GET_CONN_T:
			msg := m.Connected
			fmt.Println("Connected users")
			users := msg.Users.ConnUsers
			for _, usr := range users {
				fmt.Println("-", usr)
			}
		}
	}
}

func listenClient(conn *net.UDPConn, c chan<- []byte) {
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
			c <- res
		}
		if err != nil {
			fmt.Println("Error reading from server", err)
		}
	}
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
