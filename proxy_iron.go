package iron

import (
	"encoding/json"
	"io/ioutil"
	"reflect"

	"golang.org/x/xerrors"
)

func (p *Proxy) DispatchWithIronRequest(path string, reqCtx RequestContext, req *Request) IResponse {
	var (
		resp Response
		err  error
	)
	if !p.IsServiceExists(path) {
		err = xerrors.Errorf("%w,path:%s", ErrCmdNotFound, path)
		resp = Response{
			RespCommon{CODE_ERR, err.Error()}, nil,
		}
		return resp
	}

	var reqArgBytes []byte
	reqArgBytes, err = ioutil.ReadAll(req.R.Body)
	if err != nil {
		resp = Response{
			RespCommon{CODE_ERR, err.Error()}, nil,
		}
		return resp
	}

	var service = p.ServiceTable[path]
	var reqArgElems []interface{}

	var parseEasyKvReqArgs = func() error {
		var reqArgs = MakeEasyKvReqArgs()

		//merge url params
		reqArgs.MergeIronRequest(req)

		//merge body params
		if len(reqArgBytes) != 0 {
			var ret = make(map[string]interface{})
			err = json.Unmarshal(reqArgBytes, &ret)
			if err != nil {
				return err
			}
			reqArgs.MergeKv(ret)
		}

		reqArgElems = append(reqArgElems, reqArgs)
		return nil
	}

	var parseNormalReqArgs = func() error {
		// parse QueryString
		if service.IsHasUrlKvReqArgs {
			var reqArgs = MakeUrlKvReqArgs()
			reqArgs.MergeIronRequest(req)
			reqArgElems = append(reqArgElems, reqArgs)
		}

		// parse http body json
		if len(reqArgBytes) == 0 {
			return nil
		}

		var reqArgValues []reflect.Value
		var reqArgInterfaces []interface{}
		for i, _ := range service.Params {
			var serviceParam = service.Params[i]
			var reqArgValue = reflect.New(serviceParam)
			reqArgValues = append(reqArgValues, reqArgValue)
			reqArgInterfaces = append(reqArgInterfaces, reqArgValue.Interface())
		}

		if len(reqArgInterfaces) == 1 {
			err = json.Unmarshal(reqArgBytes, &reqArgInterfaces[0])
		} else {
			err = json.Unmarshal(reqArgBytes, &reqArgInterfaces)
		}

		if err != nil {
			return err
		}

		for i, _ := range reqArgValues {
			reqArgElems = append(reqArgElems, reqArgValues[i].Elem())
		}
		return nil
	}

	// parse EasyKvReqArgs
	if service.IsHasEasyKvReqArgs {
		err = parseEasyKvReqArgs()
	} else {
		err = parseNormalReqArgs()
	}

	if err != nil {
		resp = Response{
			RespCommon{CODE_ERR, err.Error()}, nil,
		}
		return resp
	}

	return p.Dispatch(path, reqCtx, reqArgElems...)
}
