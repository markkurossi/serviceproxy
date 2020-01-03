//
// client.go
//
// Copyright (c) 2019 Markku Rossi
//
// All rights reserved.
//

package api

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/markkurossi/authorizer/secsh/agent"
)

type Client struct {
	http *http.Client
	url  string
	id   string
}

func NewClient() (*Client, error) {
	var buf [16]byte

	_, err := rand.Read(buf[:])
	if err != nil {
		return nil, err
	}
	id := fmt.Sprintf("%x", buf[:])

	return &Client{
		http: new(http.Client),
		url:  "http://localhost:8080/client/" + id,
		id:   id,
	}, nil
}

func (client *Client) Call(msg agent.Message) (agent.Message, error) {
	req, err := http.NewRequest("POST", client.url, bytes.NewReader(msg))
	if err != nil {
		return nil, err
	}

	resp, err := client.http.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("HTTP Status %d", resp.StatusCode)
	}
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return agent.Message(data), nil
}
