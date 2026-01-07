package pipgo

import "net/netip"

type TCPKey struct {
	SrcIP   netip.Addr
	DstIP   netip.Addr
	SrcPort uint16
	DstPort uint16
}
