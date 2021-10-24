package xerror_test

import (
	"testing"

	"github.com/matryer/is"
	"github.com/tauraamui/dragondaemon/internal/xerror"
)

func TestNewErrorGivesErrInstance(t *testing.T) {
	is := is.New(t)

	err := xerror.New("")
	is.True(err != nil)
}

const TestError = xerror.Kind("test_error")
const TestParamsError = xerror.Kind("test_params_error")

func TestNewErrorGivesErrWhichPrintsOutExpectedString(t *testing.T) {
	is := is.New(t)

	err := xerror.New("fake db update failed")
	is.True(err != nil)

	is.Equal(err.Error(), "fake db update failed")
}

func TestNewErrorWithParamGivesErrWhichPrintsOutExpectedString(t *testing.T) {
	is := is.New(t)

	err := xerror.New("fake db update failed").WithParam("trace-request-id", "efw4fv32f")
	is.True(err != nil)

	is.Equal(err.Error(), "Kind: N/A | fake db update failed, Params: [trace-request-id: {efw4fv32f}]")
}

func TestNewErrorWithKindGivesErrWhichPrintsOutExpectedString(t *testing.T) {
	is := is.New(t)

	err := xerror.NewWithKind(TestError, "this was caused by something bad: some other wrapped error")
	is.True(err != nil)

	is.Equal(err.Error(), "Kind: TEST_ERROR | this was caused by something bad: some other wrapped error")
}

func TestNewErrorWithKindAndParamGivesErrWhichPrintsOutExpectedString(t *testing.T) {
	is := is.New(t)

	err := xerror.NewWithKind(
		TestParamsError, "fake request failed",
	).WithParam("test-request-trace-id", "31257919-40e6-496b-bb53-b71999222b0b")
	is.True(err != nil)

	is.Equal(err.Error(), "Kind: TEST_PARAMS_ERROR | fake request failed, Params: [test-request-trace-id: {31257919-40e6-496b-bb53-b71999222b0b}]")
}
