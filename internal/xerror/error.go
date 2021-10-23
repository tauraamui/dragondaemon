package xerror

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
)

type I interface {
	AsKind(Kind) I
	ToError() error
	Error() string
	ErrorMsg() string
	Msg(string) I
	WithParam(string, interface{}) I
	WithParams(map[string]interface{}) I
}

type Kind string

const NA = Kind("N/A")

type x struct {
	kind   Kind
	errMsg string
	error  error
	params map[string]interface{}
}

func Errorf(format string, values ...interface{}) I {
	return New(NA, fmt.Errorf(format, values...).Error())
}

func New(k Kind, es string) I {
	i := x{kind: k, errMsg: es}
	i.format()
	return &i
}

func (x *x) format() {
	x.error = errors.WithStack(errors.New(x.toString()))
}

func (x *x) AsKind(k Kind) I {
	defer x.format()
	x.kind = k
	return x
}

func (x *x) ToError() error {
	return x.error
}

func (x *x) Error() string {
	return x.error.Error()
}

func (x *x) ErrorMsg() string {
	return x.errMsg
}

func (x *x) Msg(m string) I {
	defer x.format()
	x.errMsg = m
	return x
}

// WithParams will set or merge given params map to use
// as the param values. Passing nil instead of a map will
// clear the params completely.
func (x *x) WithParams(p map[string]interface{}) I {
	defer x.format()
	if p == nil {
		x.params = p
		return x
	}

	if x.params != nil {
		mergeParams(x.params, p)
		return x
	}
	return x
}

func mergeParams(p1, p2 map[string]interface{}) {
	for k, v := range p2 {
		p1[k] = v
	}
}

// WithParam will append the param key/value pair passed in
// to the internal map.
func (x *x) WithParam(key string, v interface{}) I {
	defer x.format()
	if x.params == nil {
		x.params = map[string]interface{}{}
	}
	x.params[key] = v
	return x
}

func (x *x) toString() string {
	logMsg := fmt.Sprintf("Kind: %s | %s", x.kind, x.errMsg)

	params := []string{}
	for k, v := range x.params {
		params = append(params, fmt.Sprintf("%s: {%+v}", k, v))
	}
	return fmt.Sprintf("%s%s", logMsg, func() string {
		if len(params) != 0 {
			return fmt.Sprintf(", Params: [%+v]", strings.Join(params, " | "))
		}
		return ""
	}())
}
