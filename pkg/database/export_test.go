package data

import "github.com/spf13/afero"

func OverloadUC(overload func() (string, error)) {
	uc = overload
}

func OverloadFS(overload afero.Fs) {
	fs = overload
}
