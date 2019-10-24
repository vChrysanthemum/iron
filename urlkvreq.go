package iron

import (
	"strconv"
)

type UrlKvReqArgs map[string]interface{}

func MakeUrlKvReqArgs(kvArr ...interface{}) UrlKvReqArgs {
	var ret UrlKvReqArgs = make(map[string]interface{})
	for i := 0; i < len(kvArr); i += 1 {
		ret[kvArr[i].(string)] = kvArr[i+1]
	}
	return ret
}

func (p UrlKvReqArgs) MergeIronRequest(ir *Request) {
	ir.prepareForm()
	for k, _ := range ir.R.Form {
		if len(ir.R.Form[k]) == 0 {
			continue
		}
		p[k] = ir.R.Form[k][0]
	}
}

func (p UrlKvReqArgs) MustFormString(key string, defaultRet string) (ret string) {
	if tmp, ok := p[key].(string); ok && tmp != "" {
		return tmp
	} else {
		return defaultRet
	}
}

func (p UrlKvReqArgs) MustFormFloat64(key string, defaultRet float64) float64 {
	switch p[key].(type) {
	case float64:
		return p[key].(float64)
	case string:
		ret, err := strconv.ParseFloat(p[key].(string), 64)
		if err != nil {
			return defaultRet
		}
		return ret
	}
	return defaultRet
}

func (p UrlKvReqArgs) MustFormInt64(key string, defaultRet int64) int64 {
	switch p[key].(type) {
	case int64:
		return p[key].(int64)
	case string:
		ret, err := strconv.ParseInt(p[key].(string), 10, 64)
		if err != nil {
			return defaultRet
		}
		return ret
	}
	return defaultRet
}

func (p UrlKvReqArgs) MustFormUint64(key string, defaultRet uint64) uint64 {
	switch p[key].(type) {
	case uint64:
		return p[key].(uint64)
	case string:
		ret, err := strconv.ParseUint(p[key].(string), 10, 64)
		if err != nil {
			return defaultRet
		}
		return ret
	}
	return defaultRet
}

func (p UrlKvReqArgs) MustFormInt(key string, defaultRet int) int {
	switch p[key].(type) {
	case int:
		return p[key].(int)
	case string:
		ret, err := strconv.ParseInt(p[key].(string), 10, 64)
		if err != nil {
			return defaultRet
		}
		return int(ret)
	}
	return defaultRet
}
