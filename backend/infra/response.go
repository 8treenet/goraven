package infra

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/8treenet/freedom"
)

type PageResponse struct {
	List       interface{} `json:"list"`
	TotalPage  int         `json:"totalPage"`
	TotalCount int         `json:"totalCount"`
	Page       int         `json:"page,omitempty"`
	PageSize   int         `json:"pageSize,omitempty"`
}

type JSONResponse struct {
	Code      int
	Error     error
	Object    interface{}
	LogOutput bool
}

type ResBodyObject struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data,omitempty"`
}

func (jrep JSONResponse) Dispatch(ctx freedom.Context) {
	var content []byte
	var body ResBodyObject
	body.Data = jrep.Object
	body.Code = jrep.Code

	if jrep.Error != nil {
		body.Msg = jrep.Error.Error()
	}

	if jrep.Error != nil && body.Code == 0 {
		if werr, ok := jrep.Error.(*wrapperError); ok {
			body.Code = werr.code
		}
	}

	if jrep.Error != nil && body.Code == 0 {
		body.Code = 400
	}

	if content, jrep.Error = json.Marshal(body); jrep.Error != nil {
		content = []byte(jrep.Error.Error())
	}

	ctx.Values().Set("code", strconv.Itoa(body.Code))
	if jrep.LogOutput {
		ctx.Values().Set("response", string(content))
	}

	ctx.ContentType("application/json")
	ctx.StatusCode(200)
	if _, err := ctx.Write(content); err != nil {
		freedom.Logger().Error("JSONResponse dispatch error: %v", err)
	}
}

func WrapperError(code int, content string) error {
	return &wrapperError{err: errors.New(content), code: code}
}

func WrapperErrorf(code int, contentFormat string, a ...interface{}) error {
	return &wrapperError{err: fmt.Errorf(contentFormat, a...), code: code}
}

type wrapperError struct {
	err  error
	code int
}

func (e *wrapperError) Error() string {
	return e.err.Error()
}
