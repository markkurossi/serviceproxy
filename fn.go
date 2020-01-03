/*
 * fn.go
 */

package authorizer

import (
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
	mux.HandleFunc("/agent", Agent)
	mux.HandleFunc("/client", Client)
}

func Authorizer(w http.ResponseWriter, r *http.Request) {
	mux.ServeHTTP(w, r)
}
