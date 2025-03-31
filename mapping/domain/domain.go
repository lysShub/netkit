/*
	domain <==> address
*/

package domain

import (
	"net/netip"
	"sync"
	"syscall"
	"time"

	"github.com/lysShub/netkit/errorx"
	"github.com/lysShub/netkit/mapping"
	"github.com/miekg/dns"
	"github.com/pkg/errors"
	"gvisor.dev/gvisor/pkg/tcpip/header"
)

var (
	once   sync.Once
	global *cache
)

type cache struct {
	sniffer   Sniffer
	assembler *TcpAssembler

	mu    sync.RWMutex
	addrs map[netip.Addr][]string

	closeErr errorx.CloseErr
}

func New() (cache *cache, err error) {
	sniffer, err := newSniffer()
	if err != nil {
		return nil, err
	}
	return NewWithSniffer(sniffer), nil
}

func NewWithSniffer(sniffer Sniffer) *cache {
	once.Do(func() {
		global = newCache(sniffer)
	})
	return global
}

func newCache(sniffer Sniffer) *cache {
	var c = &cache{
		sniffer:   sniffer,
		assembler: NewTcpAssembler(),
		addrs:     map[netip.Addr][]string{},
	}

	go c.service()
	cleanupDnsCache()
	return c
}

func (c *cache) service() (_ error) {
	var ip = make([]byte, 1536)
	for {
		n, err := c.sniffer.Sniff(ip[:cap(ip)])
		if err != nil {
			return c.close(err)
		} else {
			ip = ip[:n]
		}

		hdr := header.IPv4(ip)
		switch proto := hdr.Protocol(); proto {
		case syscall.IPPROTO_UDP:
			i := hdr.HeaderLength() + header.UDPMinimumSize
			if err := c.put(ip[i:]); err != nil {
				return c.close(err)
			}
		case syscall.IPPROTO_TCP:
			data, err := c.assembler.Put(ip, time.Now())
			if err != nil {
				return c.close(err)
			} else if len(data) > 0 {
				for _, msg := range RawDnsOverTcp(data).Msgs() {
					if err := c.put(msg); err != nil {
						return c.close(err)
					}
				}
			}
		default:
			return c.close(errors.Errorf("unknown protocol %d", proto))
		}
	}
}

func (c *cache) put(msg []byte) error {
	var m dns.Msg
	if err := m.Unpack(msg); err != nil {
		return errors.WithStack(err)
	}

	var cname = map[string]string{}
	for _, e := range m.Answer {
		switch e := e.(type) {
		case *dns.A:
			names := []string{unFqdn(e.Hdr.Name)}
			for n := cname[names[0]]; len(n) > 0; {
				names = append(names, n)
				n = cname[n]
			}
			addr := netip.AddrFrom4([4]byte(e.A.To4()))

			c.mu.Lock()
			c.addrs[addr] = names
			c.mu.Unlock()
		case *dns.AAAA:
			names := []string{unFqdn(e.Hdr.Name)}
			for n := cname[names[0]]; len(n) > 0; {
				names = append(names, n)
				n = cname[n]
			}
			addr := netip.AddrFrom16([16]byte(e.AAAA.To16()))

			c.mu.Lock()
			c.addrs[addr] = names
			c.mu.Unlock()
		case *dns.CNAME:
			cname[unFqdn(e.Target)] = unFqdn(e.Hdr.Name)
		default:
			continue
		}
	}
	return nil
}

// unFqdn ref [dns.Fqdn]
func unFqdn(s string) string {
	if n := len(s); n > 1 && s[n-1] == '.' {
		return s[:n-1]
	} else {
		return s
	}
}

func (c *cache) RDNS(a netip.Addr) (names []string) {
	c.mu.RLock()
	names = c.addrs[a]
	c.mu.RUnlock()
	return names
}

func (c *cache) close(cause error) error {
	return c.closeErr.Close(func() (errs []error) {
		errs = append(errs, cause)
		if c.sniffer != nil {
			errs = append(errs, c.sniffer.Close())
		}
		return errs
	})
}
func (c *cache) Close() error { return c.close(nil) }

func RDNS(addr netip.Addr) (names []string, err error) {
	cache, err := New()
	if err != nil {
		return nil, err
	}
	names = cache.RDNS(addr)
	if len(names) == 0 {
		return nil, mapping.ErrNotRecord{}
	}
	return names, nil
}
