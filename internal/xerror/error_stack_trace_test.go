package xerror

import (
	"fmt"
	"strings"
	"testing"

	"github.com/matryer/is"
)

func (x *x) SetStackPrinter(p func(error) string) I {
	x.stackPrinter = p
	return x
}

func TestNewErrorWithStackTraceWithReplacedPrinter(t *testing.T) {
	is := is.New(t)

	err := New("fake db update failed").WithStackTrace()
	is.True(err != nil)

	xerr, ok := err.(*x)
	is.True(ok)
	xerr.SetStackPrinter(func(e error) string {
		return fmt.Sprintf("%v\nstack-goes-here", e)
	})

	is.Equal(xerr.Error(), "fake db update failed\nstack-goes-here")
}

func TestNewErrorWithoutStackTraceWithReplacedPrinter(t *testing.T) {
	is := is.New(t)

	err := New("fake db update failed")
	is.True(err != nil)

	xerr, ok := err.(*x)
	is.True(ok)
	xerr.SetStackPrinter(func(e error) string {
		return fmt.Sprintf("%v\nstack-goes-here", e)
	})

	is.Equal(xerr.Error(), "fake db update failed")
}

// basic test to check stack trace output, if necessary might
// improve using regex pattern matching
func TestNewErrorWithStackTraceWithoutReplacedPrinter(t *testing.T) {
	is := is.New(t)

	err := New("fake db update failed").WithStackTrace()
	is.True(err != nil)

	is.True(strings.Contains(err.Error(), "/xerror.(*x).format"))
}
