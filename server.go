package main

import (
        "net"
        "log"
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

        var buf []byte = make([]byte, 1500)

        for {
                n,address, err := conn.ReadFromUDP(buf)

                if err != nil {
                        log.Println("error reading data from connection")
                        log.Println(err)
                        return
                }

                if address != nil {

                        log.Println("got message from ", address, " with n = ", n)

                        if n > 0 {
                                log.Println("from address", address, "got message:", string(buf[0:n]), n)
                        }

                        // Write ack to the client
                        _, err := conn.WriteTo([]byte("Todo OK"), address)
                        if err != nil {
                                log.Println("error writing data to client")
                                log.Println(err)
                        }
                        log.Println("Wrote to client")
                }
        }

}