package message

import (
	"encoding/xml"
	"errors"
	"time"
)

const (
	LOGIN     = "Login"
	BROAD     = "Broadcast"
	DM        = "DirectMessage"
	GET_CONN  = "GetConnected"
	EXIT      = "Exit"
	ERROR     = "Error"
	BLOCK     = "Block"
	FILE      = "FILE"
	CLOCK     = "Clock"
	OFFSET    = "TimeOffset"
	ADDRESS   = "Address"
	LOGIN_RES = "LoginResponse"
)

type Type int

const (
	UNKNOWN_T   Type = iota
	ERROR_T     Type = iota
	LOGIN_T     Type = iota
	BROAD_T     Type = iota
	DM_T        Type = iota
	GET_CONN_T  Type = iota
	BLOCK_T     Type = iota
	FILE_T      Type = iota
	CLOCK_T     Type = iota
	OFFSET_T    Type = iota
	EXIT_T      Type = iota
	ADDRESS_T   Type = iota
	LOGIN_RES_T Type = iota
)

type Base struct {
	Type string `xml:"Type"`
}

// Message sent from the client when he wants to login
type Login struct {
	XMLName xml.Name `xml:"Root"`
	Base
	Nickname string `xml:"Nickname"`
}

type LoginResponse struct {
	XMLName xml.Name `xml:"Root"`
	Base
	Address int `xml:"address"`
}

// Message a user sends to server. It covers both
// Broadcast and direct message
type UMessage struct {
	XMLName xml.Name `xml:"Root"`
	Base
	To      string `xml:"To"`
	Message string `xml:"Message"`
}

// Message the server will sent to a user
type SMessage struct {
	XMLName xml.Name `xml:"Root"`
	Base
	From    string `xml:"From"`
	Message string `xml:"Message"`
}

// When the user request connected users, he will
// only specify that as type. Hence no need for more fields
type UGetConnected struct {
	XMLName xml.Name `xml:"Root"`
	Base
}

type SGetConnected struct {
	Base
	XMLName xml.Name `xml:"Root"`
	Users   Users
}

type Users struct {
	XMLName   xml.Name
	ConnUsers []GetConnUser `xml:"User"`
}

type GetConnUser struct {
	Id string `xml:",innerxml"`
}

// Since user will only specify exit no extra fields are needed
type UExit struct {
	XMLName xml.Name `xml:"Root"`
	Base
}

type Block struct {
	XMLName xml.Name `xml:"Root"`
	Base
	Blocker string
	Blocked string
}

type ErrorMessage struct {
	XMLName xml.Name `xml:"Root"`
	Base
	Message string `xml:"Message"`
}

// ClockMessage is send by the server
// so clients will respond with
// the offset from this time
type ClockSyncPetition struct {
	XMLName xml.Name `xml:"Root"`
	Base
	Time time.Time `xml:"Time"`
}

// Ths is the message that the client sends to the server
// saying "Im n duration ahead/behind you", and server
// respons with this message to make adjustments
type ClockOffset struct {
	XMLName xml.Name `xml:"Root"`
	Base
	Offset time.Duration `xml:"offset"`
}

type AddressMessage struct {
	XMLName xml.Name `xml:"Root"`
	Base
	Address int `xml:"Address"`
}

// This type will decode an incoming message
// The UserPackage will hold the actual values
type UserMessage struct {
	Base
	Data interface{}
}

type ServerMessage struct {
	Base
	Data interface{}
}

// Client-to-server
type UserPackage struct {
	Block         *Block
	Login         *Login
	UMessage      *UMessage
	UGetConnected *UGetConnected
	UExit         *UExit
	File          *FileMessage
	Clock         *ClockOffset
}

// Server-to-client
type ServerPackage struct {
	Direct    *SMessage
	Connected *SGetConnected
	Block     *Block
	Error     *ErrorMessage
	File      *FileMessage
	Clock     *ClockSyncPetition
	Offset    *ClockOffset
	Address   *AddressMessage
	Login     *LoginResponse
}

