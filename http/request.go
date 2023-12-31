package http

import "net/http"

type Request struct {
	*http.Request
	Params         map[string]string
	RequestID      string
	RestPath       string
	RawQueryParams string
}
