package main

import (
	"fmt"
	"log"
	"net/url"
	"twitterWrapper"
)

func main() {
	fmt.Println("Hola tuiter")
	client, err := twitterWrapper.NewClient()
	if err != nil {
		log.Println("Twitter creation failed, reason", err)
	}
	fmt.Println(client)
	client.Update("Hola desde mi aplicacion", url.Values{})
}
