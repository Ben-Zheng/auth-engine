package common

import "fmt"

// Error 自定义的Error，如果你希望返回时有特定的http code，可以使用这个Error
// 否则无需使用该Error，直接使用xerror即可
type Error struct {
	Code    int   // 返回的http code
	wrapErr error // err信息
}

func NewCtrlError(code int, wrapErr error) *Error {
	return &Error{
		Code:    code,
		wrapErr: wrapErr,
	}
}

func (e *Error) Error() string {
	return fmt.Sprintf("code: %d, msg: %s", e.Code, e.wrapErr.Error())
}
