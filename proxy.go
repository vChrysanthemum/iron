package iron

import (
	"reflect"
	"strings"

	"golang.org/x/xerrors"
)

type ProxyService struct {
	FunctionName        string
	Function            reflect.Value
	Params              []reflect.Type
	IsHasReqeustContext bool
	IsHasEasyKvReqArgs  bool
	IsHasUrlKvReqArgs   bool
}

type ProxyBeforeServiceHook func(path string,
	reqCtx RequestContext, resp IResponse, reqArgs ...LowReqArgs) (ret IResponse, isContinue bool)
type ProxyAfterServiceHook func(path string,
	resp IResponse,
	reqCtx RequestContext, reqArgs ...LowReqArgs) IResponse

type IProxy interface {
	RegisterService(path string, service interface{})
	HookBeforeService(path string, hook ProxyBeforeServiceHook)
	HookAfterService(path string, hook ProxyAfterServiceHook)
	Dispatch(path string, reqCtx RequestContext, resp *Response, reqArgs ...LowReqArgs) error
}

type Proxy struct {
	HookBeforeServiceTable map[string][]ProxyBeforeServiceHook
	HookAfterServiceTable  map[string][]ProxyAfterServiceHook
	ServiceTable           map[string]ProxyService

	WebRouterPrefix     string
	StandAloneWebServer Server
	AttachModeWebServer *Server
}

func (p *Proxy) Init() error {
	p.HookBeforeServiceTable = make(map[string][]ProxyBeforeServiceHook)
	p.HookAfterServiceTable = make(map[string][]ProxyAfterServiceHook)
	p.ServiceTable = make(map[string]ProxyService)
	return nil
}

func (p *Proxy) RegisterService(path string, handler interface{}) {
	var service ProxyService
	var funcType = reflect.TypeOf(handler)
	service.Function = reflect.ValueOf(handler)
	service.FunctionName = funcType.Name()

	if service.Function.Kind() != reflect.Func {
		panic("Proxy Router failed, handler is not func, service:" + service.FunctionName)
	}

	// if funcType.NumIn() == 0 {
	// panic("Proxy Router failed, handler params.size should not be 0, service:" + service.FunctionName)
	// }

	service.IsHasReqeustContext = funcType.NumIn() > 0 && strings.HasSuffix(funcType.In(0).String(), "Context")
	// if !strings.HasSuffix(funcType.In(0).Elem().Name(), "Context") {
	// panic("Proxy Router failed, handler params[0] should be RequestContext, service:" + service.FunctionName)
	// }

	// if funcType.NumOut() != 1 {
	// panic("Proxy Router failed, handl...er response.size should be 1, service:" + service.FunctionName)
	// }

	// if funcType.Out(0).Name() != "Response" {
	// panic("Proxy Router failed, handler response[0] should be Response" +
	// ", service:" + service.FunctionName +
	// ", resp:" + funcType.Out(0).Name())
	// }

	var parseArgStartAt = 0
	if service.IsHasReqeustContext {
		parseArgStartAt = 1
	}

	service.IsHasEasyKvReqArgs = false
	if funcType.NumIn() > parseArgStartAt {
		if funcType.In(parseArgStartAt) == reflect.TypeOf(EasyKvReqArgs{}) {
			service.IsHasEasyKvReqArgs = true
			service.Params = append(service.Params, reflect.TypeOf(map[string]interface{}{}))
			parseArgStartAt += 1
		}
	}

	if service.IsHasEasyKvReqArgs {
		if funcType.NumIn() > parseArgStartAt {
			panic("Proxy Router failed, service has EasyKvReqArgs, do not set other params, service:" + service.FunctionName)
		}
	}

	service.IsHasUrlKvReqArgs = false
	if funcType.NumIn() > parseArgStartAt {
		if funcType.In(parseArgStartAt) == reflect.TypeOf(UrlKvReqArgs{}) {
			service.IsHasUrlKvReqArgs = true
			parseArgStartAt += 1
		}
	}

	for i := parseArgStartAt; i < funcType.NumIn(); i++ {
		service.Params = append(service.Params, funcType.In(i))
	}

	p.ServiceTable[path] = service
}

func (p *Proxy) HookBeforeService(path string, hook ProxyBeforeServiceHook) {
	var arr, ok = p.HookBeforeServiceTable[path]
	if !ok {
		arr = []ProxyBeforeServiceHook{}
	}
	arr = append(arr, hook)
	p.HookBeforeServiceTable[path] = arr
}

