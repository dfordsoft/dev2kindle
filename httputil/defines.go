package httputil

import "net/http"

var (
	client           *http.Client
	noRedirectClient *http.Client
)
