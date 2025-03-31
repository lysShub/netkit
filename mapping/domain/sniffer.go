package domain

type Sniffer interface {
	Sniffer(ipv4 []byte) (int, error)
	Close() error
}
