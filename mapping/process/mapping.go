package process

import (
	"net/netip"
)

// mapping of local-addr <==> process
type Mapping interface {
	Close() error

	Name(laddr netip.AddrPort, proto uint8) (string, error)
	Pid(laddr netip.AddrPort, proto uint8) (uint32, error)

	Pids() []uint32
	Names() []string
}

func New() (Mapping, error) {
	return newMapping()
}

type ErrNotRecord struct{}

func (e ErrNotRecord) Error() string { return "not record" }
func (ErrNotRecord) Temporary() bool { return true }
