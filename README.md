# GoUDP - An UDP chat server written in Go

This is a school project that I worked on. Besides what's on the title,
what is so special about this is that each client can become a server
when it detects that the server is down, and it can do it via UDP.
Sadly, what makes it interesting is probably also what makes it not
suitable for real life applications.

## Features
- It communicates via XML. The following messages are allowed:
-- Broadcast to every registered user
-- Request to get all connected users
-- Send a private message
-- Exit the chat
- High availability: If the server goes down any client can take the role of the server
- The clients' clocks are synchronized via the [Berkeley algorithm](http://en.wikipedia.org/wiki/Berkeley_algorithm)
- The client needs to show weather information. This is done via [Open weather map](http://openweathermap.org)
- A client can send files to another client
- They can also send offline messages that the recipient will get as soon as he reconnects
- Block users
- Update Twitter status thanks to [Xiam's library](https://github.com/xiam/twitter)


## Usage
``` Make ``` runs the server
``` Make client ``` runs a process as a client

## Client usage
This is inspired by IRC, so you will be familiar with most of the commands

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

## Dependencies
https://github.com/xiam/twitter
https://github.com/gosexy/yaml