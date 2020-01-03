/*
 * fn.go
 */

package authorizer

import (
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
	mux.HandleFunc("/agent", Agent)
	mux.HandleFunc("/client/", Client)
}

func Authorizer(w http.ResponseWriter, r *http.Request) {
	mux.ServeHTTP(w, r)
}

func Error500f(w http.ResponseWriter, format string, a ...interface{}) {
	http.Error(w, fmt.Sprintf(format, a...), http.StatusInternalServerError)
}
