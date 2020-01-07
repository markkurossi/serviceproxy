//
// model.go
//
// Copyright (c) 2019 Markku Rossi
//
// All rights reserved.
//

package authorizer

import (
	"encoding/base64"
)

type ClientConnectResult struct {
	URL string `json:"url"`
	ID  string `json:"id"`
}

type ServerConnectResult struct {
	URL string `json:"url"`
}

type Message struct {
	From string `json:"from"`
	To   string `json:"to"`
	Data string `json:"data"`
}

func (m *Message) Bytes() ([]byte, error) {
	return base64.RawStdEncoding.DecodeString(m.Data)
}

func (m *Message) SetBytes(data []byte) {
	m.Data = base64.RawStdEncoding.EncodeToString(data)
}
