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
	Kind int    `xml:"Kind"`
	Cont string `xml:"Content"`
}

func NewFileStart() FileMessage {
	base := Base{Type: FILE}
	f := FileMessage{Base: base, Kind: FILETRANSFER_START, Cont: ""}
	return f
}

func NewFileSend(payload []byte) FileMessage {
	base := Base{Type: FILE}
	f := FileMessage{Base: base, Kind: FILETRANSFER_MID, Cont: string(payload)}
	return f
}

func NewFileEnd() FileMessage {
	base := Base{Type: FILE}
	f := FileMessage{Base: base, Kind: FILETRANSFER_END, Cont: ""}
	return f
}
