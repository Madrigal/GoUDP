PROJECT_ROOT := $(shell pwd)
GOPATH := $(PROJECT_ROOT)

all: server

server:
	go run src/server.go

client:
	go run src/client.go 127.0.0.1:1200

example:
	go run src/examples/weather.go