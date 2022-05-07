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
	WithStackTrace() I
	// WithParam will append the param key/value pair passed in
	// to the internal map.
	WithParam(string, interface{}) I
	// WithParams will set or merge given params map to use
	// as the param values. Passing nil instead of a map will
	// clear the params completely.
	WithParams(map[string]interface{}) I
}

type Kind string

const NA = Kind("N/A")

type x struct {
	kind         Kind
	causeErr     error
	errMsg       string
	error        error
	stackTrace   bool
	stackPrinter func(error) string
	params       map[string]interface{}
}

func Errorf(format string, values ...interface{}) I {
	return newFromError(fmt.Errorf(format, values...))
}

func newFromError(e error) I {
	var cause error
	if c := errors.Unwrap(e); c != nil {
		cause = c
	}
	return &x{
		error:    e,
		causeErr: cause,
		kind:     NA, errMsg: e.Error(), stackTrace: false,
		stackPrinter: func(e error) string {
			return fmt.Sprintf("%+v", e)
		},
	}
}

func New(es string) I {
	return NewWithKind(NA, es)
}

func NewWithKind(k Kind, es string) I {
	i := x{
		kind: k, errMsg: es, stackTrace: false,
		stackPrinter: func(e error) string {
			return fmt.Sprintf("%+v", e)
		},
	}
	i.format()
	return &i
}

func (x *x) format() {
	err := errors.New(x.toString())
	x.error = err
}

func (x *x) Is(e error) bool {
	if errors.Is(x.error, e) || errors.Is(x.causeErr, e) {
		return true
	}

	if err := errors.Unwrap(x.error); err != nil {
		return errors.Is(err, e)
	}

	return false
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
	if x.stackTrace {
		return x.stackPrinter(x.error)
	}
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

func (x *x) WithStackTrace() I {
	x.stackTrace = true
	x.format()
	return x
}

func (x *x) WithParams(p map[string]interface{}) I {
	defer x.format()

	if x.params != nil {
		mergeParams(x.params, p)
		return x
	}

	x.params = p
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
	logMsg := x.errMsg
	if x.kind != NA || len(x.params) > 0 {
		logMsg = fmt.Sprintf("Kind: %s | %s", strings.ToUpper(string(x.kind)), x.errMsg)
	}

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
