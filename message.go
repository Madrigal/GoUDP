package message

import (
    "encoding/xml"
)

const (
    LOGIN = "Login"
    BROAD = "Broadcast"
    DM = "DirectMessage"
    GET_CONN = "GetConnected"
    EXIT = "Exit"
)

type Base struct {
    XMLName xml.Name `xml:"Root"`
    Type string `xml:"Type"`
}

// Message sent from the client when he wants to login
type Login struct {
    Base
    Nickname string `xml:"Nickname"`
}

// Message a user sends to server. It covers both
// Broadcast and direct message
type UMessage struct {
    Base
    To string `xml:"To"`
    Message string `xml:"Message"`
}

// Message the server will sent to a user
type SMessage struct {
    Base
    From string `xml:"From"`
    Message string `xml:"Message"`
}

// When the user request connected users, he will
// only specify that as type. Hence no need for more fields
type UGetConnected struct {
    Base
}

type SGetConnected struct {
    Base
    Users []GetConnUser `xml:"users"`
}

type GetConnUser struct {
    Id string `xml:"user"`
}

// Since user will only specify exit no extra fields are needed
type UExit struct {
    Base
}

///// Client calls
func newLogin(nickname string) Login {
    base := Base{Type: LOGIN}
    login := Login{base, nickname}
    return login
}

func newBroadcast(msg string) UMessage {
    base := Base{Type: BROAD}
    message := UMessage{base, "", msg}
    return message
}

func newDirectMessage(to string, msg string) UMessage {
    base := Base{Type: DM}
    message := UMessage{base, to, msg}
    return message
}

func newGetConnected() UGetConnected {
    base := Base{Type: GET_CONN}
    getConn := SGetConnected{base}
    return getConn
}

func newExit() UExit {
    base := Base{Type: EXIT}
    exit := UExit{base}
    return exit
}

///// Server calls
func newSBroadcast (from string, msg string) SMessage {
    base := Base{Type: BROAD}
    message := SMessage{base, from, msg}
    return message
}

func newSDirectMessage (from string, msg string) SMessage {
    base := Base{Type: DM}
    message := SMessage{base, from, msg}
    return message
}

func newGetConnected(ids []string) SGetConnected {
    base := Base{Type: GET_CONN}
    users []GetConnUser
    for _, id := range ids {
        users.append(GetConnUser{id})
    }
    getConn := SGetConnected{base, users}
    return getConn
}