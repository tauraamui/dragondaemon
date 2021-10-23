package xerror

import "errors"

type I interface {
	Error() string
	Msg(string) I
	WithParams(map[string]string) I
}

type x struct {
	err    error
	params map[string]string
	errMsg string
}

func New() I {
	return &x{
		err: errors.New(""),
	}
}

func (x *x) Error() string {
	return ""
}

func (x *x) Msg(m string) I {
	x.errMsg = m
	return x
}

func (x *x) WithParams(p map[string]string) I {
	x.params = p
	return x
}
