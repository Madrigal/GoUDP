# GoUDP - An UDP chat server written in Go

This is a school project that I'm working on. It is basically a chat server with the following features
- It communicates via XML. The following messages are allowed:
-- Broadcast to every registered user
-- Request to get all connected users
-- Send a private message
-- Exit the chat
- High availability: If the server goes down any client can take the role of the server
- The clients' clocks need to be synchronized
- The client needs to show weather information
- A client can send files to another client
- Offline messages
- Block users
- Update FB status


## Usage
``` Make ``` runs the server
``` Make client ``` runs a process as a client