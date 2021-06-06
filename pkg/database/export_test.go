package data

import (
	"io"

	"github.com/spf13/afero"
)

func NewStdinPlainReader(readFrom io.Reader) stdinPlainReader {
	return stdinPlainReader{
		readFrom: readFrom,
	}
}

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

func OverloadPlainPromptReader(overload plainReader) func() {
	plainPromptReaderRef := plainPromptReader
	plainPromptReader = overload
	return func() { plainPromptReader = plainPromptReaderRef }
}

func OverloadPasswordPromptReader(overload passwordReader) func() {
	passwordPromptReaderRef := passwordPromptReader
	passwordPromptReader = overload
	return func() { passwordPromptReader = passwordPromptReaderRef }
}
