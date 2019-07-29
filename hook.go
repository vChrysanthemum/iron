package iron

type HookBase struct {
	MatchPrefix   string
	ExcludePrefix []string
	Func          func(*Request) bool
}

type HookBeforeServeRequest struct {
	HookBase
}

type HookBeforeHttpHandle struct {
	HookBase
}

type HookErrorRecover func(*Request, interface{}) bool

type HookAfterHttpHandle struct {
	HookBase
}

type HookUrlRewrite func(*Request) bool

type Hook struct {
	BeforeServeRequest []HookBeforeServeRequest
	BeforeHttpHandles  []HookBeforeHttpHandle
	ErrorRecovers      []HookErrorRecover
	AfterHttpHandles   []HookAfterHttpHandle
	UrlRewrite         []HookUrlRewrite
}

func (p *Server) HookBeforeServeRequest(matchPrefix string, hookFunc func(*Request) bool, excludePrefix ...string) {
	p.Hook.BeforeServeRequest = append(p.Hook.BeforeServeRequest, HookBeforeServeRequest{
		HookBase{
			MatchPrefix:   matchPrefix,
			ExcludePrefix: excludePrefix,
			Func:          hookFunc,
		},
	})
}

func (p *Server) HookBeforeHttpHandle(matchPrefix string, hookFunc func(*Request) bool, excludePrefix ...string) {
	p.Hook.BeforeHttpHandles = append(p.Hook.BeforeHttpHandles, HookBeforeHttpHandle{
		HookBase{
			MatchPrefix:   matchPrefix,
			ExcludePrefix: excludePrefix,
			Func:          hookFunc,
		},
	})
}

func (p *Server) HookAfterHttpHandle(matchPrefix string, hookFunc func(*Request) bool, excludePrefix ...string) {
	p.Hook.AfterHttpHandles = append(p.Hook.AfterHttpHandles, HookAfterHttpHandle{
		HookBase{
			MatchPrefix:   matchPrefix,
			ExcludePrefix: excludePrefix,
			Func:          hookFunc,
		},
	})
}

func (mux *ServeMux) IsRequestURIMatchHookBase(ir *Request, hookBase *HookBase) bool {
	handleMatchPrefixLen := len(hookBase.MatchPrefix)
	reqURILen := len(ir.R.RequestURI)

	if len(hookBase.ExcludePrefix) > 0 {
		for _, prefix := range hookBase.ExcludePrefix {
			if reqURILen > len(prefix) && ir.R.RequestURI[:len(prefix)] == prefix {
				return false
			}
		}
	}

	if hookBase.MatchPrefix == "" ||
		(reqURILen >= handleMatchPrefixLen && ir.R.RequestURI[:handleMatchPrefixLen] == hookBase.MatchPrefix) {
		return true
	}

	return false
}
