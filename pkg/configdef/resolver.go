package configdef

type Resolver interface {
	Resolve() (Values, error)
}
