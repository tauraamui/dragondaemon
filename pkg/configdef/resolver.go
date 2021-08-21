package configdef

import "errors"

var ErrConfigAlreadyExists = errors.New("config file already exists")

type Resolver interface {
	Create() error
	Resolve() (Values, error)
	Destroy() error
}
