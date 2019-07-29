package iron

import (
	"net/http"
)

type muxEntry struct {
	explicit bool
	h        Handler
	pattern  string
}

// Redirect to a fixed URL
type redirectHandler struct {
	url  string
	code int
}

func (rh *redirectHandler) TinyironServeHTTP(ir *Request) {
	http.Redirect(ir.W, ir.R, rh.url, rh.code)
}

// RedirectHandler returns a req handler that redirects
// each req it receives to the given url using the given
// status code.
//
// The provided code should be in the 3xx range and is usually
// StatusMovedPermanently, StatusFound or StatusSeeOther.
func RedirectHandler(url string, code int) Handler {
	return &redirectHandler{url, code}
}

func Handle404(ir *Request) {
	ir.ApiOutput(nil, -1, "command not found")
}

type Handler interface {
	TinyironServeHTTP(*Request)
}

type HandleFunc struct {
	Server *Server
	Handle func(ir *Request)
}

// TinyironServeHTTP calls f(w, r).
func (f *HandleFunc) TinyironServeHTTP(ir *Request) {
	for _, h := range f.Server.Hook.BeforeHttpHandles {
		if f.Server.httpMux.IsRequestURIMatchHookBase(ir, &h.HookBase) {
			if !h.Func(ir) {
				return
			}
		}
	}
	f.Handle(ir)
}

func (f *HandleFunc) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ir := f.Server.httpMux.reqGeter(r)
	ir.W = w
	ir.R = r
	for _, h := range f.Server.Hook.BeforeHttpHandles {
		if f.Server.httpMux.IsRequestURIMatchHookBase(ir, &h.HookBase) {
			if !h.Func(ir) {
				return
			}
		}
	}
	f.Handle(ir)
}
