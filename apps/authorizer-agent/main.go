//
// authorizer-agent.go
//
// Copyright (c) 2019 Markku Rossi
//
// All rights reserved.
//

package main

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"os"
	"os/signal"
	"path/filepath"

	"github.com/markkurossi/authorizer/api"
	"github.com/markkurossi/authorizer/secsh/agent"
)

var connections = make(map[string]*api.Client)

func main() {
	bindAddress := flag.String("a", "", "Unix-domain socket bind address")
	endpoint := flag.String("u", "", "Authorizer endpoint URL")
	flag.Parse()

	if len(*bindAddress) == 0 {
		dir, err := ioutil.TempDir("", "authorizer")
		if err != nil {
			log.Fatalf("TempDir: %s\n", err)
		}
		*bindAddress = filepath.Join(dir, "agent.sock")
	}
	if len(*endpoint) == 0 {
		fmt.Printf("No authorizer URL specified\n")
		os.Exit(1)
	}

	os.RemoveAll(*bindAddress)

	listener, err := net.Listen("unix", *bindAddress)
	if err != nil {
		log.Fatalf("Listen: %s\n", err)
	}
	defer listener.Close()

	fmt.Printf("SSH_AUTH_SOCK=%s\n", *bindAddress)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		s := <-c
		fmt.Println("signal", s)
		for k, c := range connections {
			fmt.Printf("%s...", k)
			err = c.Disconnect()
			if err != nil {
				fmt.Printf("%s\n", err)
			} else {
				fmt.Println()
			}
		}
		os.Exit(0)
	}()

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatalf("Accept: %s\n", err)
		}
		log.Printf("New connections\n")
		go func(c net.Conn) {
			err := handleConnection(c, *endpoint)
			if err != nil && err != io.EOF {
				log.Printf("Connection error: %s\n", err)
			}
		}(conn)
	}
}

func handleConnection(conn net.Conn, url string) error {
	client, err := api.NewClient(url)
	if err != nil {
		return err
	}

	log.Printf("Connecting to server\n")
	err = client.Connect()
	if err != nil {
		return err
	}
	connections[client.ID()] = client
	defer func() {
		delete(connections, client.ID())
		err = client.Disconnect()
		if err != nil {
			fmt.Printf("%s: %s\n", client.ID(), err)
		}
	}()

	log.Printf("Processing messages\n")
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