// Decode message FROM the server
func DecodeServerMessage(msg []byte) (Type, *ServerPackage, error) {
	var m ServerMessage
	err := xml.Unmarshal(msg, &m)
	if err != nil {
		return UNKNOWN_T, nil, err
	}
	switch m.Type {
	case LOGIN_RES:
		var b LoginResponse
		err := xml.Unmarshal(msg, &b)
		if err != nil {
			return UNKNOWN_T, nil, errors.New("Couldn't decode the message: Broadcast or direct message malformed")
		}
		mp := ServerPackage{
			Login: &b,
		}
		return LOGIN_RES_T, &mp, nil

	case BROAD, DM:
		var b SMessage
		err := xml.Unmarshal(msg, &b)
		if err != nil {
			return UNKNOWN_T, nil, errors.New("Couldn't decode the message: Broadcast or direct message malformed")
		}
		mp := ServerPackage{
			Direct:    &b,
			Connected: nil,
			Error:     nil,
		}
		if m.Type == BROAD {
			return BROAD_T, &mp, nil
		}
		return DM_T, &mp, nil
	case GET_CONN:
		var u SGetConnected
		err := xml.Unmarshal(msg, &u)
		if err != nil {
			return UNKNOWN_T, nil, errors.New("Couldn't decode the message: Get connected malformed")
		}
		mp := ServerPackage{
			Direct:    nil,
			Connected: &u,
			Error:     nil,
		}
		return GET_CONN_T, &mp, nil

	// TODO think this isn't used
	case BLOCK:
		var b Block
		err := xml.Unmarshal(msg, &b)
		if err != nil {
			return UNKNOWN_T, nil, errors.New("Couldn't decode the message: Block message malformed")
		}
		mp := ServerPackage{
			Block: &b,
		}
		return BLOCK_T, &mp, nil

	case FILE:
		var f FileMessage
		err := xml.Unmarshal(msg, &f)
		if err != nil {
			return UNKNOWN_T, nil, errors.New("Couldn't decode the message: File message malformed")
		}
		mp := ServerPackage{
			File: &f,
		}

		return FILE_T, &mp, nil

	case CLOCK:
		var c ClockSyncPetition
		err := xml.Unmarshal(msg, &c)
		if err != nil {
			return UNKNOWN_T, nil, errors.New("Couldn't decode the message: File message malformed")
		}
		mp := ServerPackage{
			Clock: &c,
		}

		return CLOCK_T, &mp, nil

	case OFFSET:
		var c ClockOffset
		err := xml.Unmarshal(msg, &c)
		if err != nil {
			return UNKNOWN_T, nil, errors.New("Couldn't decode the message: File message malformed")
		}
		mp := ServerPackage{
			Offset: &c,
		}

		return OFFSET_T, &mp, nil

	case ADDRESS:
		var a AddressMessage
		err := xml.Unmarshal(msg, &a)
		if err != nil {
			return UNKNOWN_T, nil, errors.New("Couldn't decode the message: File message malformed")
		}
		sp := ServerPackage{
			Address: &a,
		}

		return ADDRESS_T, &sp, nil

	case ERROR:
		var u ErrorMessage
		err := xml.Unmarshal(msg, &u)
		if err != nil {
			return UNKNOWN_T, nil, errors.New("Couldn't decode the message: Exit message malformed")
		}
		mp := ServerPackage{
			Direct:    nil,
			Connected: nil,
			Error:     &u,
		}
		return ERROR_T, &mp, nil
	default:
		return UNKNOWN_T, nil, errors.New("Couldn't decode the message: No matching type")
	}
}

