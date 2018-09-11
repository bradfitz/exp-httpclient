package http

type Protocol struct {
	major, minor byte
}

func (p Protocol) Major() int {
	return int(p.major)
}

// TODO
