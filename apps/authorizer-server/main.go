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
	"os"

	"github.com/markkurossi/authorizer/api"
	"github.com/markkurossi/authorizer/secsh/agent"
)

func main() {
	endpoint := flag.String("u", "", "Authorizer endpoint URL")
	flag.Parse()

	if len(*endpoint) == 0 {
		fmt.Printf("No authorizer URL specified\n")
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

		if payload.Type() == 255 {
			msg.To = msg.From
		} else {
			fmt.Printf("Agent operations not implemented yet\n")
			os.Exit(1)
		}

		err = server.Send(msg)
		if err != nil {
			fmt.Printf("Send error: %s\n", err)
			os.Exit(1)
		}
	}
}
