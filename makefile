PROJECT_ROOT := $(shell pwd)
export GO_PROJECT_ROOT=$(PROJECT_ROOT)
VENDOR_PATH  := $(PROJECT_ROOT)/vendor
GOPATH := $(PROJECT_ROOT):$(VENDOR_PATH)
export $GOPATH

all: server

server:
	go run GoUDP.go -s -port=$(PORT)

client:
	go run GoUDP.go -port=$(PORT)

example:
	go run src/examples/weather.go

twitter:
	go run src/examples/twitter.go

election:
	go run src/examples/election.go