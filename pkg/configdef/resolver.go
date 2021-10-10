package configdef

import "errors"

var ErrConfigAlreadyExists = errors.New("config file already exists")

type Resolver interface {
	Resolve() (Values, error)
}
