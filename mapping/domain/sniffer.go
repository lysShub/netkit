package domain

type Sniffer interface {
	Sniff(ipv4 []byte) (int, error)
	Close() error
}
