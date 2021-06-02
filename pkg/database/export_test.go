package data

func OverloadUC(overload func() (string, error)) {
	uc = overload
}

var UC = uc
var FS = fs
