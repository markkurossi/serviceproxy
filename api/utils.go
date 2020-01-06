//
// utils.go
//
// Copyright (c) 2019 Markku Rossi
//
// All rights reserved.
//

package api

import (
	"fmt"
	"strings"
)

func canonizeEndpoint(endpoint string) string {
	if strings.HasSuffix(endpoint, "/") {
		return endpoint[0 : len(endpoint)-1]
	}
	return endpoint
}

func httpError(code int, data []byte) error {
	if len(data) == 0 {
		return fmt.Errorf("HTTP status %d", code)
	}
	return fmt.Errorf("%d: %s", code, string(data))
}
