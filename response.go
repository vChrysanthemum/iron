package iron

import (
	"encoding/gob"

	"golang.org/x/xerrors"
)

type RespCommon struct {
	Code  int    `json:"Code"`
	Error string `json:"Error"`
}

type RespData = interface{}

type Response struct {
	RespCommon
	RespData `json:"Data"`
}

func (p RespCommon) GetErrorStr() string {
	return p.Error
}

func (p RespCommon) GetError() error {
	if p.Error == "" {
		return nil
	}
	return xerrors.Errorf(p.Error)
}

type IResponse interface {
	GetErrorStr() string
}

func init() {
	gob.Register(Response{})
}
