package iron

import "net/http"

type ServeHTTPer interface {
	ServeHTTP(http.ResponseWriter, *http.Request)
}

type SetReuqestFunc func(*http.Request, *Request) *http.Request
type GetReuqestFunc func(*http.Request) *Request
