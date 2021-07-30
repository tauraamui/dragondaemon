package config

type Camera struct {
	Title   string
	Address string
}

type Values struct {
	Cameras []Camera
}

func Load() Values {
	return Values{}
}
