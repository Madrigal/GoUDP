package message

import (
	"encoding/xml"
	"errors"
)

const (
	LOGIN    = "Login"
	BROAD    = "Broadcast"
	DM       = "DirectMessage"
	GET_CONN = "GetConnected"
	EXIT     = "Exit"
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

// This type will decode an incoming message
// The UserPackage will hold the actual values
type UserMessage struct {
	Base
	Data interface{}
}

type UserPackage struct {
	Login         *Login
	UMessage      *UMessage
	UGetConnected *UGetConnected
	UExit         *UExit
}

// temp function
func DecodeUserMessage(msg []byte) (*UserPackage, error) {
	var m UserMessage
	err := xml.Unmarshal(msg, &m)
	if err != nil {
		return nil, err
	}
	switch m.Type {
	case LOGIN:
		var l Login
		err := xml.Unmarshal(msg, &l)
		if err != nil {
			return nil, errors.New("Couldn't decode the message: Login malformed")
		}
		up := UserPackage{
			Login:         &l,
			UMessage:      nil,
			UGetConnected: nil,
			UExit:         nil,
		}
		return &up, nil
	case BROAD, DM:
		var b UMessage
		err := xml.Unmarshal(msg, &b)
		if err != nil {
			return nil, errors.New("Couldn't decode the message: Login malformed")
		}
		up := UserPackage{
			Login:         nil,
			UMessage:      &b,
			UGetConnected: nil,
			UExit:         nil,
		}
		return &up, nil
	case GET_CONN:
		var u UGetConnected
		err := xml.Unmarshal(msg, &u)
		if err != nil {
			return nil, errors.New("Couldn't decode the message: Login malformed")
		}
		up := UserPackage{
			Login:         nil,
			UMessage:      nil,
			UGetConnected: &u,
			UExit:         nil,
		}
		return &up, nil
	case EXIT:
		var u UExit
		err := xml.Unmarshal(msg, &u)
		if err != nil {
			return nil, errors.New("Couldn't decode the message: Login malformed")
		}
		up := UserPackage{
			Login:         nil,
			UMessage:      nil,
			UGetConnected: nil,
			UExit:         &u,
		}
		return &up, nil
	default:
		return nil, errors.New("Couldn't decode the message: No matching type")
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
