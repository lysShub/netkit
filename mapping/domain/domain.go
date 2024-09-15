/*
	domain <==> address
*/

package domain

import (
	"encoding/hex"
	"log/slog"
	"net/netip"
	"sync"

	"github.com/lysShub/netkit/debug"
	"github.com/lysShub/netkit/errorx"
	"github.com/lysShub/netkit/packet"
	"github.com/miekg/dns"
	"github.com/pkg/errors"
)

type capture interface {
	Capture(b *packet.Packet) error
	Close() error
}

type Cache struct {
	c capture

	mu    sync.RWMutex
	addrs map[netip.Addr][]string

	closeErr errorx.CloseErr
}

func NewCache() (*Cache, error) {
	var c = &Cache{addrs: map[netip.Addr][]string{}}
	var err error

	c.c, err = newCapture()
	if err != nil {
		return nil, err
	}

	go c.captureService()
	return c, nil
}

func (c *Cache) put(pkt *packet.Packet) error {
	return c.putRaw(pkt.Bytes())
}

func (c *Cache) putRaw(b []byte) error {
	var m dns.Msg
	if err := m.Unpack(b); err != nil {
		return errors.WithStack(err)
	}

	var addr netip.Addr
	var cname = map[string]string{}
	for _, e := range m.Answer {
		switch e := e.(type) {
		case *dns.A:
			addr = netip.AddrFrom4([4]byte(e.A.To4()))
		case *dns.AAAA:
			addr = netip.AddrFrom16([16]byte(e.AAAA.To16()))
		case *dns.CNAME:
			cname[e.Target] = e.Hdr.Name
			continue
		default:
			continue
		}

		ns := []string{trim(e.Header().Name)}
		for t := e.Header().Name; ; {
			if n, has := cname[t]; has {
				ns = append(ns, trim(n))
				t = n
			} else {
				break
			}
		}
		if debug.Debug() && len(ns) != len(cname)+1 {
			println("exception dns packet: ", hex.EncodeToString(b))
		}

		c.mu.Lock()
		c.addrs[addr] = ns
		c.mu.Unlock()
	}
	return nil
}

func trim(s string) string {
	i := len(s) - 1
	if i < 0 {
		return s
	} else if s[i] == '.' {
		return s[:i]
	} else {
		return s
	}
}

func (c *Cache) RDNS(a netip.Addr) (names []string) {
	c.mu.RLock()
	names = c.addrs[a]
	c.mu.RUnlock()
	return names
}

func (c *Cache) close(cause error) error {
	return c.closeErr.Close(func() (errs []error) {
		if debug.Debug() && cause != nil {
			slog.Warn(cause.Error(), errorx.Trace(cause)) // todo: add log opt
		}
		errs = append(errs, cause)
		if c.c != nil {
			errs = append(errs, c.c.Close())
		}
		return errs
	})
}
func (c *Cache) Close() error { return c.close(nil) }
func (c *Cache) Closed() bool { return c.closeErr.Closed() }

func (c *Cache) captureService() (_ error) {
	var b = packet.Make(0, 2048)

	for {
		err := c.c.Capture(b.Sets(0, 0xffff))
		if err != nil {
			return c.close(err)
		}

		if err := c.put(b); err != nil {
			if debug.Debug() {
				slog.Warn(err.Error(), errorx.Trace(err))
			}
		}
	}
}