func (p *Proxy) HookAfterService(path string, hook ProxyAfterServiceHook) {
	var arr, ok = p.HookAfterServiceTable[path]
	if !ok {
		arr = []ProxyAfterServiceHook{}
	}
	arr = append(arr, hook)
	p.HookAfterServiceTable[path] = arr
}

func (p *Proxy) Dispatch(path string,
	reqCtx RequestContext,
	reqArgs ...LowReqArgs) IResponse {
	var (
		resp IResponse
		err  error
	)

	if !p.IsServiceExists(path) {
		err = xerrors.Errorf("%w,path:%s", ErrCmdNotFound, path)
		resp = Response{
			RespCommon{CODE_ERR, err.Error()}, nil,
		}
		return resp
	}

	var (
		paramReflectValueArr []reflect.Value
		service              = p.ServiceTable[path]
	)

	var ok bool
	var paramReflectValueArrIndex = 0
	var reqArgsIndex = 0
	if service.IsHasReqeustContext {
		paramReflectValueArr = make([]reflect.Value, len(reqArgs)+1)
		paramReflectValueArr[paramReflectValueArrIndex] = reflect.ValueOf(reqCtx)
		paramReflectValueArrIndex++
	} else {
		paramReflectValueArr = make([]reflect.Value, len(reqArgs))
	}

	for reqArgsIndex = 0; reqArgsIndex < len(reqArgs); reqArgsIndex++ {
		if paramReflectValueArr[paramReflectValueArrIndex], ok = reqArgs[reqArgsIndex].(reflect.Value); !ok {
			paramReflectValueArr[paramReflectValueArrIndex] = reflect.ValueOf(reqArgs[reqArgsIndex])
		}
		paramReflectValueArrIndex++
	}

	var (
		isContinue       bool
		serviceParamsLen int
		errStr           string
		errCode          int
	)

	serviceParamsLen = len(reqArgs)
	if service.IsHasUrlKvReqArgs {
		serviceParamsLen--
	}

	for hookPath, hooks := range p.HookBeforeServiceTable {
		if strings.HasPrefix(path, hookPath) {
			for _, hook := range hooks {
				resp, isContinue = hook(path, reqCtx, resp, reqArgs...)
				if isContinue == false {
					return resp
				}
			}
			break
		}
	}

	if len(service.Params) != serviceParamsLen {
		err = ErrCmdParamInvalid
		resp = Response{
			RespCommon{CODE_ERR, err.Error()}, nil,
		}
		return resp
	}

	var (
		out    = service.Function.Call(paramReflectValueArr[:])
		outLen = len(out)
		match  bool
		ret    []interface{}
		index  int
	)

	if outLen == 0 {
		resp = Response{
			RespCommon{CODE_OK, ""}, nil,
		}
		goto PARSE_RESP_DONE
	}

	if resp, match = out[0].Interface().(IResponse); match {
		goto PARSE_RESP_DONE
	}

	if outLen == 1 {
		if _, match = out[0].Interface().(error); match {
			resp = Response{
				RespCommon{CODE_ERR, out[0].Interface().(error).Error()}, nil,
			}
			goto PARSE_RESP_DONE
		} else {
			resp = Response{
				RespCommon{CODE_OK, ""}, out[0].Interface(),
			}
			goto PARSE_RESP_DONE
		}
	}

	if out[outLen-1].Type().Name() == "error" {
		if out[outLen-1].Interface() != nil {
			errCode = CODE_ERR
			errStr = out[outLen-1].Interface().(error).Error()
		} else {
			errCode = CODE_OK
			errStr = ""
		}

		if outLen == 2 {
			resp = Response{
				RespCommon{errCode, errStr}, out[0].Interface(),
			}
		} else {
			for index = 0; index < outLen-1; index++ {
				ret = append(ret, out[index].Interface())
			}
			resp = Response{
				RespCommon{errCode, errStr}, ret,
			}
		}
		goto PARSE_RESP_DONE

	}

	for index = 0; index < outLen-1; index++ {
		ret = append(ret, out[index].Interface())
	}
	resp = Response{
		RespCommon{errCode, errStr}, ret,
	}
	goto PARSE_RESP_DONE

PARSE_RESP_DONE:

	for hookPath, hooks := range p.HookAfterServiceTable {
		if strings.HasPrefix(path, hookPath) {
			for _, hook := range hooks {
				resp = hook(path, resp, reqCtx, reqArgs...)
			}
			break
		}
	}

	return resp
}

func (p *Proxy) IsServiceExists(path string) bool {
	var _, ok = p.ServiceTable[path]
	return ok
}
