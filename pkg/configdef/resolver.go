package configdef

import "github.com/tauraamui/xerror"

var ErrConfigAlreadyExists = xerror.New("config file already exists")

type Resolver interface {
	Resolve() (Values, error)
}
