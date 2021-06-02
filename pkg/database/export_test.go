package data

import (
	"io"

	"github.com/spf13/afero"
)

func OverloadUC(overload func() (string, error)) func() {
	ucRef := uc
	uc = overload
	return func() { uc = ucRef }
}

func OverloadFS(overload afero.Fs) func() {
	fsRef := fs
	fs = overload
	return func() { fs = fsRef }
}

func OverloadPromptReader(overload io.Reader) func() {
	promptReaderRef := promptReader
	promptReader = overload
	return func() { promptReader = promptReaderRef }
}
