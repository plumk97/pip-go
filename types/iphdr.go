package types

import (
	"encoding/binary"
	"net"
)

type IPHdr []byte

func NewIPHdr() IPHdr {
	return make(IPHdr, 20)
}

func (hdr IPHdr) Version() uint8 {
	return hdr[0] >> 4
}

func (hdr IPHdr) SetVersion(n uint8) {
	hdr[0] = n<<4 | hdr.IHL()
}

func (hdr IPHdr) IHL() uint8 {
	return hdr[0] & 0x0F
}

func (hdr IPHdr) SetIHL(n uint8) {
	hdr[0] = hdr.Version()<<4 | n
}

func (hdr IPHdr) Tos() uint8 {
	return hdr[1]
}

func (hdr IPHdr) SetTos(n uint8) {
	hdr[1] = n
}

func (hdr IPHdr) Len() uint16 {
	return binary.BigEndian.Uint16(hdr[2:4])
}

func (hdr IPHdr) SetLen(n uint16) {
	b := make([]byte, 2)
	binary.BigEndian.PutUint16(b, n)
	copy(hdr[2:4], b)
}

func (hdr IPHdr) ID() uint16 {
	return binary.BigEndian.Uint16(hdr[4:6])
}

func (hdr IPHdr) SetID(n uint16) {
	b := make([]byte, 2)
	binary.BigEndian.PutUint16(b, n)
	copy(hdr[4:6], b)
}

func (hdr IPHdr) Off() uint16 {
	return binary.BigEndian.Uint16(hdr[6:8])
}

func (hdr IPHdr) SetOff(n uint16) {
	b := make([]byte, 2)
	binary.BigEndian.PutUint16(b, n)
	copy(hdr[6:8], b)
}

func (hdr IPHdr) TTL() uint8 {
	return hdr[8]
}

func (hdr IPHdr) SetTTL(n uint8) {
	hdr[8] = n
}

func (hdr IPHdr) Protocol() uint8 {
	return hdr[9]
}

func (hdr IPHdr) SetProtocol(n uint8) {
	hdr[9] = n
}

func (hdr IPHdr) Sum() uint16 {
	return binary.BigEndian.Uint16(hdr[10:12])
}

func (hdr IPHdr) SetSum(n uint16) {
	b := make([]byte, 2)
	binary.BigEndian.PutUint16(b, n)
	copy(hdr[10:12], b)
}

func (hdr IPHdr) Src() net.IP {
	return net.IP(hdr[12:16])
}

func (hdr IPHdr) SetSrc(ip net.IP) {
	copy(hdr[12:16], ip.To4())
}

func (hdr IPHdr) Dst() net.IP {
	return net.IP(hdr[16:20])
}

func (hdr IPHdr) SetDst(ip net.IP) {
	copy(hdr[16:20], ip.To4())
}
