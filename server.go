package iron

import (
	"net"
	"net/http"
	"net/http/fcgi"
	"regexp"
	"soloos/common/log"
	"time"

	"golang.org/x/net/http2"
)

type Server struct {
	NetListener         net.Listener
	httpServer          *http.Server
	httpMux             *ServeMux
	isClosedAfterHandle bool

	NotFoundHandler Handler
	Hook            Hook
	views           map[string]*View

	ImgExts []string

	Options Options
}

func (p *Server) Init(options Options) error {
	var err error

	err = p.loadOptions(options)
	if err != nil {
		return err
	}

	p.views = make(map[string]*View)
	p.httpMux = p.NewServeMux()

	p.ImgExts = []string{"jpeg", "gif", "png", "jpg"}

	initEncoder()

	p.NotFoundHandler = &HandleFunc{p, Handle404}
	p.Hook.BeforeServeRequest = make([]HookBeforeServeRequest, 0)
	p.Hook.BeforeHttpHandles = make([]HookBeforeHttpHandle, 0)
	p.Hook.ErrorRecovers = make([]HookErrorRecover, 0)
	p.Hook.AfterHttpHandles = make([]HookAfterHttpHandle, 0)
	p.Hook.UrlRewrite = make([]HookUrlRewrite, 0)

	p.HookBeforeServeRequest("", p.HookAccessWhiteListRequest)

	p.httpServer = &http.Server{
		Handler:        p.httpMux,
		ReadTimeout:    90 * time.Second,
		WriteTimeout:   90 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	http2.ConfigureServer(p.httpServer, &http2.Server{})

	return nil
}

func (p *Server) SetRequestSeter(reqSeter SetReuqestFunc) {
	p.httpMux.reqSeter = reqSeter
}

func (p *Server) SetRequestGeter(reqGeter GetReuqestFunc) {
	p.httpMux.reqGeter = reqGeter
}

func (p *Server) SetServeHTTPer(serveHTTPer ServeHTTPer) {
	p.httpMux.serveHTTPer = serveHTTPer
}

func (p *Server) SetCloseAfterHandle() {
	p.isClosedAfterHandle = true
}

func (p *Server) HookAccessWhiteListRequest(ir *Request) bool {
	if nil == p.Options.AccessWhiteList {
		return true
	}

	for _, value := range p.Options.AccessWhiteList {
		if isOk, _ := regexp.MatchString(value, ir.RemoteIp); true == isOk {
			return true
		}
	}

	Handle404(ir)
	return false
}

func (p *Server) Router(path string, handler func(*Request)) {
	p.httpMux.HandleFunc(path, handler)
}

func (p *Server) HandlerToServeHTTPFunc(handler func(*Request)) func(http.ResponseWriter, *http.Request) {
	h := &HandleFunc{p, handler}
	return h.ServeHTTP
}

func (p *Server) Serve() error {
	var err error
	log.Info("Server started, listen at:", p.Options.ListenStr)

	p.isClosedAfterHandle = false

	p.NetListener, err = net.Listen("tcp", p.Options.ListenStr)
	if nil != err {
		return err
	}

	switch p.Options.ServeType {
	case "fcgi":
		fcgi.Serve(p.NetListener, p.httpMux)

	case "server":
		p.httpServer.Addr = p.Options.ListenStr
		//err := p.httpServer.ListenAndServeTLS("./cert.pem", "./key.pem")
		err = p.httpServer.Serve(p.NetListener.(*net.TCPListener))
		if nil != err {
			return err
		}
	}

	log.Info("Server closed.")

	return nil

}

func (p *Server) Close() error {
	p.NetListener.Close()
	return nil
}
