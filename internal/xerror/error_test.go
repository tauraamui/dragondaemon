package xerror_test

import (
	"testing"

	"github.com/matryer/is"
	"github.com/tauraamui/dragondaemon/internal/xerror"
)

func TestNewErrorGivesErrInstance(t *testing.T) {
	is := is.New(t)

	err := xerror.New("", "")
	is.True(err != nil)
}

func TestNewErrorGivesErrWhichPrintsOutExpectedString(t *testing.T) {
	is := is.New(t)

	err := xerror.New("TEST_ERROR", "this was caused by something bad: some other wrapped error")
	is.True(err != nil)

	is.Equal(err.ToString(), "Kind: TEST_ERROR, Msg: this was caused by something bad: some other wrapped error")
}

func TestNewErrorWithParamsGivesErrWhichPrintsOutExpectedString(t *testing.T) {
	is := is.New(t)

	err := xerror.New(
		"TEST_PARAMS_ERROR", "fake request failed",
	).WithParam("test-request-trace-id", "31257919-40e6-496b-bb53-b71999222b0b")
	is.True(err != nil)

	is.Equal(err.ToString(), "Kind: TEST_PARAMS_ERROR, Msg: fake request failed, Params: [TEST-REQUEST-TRACE-ID: {31257919-40e6-496b-bb53-b71999222b0b}]")
}
