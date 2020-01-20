//
// fn.go
//
// Copyright (c) 2020 Markku Rossi
//
// All rights reserved.
//

package authorizer

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"os"

	"github.com/markkurossi/cloudsdk/api/auth"
	"github.com/markkurossi/go-libs/fn"
)

const (
	REALM            = "Service Proxy"
	TOPIC_AUTHORIZER = "Authorizer"
	SUB_REQUESTS     = "Requests"
	ATTR_RESPONSE    = "response"
)

var (
	mux        *http.ServeMux
	projectID  string
	store      *auth.ClientStore
	authPubkey ed25519.PublicKey
)

func Fatalf(format string, a ...interface{}) {
	fmt.Printf(format, a...)
	os.Exit(1)
}

func init() {
	mux = http.NewServeMux()
	mux.HandleFunc("/agents", Agents)
	mux.HandleFunc("/agents/", Agent)
	mux.HandleFunc("/clients", Clients)
	mux.HandleFunc("/clients/", Client)

	id, err := fn.GetProjectID()
	if err != nil {
		Fatalf("GetProjectID: %s\n", err)
	}
	projectID = id

	store, err = auth.NewClientStore()
	if err != nil {
		Fatalf("NewClientStore: %s\n", err)
	}

	assets, err := store.Asset(auth.ASSET_AUTH_PUBKEY)
	if err != nil {
		Fatalf("store.Asset(%s)\n", auth.ASSET_AUTH_PUBKEY)
	}
	if len(assets) == 0 {
		Fatalf("No auth public key\n")
	}
	authPubkey = ed25519.PublicKey(assets[0].Data)
}

func ServiceProxy(w http.ResponseWriter, r *http.Request) {
	mux.ServeHTTP(w, r)
}

func tokenVerifier(message, sig []byte) bool {
	return ed25519.Verify(authPubkey, message, sig)
}

func Errorf(w http.ResponseWriter, code int, format string, a ...interface{}) {
	http.Error(w, fmt.Sprintf(format, a...), code)
}

func Error500f(w http.ResponseWriter, format string, a ...interface{}) {
	Errorf(w, http.StatusInternalServerError, format, a...)
}

type ID []byte

func (id ID) String() string {
	return fmt.Sprintf("%x", []byte(id))
}

func (id ID) Topic() string {
	return fmt.Sprintf("t%x", []byte(id))
}

func (id ID) Subscription() string {
	return fmt.Sprintf("s%x", []byte(id))
}

func NewID() (ID, error) {
	var buf [16]byte

	_, err := rand.Read(buf[:])
	if err != nil {
		return nil, err
	}
	return buf[:], nil
}

func ParseID(str string) (ID, error) {
	data, err := hex.DecodeString(str)
	if err != nil {
		return nil, err
	}
	if len(data) < 4 {
		return nil, fmt.Errorf("Truncated ID '%s'", str)
	}
	return data, nil
}
