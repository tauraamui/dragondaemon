package configdef

type ErrConfigAlreadyExists error

type Resolver interface {
	Create() error
	Resolve() (Values, error)
	Destroy() error
}
