package main

import (
        "net"
        "log"
        "fmt"
)

func main() {

        port := "127.0.0.1:1200"

        udpAddress, err := net.ResolveUDPAddr("udp4",port)

        if err != nil {
                log.Println("error resolving UDP address on ", port)
                log.Println(err)
                return
        }

        conn ,err := net.ListenUDP("udp",udpAddress)

        if err != nil {
                log.Println("error listening on UDP port ", port)
                log.Println(err)
                return
        }
        log.Println("Got a connection")
        defer conn.Close()

        read := listen(conn)
        for {

                message := <- read
                if message != nil {
                        fmt.Println("From main:", string(message))
                }
        }

}

// Each incoming connection will have a message with whatever they want to send
// and who sent it
type Message struct {
        Content []byte
        Sender *net.UDPAddr
}

func listen(conn *net.UDPConn) <-chan []byte {
        c := make(chan []byte)

        go func() {
                buff := make([]byte, 1024)

                for {
                        n, _, err := conn.ReadFromUDP(buff)
                        if n > 0 {
                                res := make([]byte, n)
                                copy(res, buff[:n])
                                c <- res
                        }
                        if err != nil {
                                c <- nil
                                break
                        }
                 }
        }()

        return c
}