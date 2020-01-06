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

	var get *http.Request

	for {
		fmt.Printf("Req %v\n", req)
		resp, err := client.http.Do(req)
		if err != nil {
			return nil, err
		}
		fmt.Printf("Resp %v\n", resp)
		switch resp.StatusCode {
		case http.StatusOK:
			data, err := ioutil.ReadAll(resp.Body)
			resp.Body.Close()
			if err != nil {
				return nil, err
			}
			return agent.Message(data), nil

		case http.StatusAccepted:
			resp.Body.Close()
			if get == nil {
				get, err = http.NewRequest("GET", client.url, nil)
				if err != nil {
					return nil, err
				}
			}
			req = get

		default:
			resp.Body.Close()
			return nil, fmt.Errorf("HTTP status %s", resp.Status)
		}
	}
}
