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
	ERROR    = "Error"
	BLOCK    = "Block"
	FILE     = "FILE"
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
	STARTFILE_T Type = iota
	SENDFILE_T  Type = iota
	ENDFILE_T   Type = iota
	EXIT_T      Type = iota
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

type UserPackage struct {
	Block         *Block
	Login         *Login
	UMessage      *UMessage
	UGetConnected *UGetConnected
	UExit         *UExit
	File          *FileMessage
}

type ServerPackage struct {
	Direct    *SMessage
	Connected *SGetConnected
	Block     *Block
	Error     *ErrorMessage
	File      *FileMessage
}

func DecodeServerMessage(msg []byte) (Type, *ServerPackage, error) {
	var m ServerMessage
	err := xml.Unmarshal(msg, &m)
	if err != nil {
		return UNKNOWN_T, nil, err
	}
	switch m.Type {
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

		var t Type

		if f.Kind == FILETRANSFER_START {
			t = STARTFILE_T
		}
		if f.Kind == FILETRANSFER_MID {
			t = SENDFILE_T
		}
		if f.Kind == FILETRANSFER_END {
			t = ENDFILE_T
		}

		return t, &mp, nil

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

		var t Type

		if f.Kind == FILETRANSFER_START {
			t = STARTFILE_T
		}
		if f.Kind == FILETRANSFER_MID {
			t = SENDFILE_T
		}
		if f.Kind == FILETRANSFER_END {
			t = ENDFILE_T
		}

		return t, &up, nil

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

func IsConfirmation(msg []byte) bool {
	// TODO This will handle the messages from the other user
	return string(msg) == "OK"
}
