//
// authorizer-agent.go
//
// Copyright (c) 2019 Markku Rossi
//
// All rights reserved.
//

package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"path/filepath"

	"github.com/markkurossi/authorizer/api"
	"github.com/markkurossi/authorizer/secsh/agent"
)

func main() {
	dir, err := ioutil.TempDir("", "authorizer")
	if err != nil {
		log.Fatalf("TempDir: %s\n", err)
	}
	path := filepath.Join(dir, "agent.sock")

	listener, err := net.Listen("unix", path)
	if err != nil {
		log.Fatalf("Listen: %s\n", err)
	}
	defer listener.Close()

	fmt.Printf("SSH_AUTH_SOCK=%s\n", path)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatalf("Accept: %s\n", err)
		}
		go func() {
			err := handleConnection(conn)
			if err != nil && err != io.EOF {
				log.Printf("handleConnection: %s\n", err)
			}
		}()
	}
}

func handleConnection(conn net.Conn) error {
	client, err := api.NewClient()
	if err != nil {
		return err
	}

	for {
		msg, err := agent.Read(conn)
		if err != nil {
			return err
		}
		log.Printf("<- %s\n", msg)
		reply, err := client.Call(msg)
		if err != nil {
			return err
		}
		_, err = conn.Write(reply)
		if err != nil {
			return err
		}
	}
}
