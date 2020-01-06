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

	"github.com/markkurossi/authorizer/secsh/agent"
)

type Client struct {
	http    *http.Client
	baseURL string
	url     string
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

type ClientConnectResult struct {
	URL string `json:"url"`
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

	response := new(ClientConnectResult)
	err = json.Unmarshal(data, response)
	if err != nil {
		return err
	}
	client.url = client.baseURL + response.URL

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

func (client *Client) Call(msg agent.Message) (agent.Message, error) {
	req, err := http.NewRequest("POST", client.url, bytes.NewReader(msg))
	if err != nil {
		return nil, err
	}

	var get *http.Request

	for {
		//fmt.Printf("Req %v\n", req)
		resp, err := client.http.Do(req)
		if err != nil {
			return nil, err
		}
		//fmt.Printf("Resp %v\n", resp)
		data, err := ioutil.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return nil, err
		}

		switch resp.StatusCode {
		case http.StatusOK:
			return agent.Wrap(data)

		case http.StatusAccepted:
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
