PROJECT_ROOT := $(shell pwd)
GOPATH := $(PROJECT_ROOT)

all: server

server:
	go run src/server.go
