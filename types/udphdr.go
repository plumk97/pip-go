package types

import (
	"encoding/binary"
)

type UDPHdr []byte

func NewUDPHdr() UDPHdr {
	return make(UDPHdr, 8)
}

func (hdr UDPHdr) SrcPort() uint16 {
	return binary.BigEndian.Uint16(hdr[0:2])
}

func (hdr UDPHdr) SetSrcPort(n uint16) {
	var b = make([]byte, 2)
	binary.BigEndian.PutUint16(b, n)
	copy(hdr[0:2], b)
}

func (hdr UDPHdr) DstPort() uint16 {
	return binary.BigEndian.Uint16(hdr[2:4])
}

func (hdr UDPHdr) SetDstPort(n uint16) {
	var b = make([]byte, 2)
	binary.BigEndian.PutUint16(b, n)
	copy(hdr[2:4], b)
}

func (hdr UDPHdr) Len() uint16 {
	return binary.BigEndian.Uint16(hdr[4:6])
}

func (hdr UDPHdr) SetLen(n uint16) {
	var b = make([]byte, 2)
	binary.BigEndian.PutUint16(b, n)
	copy(hdr[4:6], b)
}

func (hdr UDPHdr) Sum() uint16 {
	return binary.BigEndian.Uint16(hdr[6:8])
}

func (hdr UDPHdr) SetSum(n uint16) {
	var b = make([]byte, 2)
	binary.BigEndian.PutUint16(b, n)
	copy(hdr[6:8], b)
}
