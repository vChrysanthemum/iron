package iron

import (
	"log"
	"net/http"
	"os"
	"path"
	"runtime/debug"
	"strings"
	"sync"
)

type ServeMux struct {
	server      *Server
	mu          sync.RWMutex
	m           map[string]muxEntry
	hosts       bool // whether any patterns contain hostnames
	serveHTTPer ServeHTTPer
	reqSeter    SetReuqestFunc
	reqGeter    GetReuqestFunc
}

// NewServeMux allocates and returns a new ServeMux.
func (p *Server) NewServeMux() *ServeMux {
	ret := &ServeMux{server: p, m: make(map[string]muxEntry)}
	return ret
}

func (mux *ServeMux) Handle(pattern string, handler Handler) {
	mux.mu.Lock()
	defer mux.mu.Unlock()

	if pattern == "" {
		panic("http: invalid pattern " + pattern)
	}
	if handler == nil {
		panic("http: nil handler")
	}
	if mux.m[pattern].explicit {
		panic("http: multiple registrations for " + pattern)
	}

	mux.m[pattern] = muxEntry{explicit: true, h: handler, pattern: pattern}

	if pattern[0] != '/' {
		mux.hosts = true
	}

	// Helpful behavior:
	// If pattern is /tree/, insert an implicit permanent redirect for /tree.
	// It can be overridden by an explicit registration.
	n := len(pattern)
	if n > 0 && pattern[n-1] == '/' && !mux.m[pattern[0:n-1]].explicit {
		// If pattern contains a host name, strip it and use remaining
		// path for redirect.
		path := pattern
		if pattern[0] != '/' {
			// In pattern, at least the last character is a '/', so
			// strings.Index can't be -1.
			path = pattern[strings.Index(pattern, "/"):]
		}
		mux.m[pattern[0:n-1]] = muxEntry{h: RedirectHandler(path, http.StatusMovedPermanently), pattern: pattern}
	}
}

// Does path match pattern?
func pathMatch(pattern, path string) bool {
	if len(pattern) == 0 {
		// should not happen
		return false
	}

	n := len(pattern)

	if pattern == "*" {
		return true
	}

	if pattern[n-1] == '*' {
		if strings.HasPrefix(path, pattern[:n-1]) {
			return true
		}
	}

	if pattern[n-1] != '/' {
		return pattern == path
	}
	return len(path) >= n && path[0:n] == pattern
}

// Return the canonical path for p, eliminating . and .. elements.
func cleanPath(p string) string {
	if p == "" {
		return "/"
	}
	if p[0] != '/' {
		p = "/" + p
	}
	np := path.Clean(p)
	// path.Clean removes trailing slash except for root;
	// put the trailing slash back if necessary.
	if p[len(p)-1] == '/' && np != "/" {
		np += "/"
	}
	return np
}

// Find a handler on a handler map given a path string
// Most-specific (longest) pattern wins
func (mux *ServeMux) match(path string) (h Handler, pattern string) {
	var n = 0
	for k, v := range mux.m {
		if !pathMatch(k, path) {
			continue
		}
		if h == nil || len(k) > n {
			n = len(k)
			h = v.h
			pattern = v.pattern
		}
	}
	return
}

func (mux *ServeMux) handler(host, path string) (h Handler, pattern string) {
	mux.mu.RLock()
	defer mux.mu.RUnlock()

	// Host-specific pattern takes precedence over generic ones
	if mux.hosts {
		h, pattern = mux.match(host + path)
	}
	if h == nil {
		h, pattern = mux.match(path)
	}
	if h == nil {
		h, pattern = mux.server.NotFoundHandler, ""
	}
	return
}

func (mux *ServeMux) Handler(r *http.Request) (h Handler, pattern string) {
	if r.Method != "CONNECT" {
		if p := cleanPath(r.URL.Path); p != r.URL.Path {
			_, pattern = mux.handler(r.Host, p)
			url := *r.URL
			url.Path = p
			return RedirectHandler(url.String(), http.StatusMovedPermanently), pattern
		}
	}

	return mux.handler(r.Host, r.URL.Path)
}

func (mux *ServeMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.RequestURI == "*" {
		if r.ProtoAtLeast(1, 1) {
			w.Header().Set("Connection", "close")
		}
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var ok bool
	var ir *Request = &Request{server: mux.server}
	if mux.reqSeter != nil {
		r = mux.reqSeter(r, ir)
	}
	ir.Init(w, r)

	for _, h := range mux.server.Hook.BeforeServeRequest {
		if mux.IsRequestURIMatchHookBase(ir, &h.HookBase) {
			if !h.Func(ir) {
				return
			}
		}
	}

	for _, h := range mux.server.Hook.UrlRewrite {
		if !h(ir) {
			return
		}
	}

	defer func() {
		err := recover()
		if nil == err {
			return
		}
		log.Println(string(debug.Stack()))
		log.Println(err)
		for _, h := range mux.server.Hook.ErrorRecovers {
			h(ir, err)
		}

		for _, h := range mux.server.Hook.AfterHttpHandles {
			if mux.IsRequestURIMatchHookBase(ir, &h.HookBase) {
				if !h.Func(ir) {
					break
				}
			}
		}

		if mux.server.isClosedAfterHandle {
			log.Println("Server closed.")
			os.Exit(0)
		}
	}()

	if mux.serveHTTPer != nil {
		mux.serveHTTPer.ServeHTTP(w, r)

	} else {
		_, ok = mux.server.httpMux.m[ir.R.URL.Path]
		if ok {
			h, _ := mux.Handler(r)
			h.TinyironServeHTTP(ir)
			return
		}

		if strings.HasPrefix(ir.R.URL.Path, "/static") {
			file := path.Join(mux.server.Options.SiteStaticBasePath, ir.R.URL.Path[8:])
			if FileExists(file) {
				http.ServeFile(ir.W, ir.R, file)
				return
			}
		}

		h, _ := mux.Handler(r)
		h.TinyironServeHTTP(ir)
		return
		// w.WriteHeader(http.StatusNotFound)
		// Handle404(ir)
		// return
	}
}

// HandleFunc registers the handler function for the given pattern.
func (mux *ServeMux) HandleFunc(pattern string, handler func(*Request)) {
	h := &HandleFunc{mux.server, handler}
	mux.Handle(pattern, h)
}
