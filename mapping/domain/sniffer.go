package domain

type Sniffer interface {
	// Sniff 读取dns回复数据包
	Sniff(ipv4 []byte) (int, error)

	Close() error
}
