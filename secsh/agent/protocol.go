//
// protocol.go
//
// Copyright (c) 2019 Markku Rossi
//
// All rights reserved.
//

package agent

import (
	"encoding/binary"
	"fmt"
	"io"
)

type Type uint8

const (
	SSH_AGENTC_REQUEST_IDENTITIES            Type = 11
	SSH_AGENTC_SIGN_REQUEST                  Type = 13
	SSH_AGENTC_ADD_IDENTITY                  Type = 17
	SSH_AGENTC_REMOVE_IDENTITY               Type = 18
	SSH_AGENTC_REMOVE_ALL_IDENTITIES         Type = 19
	SSH_AGENTC_ADD_ID_CONSTRAINED            Type = 25
	SSH_AGENTC_ADD_SMARTCARD_KEY             Type = 20
	SSH_AGENTC_REMOVE_SMARTCARD_KEY          Type = 21
	SSH_AGENTC_LOCK                          Type = 22
	SSH_AGENTC_UNLOCK                        Type = 23
	SSH_AGENTC_ADD_SMARTCARD_KEY_CONSTRAINED Type = 26
	SSH_AGENTC_EXTENSION                     Type = 27

	SSH_AGENT_FAILURE           Type = 5
	SSH_AGENT_SUCCESS           Type = 6
	SSH_AGENT_EXTENSION_FAILURE Type = 28
	SSH_AGENT_IDENTITIES_ANSWER Type = 12
	SSH_AGENT_SIGN_RESPONSE     Type = 14
)

var types = map[Type]string{
	SSH_AGENTC_REQUEST_IDENTITIES:            "SSH_AGENTC_REQUEST_IDENTITIES",
	SSH_AGENTC_SIGN_REQUEST:                  "SSH_AGENTC_SIGN_REQUEST",
	SSH_AGENTC_ADD_IDENTITY:                  "SSH_AGENTC_ADD_IDENTITY",
	SSH_AGENTC_REMOVE_IDENTITY:               "SSH_AGENTC_REMOVE_IDENTITY",
	SSH_AGENTC_REMOVE_ALL_IDENTITIES:         "SSH_AGENTC_REMOVE_ALL_IDENTITIES",
	SSH_AGENTC_ADD_ID_CONSTRAINED:            "SSH_AGENTC_ADD_ID_CONSTRAINED",
	SSH_AGENTC_ADD_SMARTCARD_KEY:             "SSH_AGENTC_ADD_SMARTCARD_KEY",
	SSH_AGENTC_REMOVE_SMARTCARD_KEY:          "SSH_AGENTC_REMOVE_SMARTCARD_KEY",
	SSH_AGENTC_LOCK:                          "SSH_AGENTC_LOCK",
	SSH_AGENTC_UNLOCK:                        "SSH_AGENTC_UNLOCK",
	SSH_AGENTC_ADD_SMARTCARD_KEY_CONSTRAINED: "SSH_AGENTC_ADD_SMARTCARD_KEY_CONSTRAINED",
	SSH_AGENTC_EXTENSION:                     "SSH_AGENTC_EXTENSION",
	SSH_AGENT_FAILURE:                        "SSH_AGENT_FAILURE",
	SSH_AGENT_SUCCESS:                        "SSH_AGENT_SUCCESS",
	SSH_AGENT_EXTENSION_FAILURE:              "SSH_AGENT_EXTENSION_FAILURE",
	SSH_AGENT_IDENTITIES_ANSWER:              "SSH_AGENT_IDENTITIES_ANSWER",
	SSH_AGENT_SIGN_RESPONSE:                  "SSH_AGENT_SIGN_RESPONSE",
}

func (t Type) String() string {
	name, ok := types[t]
	if ok {
		return name
	}
	return fmt.Sprintf("{Type %d}", t)
}

var (
	bo = binary.BigEndian
)

type Message []byte

func (m Message) Type() Type {
	return Type(m[4])
}

func (m Message) Data() []byte {
	return m[5:]
}

func (m Message) String() string {
	t := m.Type()
	switch t {
	case SSH_AGENT_FAILURE, SSH_AGENT_SUCCESS:
		return t.String()

	default:
		if len(m) > 5 {
			return fmt.Sprintf("%s: %d bytes", t, len(m)-5)
		}
		return t.String()
	}
}

func Read(r io.Reader) (Message, error) {
	var buf [4]byte

	// Message length.
	_, err := io.ReadFull(r, buf[:])
	if err != nil {
		return nil, err
	}
	length := bo.Uint32(buf[:])
	if length < 1 || length > 65535 {
		return nil, fmt.Errorf("Invalid message length %d", length)
	}

	data := make([]byte, 4+length)
	copy(data, buf[:])

	_, err = io.ReadFull(r, data[4:])
	if err != nil {
		return nil, err
	}
	return Message(data), nil
}
