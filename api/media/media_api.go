package mediaapi

type Connection interface {
	UUID() string
	Title() string
}
