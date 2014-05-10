PROJECT_ROOT := $(shell pwd)
export GO_PROJECT_ROOT=$(PROJECT_ROOT)
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
	go run src/examples/twitter.go