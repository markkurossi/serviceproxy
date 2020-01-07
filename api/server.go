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
)

type Server struct {
	http    *http.Client
	baseURL string
	url     string
}

func NewServer(endpoint string) (*Server, error) {
	return &Server{
		http:    new(http.Client),
		baseURL: canonizeEndpoint(endpoint),
	}, nil
}

func (server *Server) Connect() error {
	req, err := http.NewRequest("POST", server.baseURL+"/agents", nil)
	if err != nil {
		return err
	}

	resp, err := server.http.Do(req)
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

	response := new(ServerConnectResult)
	err = json.Unmarshal(data, response)
	if err != nil {
		return err
	}
	server.url = server.baseURL + response.URL

	return nil
}

func (server *Server) Receive() (*Message, error) {
	req, err := http.NewRequest("GET", server.url, nil)
	if err != nil {
		return nil, err
	}

	for {
		resp, err := server.http.Do(req)
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
			msg := new(Message)
			err = json.Unmarshal(data, msg)
			if err != nil {
				return nil, err
			}
			return msg, nil

		case http.StatusNoContent:
			// Retry

		default:
			return nil, httpError(resp.StatusCode, data)
		}
	}
}

func (server *Server) Send(msg *Message) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	req, err := http.NewRequest("POST", server.url, bytes.NewReader(data))
	if err != nil {
		return err
	}

	resp, err := server.http.Do(req)
	if err != nil {
		return err
	}
	data, err = ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return httpError(resp.StatusCode, data)
	}

	return nil
}