// temp function
func DecodeUserMessage(msg []byte) (Type, *UserPackage, error) {
	var m UserMessage
	err := xml.Unmarshal(msg, &m)
	if err != nil {
		return UNKNOWN_T, nil, err
	}
	switch m.Type {
	case LOGIN:
		var l Login
		err := xml.Unmarshal(msg, &l)
		if err != nil {
			return UNKNOWN_T, nil, errors.New("Couldn't decode the message: Login malformed")
		}
		up := UserPackage{
			Login:         &l,
			UMessage:      nil,
			UGetConnected: nil,
			UExit:         nil,
		}
		return LOGIN_T, &up, nil
	case BROAD, DM:
		var b UMessage
		err := xml.Unmarshal(msg, &b)
		if err != nil {
			return UNKNOWN_T, nil, errors.New("Couldn't decode the message: Broadcast or direct message malformed")
		}
		up := UserPackage{
			Login:         nil,
			UMessage:      &b,
			UGetConnected: nil,
			UExit:         nil,
		}
		if m.Type == BROAD {
			return BROAD_T, &up, nil
		}
		return DM_T, &up, nil
	case GET_CONN:
		var u UGetConnected
		err := xml.Unmarshal(msg, &u)
		if err != nil {
			return UNKNOWN_T, nil, errors.New("Couldn't decode the message: Get connected malformed")
		}
		up := UserPackage{
			Login:         nil,
			UMessage:      nil,
			UGetConnected: &u,
			UExit:         nil,
		}
		return GET_CONN_T, &up, nil

	case BLOCK:
		var b Block
		err := xml.Unmarshal(msg, &b)
		if err != nil {
			return UNKNOWN_T, nil, errors.New("Couldn't decode the message: Block message malformed")
		}
		up := UserPackage{
			Block: &b,
		}
		return BLOCK_T, &up, nil

	case FILE:
		var f FileMessage
		err := xml.Unmarshal(msg, &f)
		if err != nil {
			return UNKNOWN_T, nil, errors.New("Couldn't decode the message: File message malformed")
		}
		up := UserPackage{
			File: &f,
		}

		return FILE_T, &up, nil

	case OFFSET:
		var c ClockOffset
		err := xml.Unmarshal(msg, &c)
		if err != nil {
			return UNKNOWN_T, nil, errors.New("Couldn't decode the message: File message malformed")
		}
		up := UserPackage{
			Clock: &c,
		}

		return OFFSET_T, &up, nil

	case EXIT:
		var u UExit
		err := xml.Unmarshal(msg, &u)
		if err != nil {
			return UNKNOWN_T, nil, errors.New("Couldn't decode the message: Exit message malformed")
		}
		up := UserPackage{
			Login:         nil,
			UMessage:      nil,
			UGetConnected: nil,
			UExit:         &u,
		}
		return EXIT_T, &up, nil
	default:
		return UNKNOWN_T, nil, errors.New("Couldn't decode the message: No matching type")
	}
}

///// Client calls
func NewLogin(nickname string) Login {
	base := Base{Type: LOGIN}
	login := Login{Base: base, Nickname: nickname}
	return login
}

func NewBroadcast(msg string) UMessage {
	base := Base{Type: BROAD}
	message := UMessage{Base: base, To: "", Message: msg}
	return message
}

func NewDirectMessage(to string, msg string) UMessage {
	base := Base{Type: DM}
	message := UMessage{Base: base, To: to, Message: msg}
	return message
}

func NewUGetConnected() UGetConnected {
	base := Base{Type: GET_CONN}
	getConn := UGetConnected{Base: base}
	return getConn
}

func NewExit() UExit {
	base := Base{Type: EXIT}
	exit := UExit{Base: base}
	return exit
}

func NewBlock(who string, blocking string) Block {
	base := Base{Type: BLOCK}
	bm := Block{Base: base, Blocker: who, Blocked: blocking}
	return bm
}

func NewClockSyncPetition(t time.Time) ClockSyncPetition {
	base := Base{Type: CLOCK}
	cm := ClockSyncPetition{Base: base, Time: t}
	return cm
}

func NewClockOffset(t time.Duration) ClockOffset {
	base := Base{Type: OFFSET}
	co := ClockOffset{Base: base, Offset: t}
	return co
}

///// Server calls
func NewSBroadcast(from string, msg string) SMessage {
	base := Base{Type: BROAD}
	message := SMessage{Base: base, From: from, Message: msg}
	return message
}

func NewSDirectMessage(from string, msg string) SMessage {
	base := Base{Type: DM}
	message := SMessage{Base: base, From: from, Message: msg}
	return message
}

func NewSGetConnected(ids []string) SGetConnected {
	base := Base{Type: GET_CONN}
	users := make([]GetConnUser, len(ids))
	for i, id := range ids {
		users[i] = GetConnUser{Id: id}
	}
	u := Users{ConnUsers: users}
	getConn := SGetConnected{Base: base, Users: u}
	return getConn
}

func NewErrorMessage(msg string) ErrorMessage {
	base := Base{Type: ERROR}
	message := ErrorMessage{Base: base, Message: msg}
	return message
}

func NewAddressMessage(addr int) AddressMessage {
	base := Base{Type: ADDRESS}
	message := AddressMessage{Base: base, Address: addr}
	return message
}

func NewLoginResponse(addr int) LoginResponse {
	base := Base{Type: LOGIN_RES}
	message := LoginResponse{Base: base, Address: addr}
	return message
}

func IsConfirmation(msg []byte) bool {
	// TODO This will handle the messages from the other user
	return string(msg) == "OK"
}
