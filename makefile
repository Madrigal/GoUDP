PROJECT_ROOT := $(shell pwd)
GOPATH := $(PROJECT_ROOT)

all: server

server:
	go run GoUDP.go -s

client:
	go run GoUDP.go -port=127.0.0.1:1200

example:
	go run src/examples/weather.go