/*
 * fn.go
 */

package authorizer

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
)

const (
	TOPIC_AUTHORIZER = "Authorizer"
	SUB_REQUESTS     = "Requests"
)

var (
	mux *http.ServeMux
)

func init() {
	mux = http.NewServeMux()
	mux.HandleFunc("/agents", Agents)
	mux.HandleFunc("/agents/", Agent)
	mux.HandleFunc("/clients", Clients)
	mux.HandleFunc("/clients/", Client)
}

func Authorizer(w http.ResponseWriter, r *http.Request) {
	mux.ServeHTTP(w, r)
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
	return hex.DecodeString(str)
}
