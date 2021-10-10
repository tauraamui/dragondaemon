package configdef

type Creator interface {
	Create() error
}
