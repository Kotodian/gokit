package exerr

import "errors"

// Error 错误信息封装
type Error struct {
	Code  int
	Error error
}

// Desc 错误文字
func (c *Error) Desc() string {
	return c.Error.Error()
}

// Msg 普通
func Msg(msg string, codes ...int) *Error {
	code := 1
	if len(codes) > 0 {
		code = codes[0]
	}
	return &Error{
		Code:  code,
		Error: errors.New(msg),
	}
}

// Code Code
func Code(code int, msgs ...string) *Error {
	var err error
	if len(msgs) > 0 {
		err = errors.New(msgs[0])
	}
	return &Error{
		Code:  code,
		Error: err,
	}
}
