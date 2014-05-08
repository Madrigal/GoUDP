package message

import (
	"encoding/xml"
)

const (
	FILETRANSFER_START = iota
	FILETRANSFER_MID
	FILETRANSFER_END
)

type FileMessage struct {
	XMLName xml.Name `xml:"Root"`
	Base
	Kind     int    `xml:"Kind"`
	To       string `xml:To`
	Filename string `xml:"Id"`
	Cont     string `xml:"Content"`
}

func NewFileStart(to string, filename string) FileMessage {
	base := Base{Type: FILE}
	f := FileMessage{Base: base, Kind: FILETRANSFER_START, To: to, Filename: filename}
	return f
}

func NewFileSend(to string, filename string, payload []byte) FileMessage {
	base := Base{Type: FILE}
	f := FileMessage{Base: base, Kind: FILETRANSFER_MID, To: to, Filename: filename, Cont: string(payload[:len(payload)])}
	return f
}

func NewFileEnd(to string, filename string) FileMessage {
	base := Base{Type: FILE}
	f := FileMessage{Base: base, Kind: FILETRANSFER_END, To: to, Filename: filename}
	return f
}

// TODO Need to add a different file for "from" and "to"
