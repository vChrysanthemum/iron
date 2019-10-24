package iron

import "golang.org/x/xerrors"

var (
	ErrCmdNotFound       = xerrors.New("command not found.")
	ErrCmdParamInvalid   = xerrors.New("command params invalid.")
	ErrCmdParamEmpty     = xerrors.New("command params empty.")
	ErrRespIsNotRespData = xerrors.New("resp is not IRespData")
)
