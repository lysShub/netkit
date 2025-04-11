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
	"github.com/miekg/dns"
	"gvisor.dev/gvisor/pkg/tcpip/header"
)

type Cache struct {
	sniffer   Sniffer
	assembler *TcpAssembler

	mu    sync.RWMutex
	addrs map[netip.Addr][]string

	closeErr errorx.CloseErr
}

func New() (cache *Cache, err error) {
	sniffer, err := newSniffer()
	if err != nil {
		return nil, err
	}
	return NewWithSniffer(sniffer), nil
}

func NewWithSniffer(sniffer Sniffer) *Cache {
	return newCache(sniffer)
}

func newCache(sniffer Sniffer) *Cache {
	var c = &Cache{
		sniffer:   sniffer,
		assembler: NewTcpAssembler(),
		addrs:     map[netip.Addr][]string{},
	}

	go c.service()
	cleanupDnsCache()
	return c
}

func (c *Cache) service() (_ error) {
	var ip = make(header.IPv4, 1536)
	for {
		n, err := c.sniffer.Sniff(ip[:cap(ip)])
		if err != nil {
			return c.close(err)
		} else {
			ip = ip[:n]
		}

		switch proto := ip.Protocol(); proto {
		case syscall.IPPROTO_UDP:
			i := ip.HeaderLength() + header.UDPMinimumSize
			c.put(ip[i:])
		case syscall.IPPROTO_TCP:
			data, err := c.assembler.Put(ip, time.Now())
			if err != nil {
			} else {
				for _, msg := range RawDnsOverTcp(data).Msgs() {
					c.put(msg)
				}
			}
		default:
		}
	}
}

func (c *Cache) put(msg []byte) {
	var m dns.Msg
	if err := m.Unpack(msg); err != nil {
		return
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
}

// unFqdn ref [dns.Fqdn]
func unFqdn(s string) string {
	if n := len(s); n > 1 && s[n-1] == '.' {
		return s[:n-1]
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
		errs = append(errs, cause)
		if c.sniffer != nil {
			errs = append(errs, c.sniffer.Close())
		}
		return errs
	})
}
func (c *Cache) Close() error { return c.close(nil) }

func RDNS(addr netip.Addr) (names []string, err error) {
	cache, err := New()
	if err != nil {
		return nil, err
	}
	names = cache.RDNS(addr)
	if len(names) == 0 {
		return nil, errorx.WrapNotfound(errorx.ErrEmpty)
	}
	return names, nil
}
