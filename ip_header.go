package pipgo

import (
	"encoding/binary"
	"net"

	"github.com/plumk97/pip-go/types"
)

type IPHeader struct {
	Version    uint8
	Protocol   uint8
	HasOptions bool
	TTL        uint8
	Headerlen  uint16
	Datalen    uint16
	Src        net.IP
	Dst        net.IP
}

func NewIPHeader(bytes []byte) *IPHeader {
	header := &IPHeader{}

	version := bytes[0] & 0xF0 >> 4
	if version == 4 {
		hdr := types.IPHdr(bytes)
		header.Version = hdr.Version()
		header.Protocol = hdr.Protocol()
		header.HasOptions = hdr.IHL() > 5
		header.TTL = hdr.TTL()
		header.Headerlen = uint16(hdr.IHL()) * 4
		header.Datalen = hdr.Len() - header.Headerlen
		header.Src = hdr.Src()
		header.Dst = hdr.Dst()
	} else {
		hdr := types.IP6Hdr(bytes)
		header.Version = hdr.Version()
		header.Protocol = hdr.NextHeader()
		header.HasOptions = false
		header.TTL = hdr.HopLimit()
		header.Headerlen = 40
		header.Datalen = hdr.PayloadLen()
		header.Src = hdr.Src()
		header.Dst = hdr.Dst()
	}

	return header
}

// 生成32位标识
func (h *IPHeader) GenerateIden() uint32 {
	if h.Version == 4 {
		return binary.BigEndian.Uint32(h.Src) ^ binary.BigEndian.Uint32(h.Dst) ^ 4
	}

	var iden uint32 = 0
	for i := 0; i < 16; i += 4 {
		iden ^= binary.BigEndian.Uint32(h.Src[i : i+4])
	}

	for i := 0; i < 16; i += 4 {
		iden ^= binary.BigEndian.Uint32(h.Dst[i : i+4])
	}
	iden ^= 6
	return iden
}
