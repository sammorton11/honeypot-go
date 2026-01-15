package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"sync"
)

type Attempt struct {
	Address string `json:"address"`
	Network string `json:"network"`
	Message string `json:"message"`
}

var m sync.Mutex

func main() {
	// List
	/* c, err := net.ListenIP("ip4:tcp", &net.IPAddr{IP: net.ParseIP("0.0.0.0")}) // fake SSH port
	if err != nil {
		log.Fatal(err)
	} */

	listen, err := net.Listen("tcp", ":2222")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("honeypot listening on :2222")
	fmt.Println("try: nc localhost 2222")

	/* 	_lookUp := make(map[string][]byte) */
	lookUp := make(map[string]Attempt)

	for {
		c, err := listen.Accept()
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
			for {
				// read data from the connection
				n, err := c.Read(buf)
				if err != nil {
					break
				}
				attempt := Attempt{
					Network: userNetwork,
					Address: userAddr,
					Message: string(buf[:n]),
				}

/* 				hash := rand.Text() */

				// buffer has data from client -- add to lookup table
				// Or do i .. push to a channel?
				// Replace this with an http request to localhost:8080/attempt
				/* m.Lock()
				lookUp[hash] = attempt
				m.Unlock() */

				jsonData, err := json.Marshal(attempt)
				if err != nil {
					fmt.Println("Error marshaling attempt to rest server", err)
				}

				req, err := http.Post("http://localhost:8080/attempt", "application/json", bytes.NewBuffer(jsonData))
				if err != nil {
					fmt.Println("Error sending attempt to rest server", err)
					continue
				}
				defer req.Body.Close()

				fmt.Printf("Network: %s\n", attempt.Network)
				fmt.Printf("Address: %s\n", attempt.Address)
				fmt.Printf("Buffer: %v\n", attempt.Message)
				for v := range lookUp {
					fmt.Printf("Attempt: %v\n", v)
				}
			}
			c.Close()
		}(c)
	}
}
