package types

import (
	"net/netip"
)

type IPHeader struct {
	Version    uint8
	Protocol   uint8
	HasOptions bool
	TTL        uint8
	Headerlen  uint16
	Datalen    uint16
	Src        netip.Addr
	Dst        netip.Addr
}

func NewIPHeader(bytes []byte) *IPHeader {
	header := &IPHeader{}

	version := bytes[0] & 0xF0 >> 4
	if version == 4 {
		hdr := IPHdr(bytes)
		header.Version = hdr.Version()
		header.Protocol = hdr.Protocol()
		header.HasOptions = hdr.IHL() > 5
		header.TTL = hdr.TTL()
		header.Headerlen = uint16(hdr.IHL()) * 4
		header.Datalen = hdr.Len() - header.Headerlen

		header.Src = netip.AddrFrom4([4]byte(hdr.Src()))
		header.Dst = netip.AddrFrom4([4]byte(hdr.Dst()))

	} else {
		hdr := IP6Hdr(bytes)
		header.Version = hdr.Version()
		header.Protocol = hdr.NextHeader()
		header.HasOptions = false
		header.TTL = hdr.HopLimit()
		header.Headerlen = 40
		header.Datalen = hdr.PayloadLen()

		header.Src = netip.AddrFrom16([16]byte(hdr.Src()))
		header.Dst = netip.AddrFrom16([16]byte(hdr.Dst()))
	}

	return header
}
