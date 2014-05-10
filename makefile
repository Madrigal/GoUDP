PROJECT_ROOT := $(shell pwd)
VENDOR_PATH  := $(PROJECT_ROOT)/vendor
GOPATH := $(PROJECT_ROOT):$(VENDOR_PATH)

all: server

server:
	go run GoUDP.go -s

client:
	go run GoUDP.go -port=127.0.0.1:1200

example:
	go run src/examples/weather.go

twitter:
	go run deleteme.go