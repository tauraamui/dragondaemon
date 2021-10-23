package xerror

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
)

type I interface {
	Error() string
	Msg(string) I
	WithParam(string, interface{}) I
	WithParams(map[string]interface{}) I
	ToString() string
}

type Kind string

type x struct {
	kind   Kind
	errMsg string
	error  error
	params map[string]interface{}
}

func New(k Kind, es string) I {
	i := x{kind: k, errMsg: es}
	e := fmt.Errorf("KIND: %s | %s", i.kind, i.errMsg)
	i.error = errors.WithStack(e)
	return &i
}

func (x *x) Error() string {
	return ""
}

func (x *x) Msg(m string) I {
	x.errMsg = m
	return x
}

// WithParams will set or merge given params map to use
// as the param values. Passing nil instead of a map will
// clear the params completely.
func (x *x) WithParams(p map[string]interface{}) I {
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
	if x.params == nil {
		x.params = map[string]interface{}{}
	}
	x.params[key] = v
	return x
}

func (x *x) ToString() string {
	logMsg := fmt.Sprintf("Kind: %s, Msg: %s", x.kind, x.errMsg)

	params := []string{}
	for k, v := range x.params {
		params = append(params, fmt.Sprintf("%s: {%+v}", strings.ToUpper(k), v))
	}
	return fmt.Sprintf("%s%s", logMsg, func() string {
		if len(params) != 0 {
			return fmt.Sprintf(", Params: [%+v]", strings.Join(params, " | "))
		}
		return ""
	}())
}
