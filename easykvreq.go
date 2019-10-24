package iron

type EasyKvReqArgs struct {
	UrlKvReqArgs
}

func MakeEasyKvReqArgs(kvArr ...interface{}) EasyKvReqArgs {
	var ret EasyKvReqArgs
	ret.UrlKvReqArgs = MakeUrlKvReqArgs(kvArr...)
	return ret
}

func (p EasyKvReqArgs) MergeKv(kv map[string]interface{}) EasyKvReqArgs {
	for k, v := range kv {
		p.UrlKvReqArgs[k] = v
	}
	return p
}
