//
// authorizer-server.go
//
// Copyright (c) 2019 Markku Rossi
//
// All rights reserved.
//

package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"

	"github.com/markkurossi/authorizer/api"
	"github.com/markkurossi/authorizer/secsh/agent"
)

func main() {
	endpoint := flag.String("u", "", "Authorizer endpoint URL")
	sock := flag.String("a", "", "SSH Agent endpoint (default $SSH_AUTH_SOCK)")
	flag.Parse()

	if len(*endpoint) == 0 {
		fmt.Printf("No authorizer URL specified\n")
		os.Exit(1)
	}
	if len(*sock) == 0 {
		path := os.Getenv("SSH_AUTH_SOCK")
		if len(path) == 0 {
			fmt.Printf("No -a specified and SSH_AUTH_SOCK is unset\n")
			os.Exit(1)
		}
		*sock = path
	}

	conn, err := net.Dial("unix", *sock)
	if err != nil {
		fmt.Printf("Could not connect to agent '%s': %s\n", *sock, err)
		os.Exit(1)
	}

	server, err := api.NewServer(*endpoint)
	if err != nil {
		fmt.Printf("Failed to create API client: %s\n", err)
		os.Exit(1)
	}

	err = server.Connect()
	if err != nil {
		fmt.Printf("Failed to connecto to server: %s\n", err)
		os.Exit(1)
	}

	for {
		msg, err := server.Receive()
		if err != nil {
			fmt.Printf("Receive error: %s\n", err)
			// XXX server.Disconnect
			os.Exit(1)
		}
		data, err := msg.Bytes()
		if err != nil {
			fmt.Printf("Invalid message: %v\n", msg)
			continue
		}
		payload, err := agent.Wrap(data)
		if err != nil {
			fmt.Printf("Invalid SSH agent message: %v\n", err)
			continue
		}
		log.Printf("%s <- %s\n", msg.From, payload)

		msg.To = msg.From

		if payload.Type() != 255 { // 255 is ping for benchmark
			_, err = conn.Write(payload)
			if err != nil {
				fmt.Printf("Agent write failed: %s\n", err)
				os.Exit(1)
			}
			resp, err := agent.Read(conn)
			if err != nil {
				fmt.Printf("Agent read failed: %s\n", err)
				os.Exit(1)
			}
			msg.SetBytes(resp)

			log.Printf("%s -> %s\n", msg.To, resp)
		}

		err = server.Send(msg)
		if err != nil {
			fmt.Printf("Send error: %s\n", err)
			os.Exit(1)
		}
	}
}
