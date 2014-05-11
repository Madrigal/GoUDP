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

## Client usage
This is based on IRC, so you will be familiar with most of the commands

someMessage
Broadcasts a message to all users. This is the option by default

/nick SomeNick
This changes your nickname. It is necessary at login

/names
Gives you the names of all connected users.

/msg Buddy Hello man
Says "Hello man" to the user with the nickname "Buddy"

/send Buddy file.jpg
Sends file "file.jpg" to the user with the nickname "Buddy"

/block Buddy
Blocks the user "Buddy" from sending messages to you

/twitter I like this day!
Updates your Twitter status with the message shown

/quit
Exits the chat

## Admin
/admin start
Starts a server on this instance

/admin stop
Stops the server on this instance