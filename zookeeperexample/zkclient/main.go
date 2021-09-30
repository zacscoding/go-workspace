package main

import (
	"log"
	"time"
)

func main() {
	var (
		servers        = []string{"localhost:2181"}
		sessionTimeout = time.Minute
	)

	zkClient, err := NewZKClient(servers, sessionTimeout)
	if err != nil {
		log.Fatal(err)
	}
	defer zkClient.Close()
	log.Println(">> State::", zkClient.State())
}
