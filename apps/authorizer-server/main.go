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
		req, err := server.Receive()
		if err != nil {
			fmt.Printf("Receive error: %s\n", err)
			// XXX server.Disconnect
			os.Exit(1)
		}
		log.Printf("<- %s\n", req)
	}
}
