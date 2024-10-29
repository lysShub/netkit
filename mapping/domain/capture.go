package domain

type Capture interface {
	Capture(ipv4 []byte) (int, error)
	Close() error
}
