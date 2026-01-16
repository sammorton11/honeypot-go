package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
)

type Attempt struct {
	Address string `json:"address"`
	Network string `json:"network"`
	Message string `json:"message"`
}

func main() {
	listener, err := net.Listen("tcp", ":2222")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("honeypot listening on :2222")
	fmt.Println("try: nc localhost 2222")

	for {
		c, err := listener.Accept()
		if err != nil {
			log.Fatal(err)
		}

		// is this the servers addrs or is the sender?
		userAddr := c.RemoteAddr().String()
		userNetwork := c.RemoteAddr().Network()

		// fake SSH banner to bait them
		c.Write([]byte("SSH-2.0-OpenSSH_7.4\r\n"))

		// launch connections in a go routine -- so we can have more than one connection
		go func(c net.Conn) {
			// read whatever they send (login attempts, etc)
			buf := make([]byte, 1024)

			// Infinite loop
			for {
				// read data from the connection
				n, err := c.Read(buf)
				if err != nil {
					break
				}

				// Package data into an object
				attempt := Attempt{
					Network: userNetwork,
					Address: userAddr,
					Message: string(buf[:n]),
				}

				// Serialize 
				jsonData, err := json.Marshal(attempt)
				if err != nil {
					fmt.Println("Error marshaling attempt to rest server", err)
				}

				// Send over the wire to proxy server to store in db
				req, err := http.Post("http://localhost:8080/attempt", "application/json", bytes.NewBuffer(jsonData))
				if err != nil {
					fmt.Println("Error sending attempt to rest server", err)
					continue
				}
				defer req.Body.Close()

				// Log info
				fmt.Printf("Network: %s\n", attempt.Network)
				fmt.Printf("Address: %s\n", attempt.Address)
				fmt.Printf("Buffer: %v\n", attempt.Message)
			}
			// Close that damn connection
			c.Close()
		}(c)
	}
}
