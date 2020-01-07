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
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/markkurossi/authorizer"
)

type Client struct {
	http    *http.Client
	baseURL string
	url     string
	id      string
}

func NewClient(endpoint string) (*Client, error) {
	return &Client{
		http:    new(http.Client),
		baseURL: canonizeEndpoint(endpoint),
	}, nil
}

func (client *Client) ID() string {
	parts := strings.Split(client.url, "/")
	return parts[len(parts)-1]
}

func (client *Client) Connect() error {
	req, err := http.NewRequest("POST", client.baseURL+"/clients", nil)
	if err != nil {
		return err
	}

	resp, err := client.http.Do(req)
	if err != nil {
		return err
	}
	data, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return httpError(resp.StatusCode, data)
	}

	response := new(authorizer.ClientConnectResult)
	err = json.Unmarshal(data, response)
	if err != nil {
		return err
	}
	client.url = client.baseURL + response.URL
	client.id = response.ID

	return nil
}

func (client *Client) Disconnect() error {
	req, err := http.NewRequest("DELETE", client.url, nil)
	if err != nil {
		return err
	}

	resp, err := client.http.Do(req)
	if err != nil {
		return err
	}
	data, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return httpError(resp.StatusCode, data)
	}
	return nil
}

func (client *Client) Call(msg []byte) ([]byte, error) {
	envelope := &authorizer.Message{
		From: client.id,
	}
	envelope.SetBytes(msg)

	data, err := json.Marshal(envelope)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", client.url, bytes.NewReader(data))
	if err != nil {
		return nil, err
	}

	var get *http.Request

	for {
		resp, err := client.http.Do(req)
		if err != nil {
			return nil, err
		}
		data, err := ioutil.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return nil, err
		}

		switch resp.StatusCode {
		case http.StatusOK:
			env := new(authorizer.Message)
			err = json.Unmarshal(data, env)
			if err != nil {
				return nil, err
			}
			return env.Bytes()

		case http.StatusAccepted, http.StatusRequestTimeout:
			if get == nil {
				get, err = http.NewRequest("GET", client.url, nil)
				if err != nil {
					return nil, err
				}
			}
			req = get

		default:
			return nil, httpError(resp.StatusCode, data)
		}
	}
}
