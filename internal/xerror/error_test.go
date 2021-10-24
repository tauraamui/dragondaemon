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

type xerrorTest struct {
	skip     bool
	title    string
	err      error
	expected string
}

func TestNewErrorOutputsExpectedString(t *testing.T) {
	tests := []xerrorTest{
		{
			title:    "simple new error just prints out msg string",
			err:      xerror.New("fake db update failed"),
			expected: "fake db update failed",
		},
		{
			title:    "new error with param prints out msg string with not assigned kind and with param",
			err:      xerror.New("fake db update failed").WithParam("trace-request-id", "efw4fv32f"),
			expected: "Kind: N/A | fake db update failed, Params: [trace-request-id: {efw4fv32f}]",
		},
		{
			title:    "new error with kind prints out msg string with assigned kind",
			err:      xerror.NewWithKind(TestError, "fake db update failed"),
			expected: "Kind: TEST_ERROR | fake db update failed",
		},
		{
			title:    "new error with kind and param prints out msg string with assigned kind and with param",
			err:      xerror.NewWithKind(TestParamsError, "fake db update failed").WithParam("trace-request-id", "wdgrte4fg"),
			expected: "Kind: TEST_PARAMS_ERROR | fake db update failed, Params: [trace-request-id: {wdgrte4fg}]",
		},
		{
			title: "new error with not assigned kind and with params prints out msg string with not assigned kind and with params",
			err: xerror.New("fake db update failed").WithParams(
				map[string]interface{}{
					"trace-request-id": "fr495fre",
					"request-ip":       "39.49.13.45",
				},
			),
			expected: "Kind: N/A | fake db update failed, Params: [trace-request-id: {fr495fre} | request-ip: {39.49.13.45}]",
		},
	}

	for _, tt := range tests {
		runTest(t, tt)
	}
}

func runTest(t *testing.T, tt xerrorTest) {
	t.Run(tt.title, func(t *testing.T) {
		if len(tt.title) == 0 {
			t.Error("table tests must all have titles")
		}

		if tt.skip {
			t.Skip()
		}

		is := is.NewRelaxed(t)

		is.Equal(tt.err.Error(), tt.expected)
	})
}
