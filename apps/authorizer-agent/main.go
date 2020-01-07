//
// authorizer-agent.go
//
// Copyright (c) 2019 Markku Rossi
//
// All rights reserved.
//

package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"time"

	"github.com/markkurossi/authorizer/api"
	"github.com/markkurossi/authorizer/secsh/agent"
)

var connections = make(map[string]*api.Client)

func main() {
	bindAddress := flag.String("a", "", "Unix-domain socket bind address")
	endpoint := flag.String("u", "", "Authorizer endpoint URL")
	benchmark := flag.Bool("b", false, "Benchmark server")
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

	if *benchmark {
		client, err := api.NewClient(*endpoint)
		if err != nil {
			log.Fatalf("api.NewClient: %s\n", err)
		}

		err = runBenchmark(client)
		client.Disconnect()

		if err != nil {
			log.Fatalf("runBenchmark: %s\n", err)
		}
		return
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
		req, err := agent.Read(conn)
		if err != nil {
			return err
		}
		log.Printf("<- %s\n", req)

		data, err := client.Call(req)
		if err != nil {
			return err
		}

		resp, err := agent.Wrap(data)
		if err != nil {
			return err
		}
		log.Printf("-> %s\n", resp)

		_, err = conn.Write(resp)
		if err != nil {
			return err
		}
	}
}

func runBenchmark(client *api.Client) error {
	log.Printf("Connecting to server\n")
	err := client.Connect()
	if err != nil {
		return err
	}
	defer client.Disconnect()

	data := []byte{0, 0, 0, 1, 255}

	log.Printf("Running benchmark\n")

	var min, max, total time.Duration

	iterations := 10

	for i := 0; i < iterations; i++ {
		start := time.Now()

		resp, err := client.Call(data)
		if err != nil {
			return err
		}
		if bytes.Compare(data, resp) != 0 {
			return fmt.Errorf("Invalid response data")
		}

		d := time.Now().Sub(start)

		total += d
		if i == 0 || d < min {
			min = d
		}
		if i == 0 || d > max {
			max = d
		}
	}

	fmt.Printf("%d iterations min/avg/max = %s/%s/%s\n", iterations,
		min, total/time.Duration(iterations), max)

	return nil
}
