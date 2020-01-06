//
// client.go
//
// Copyright (c) 2019 Markku Rossi
//
// All rights reserved.
//

package api

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/markkurossi/authorizer/secsh/agent"
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

type ServerConnectResult struct {
	URL string `json:"url"`
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

func (server *Server) Receive() (agent.Message, error) {
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
			return agent.Wrap(data)

		case http.StatusNoContent:
			// Retry

		default:
			return nil, httpError(resp.StatusCode, data)
		}
	}
}
