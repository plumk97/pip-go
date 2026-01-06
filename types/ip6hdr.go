package types

import (
	"encoding/binary"
	"net"
)

type IP6Hdr []byte

func NewIP6Hdr() IP6Hdr {
	return make(IP6Hdr, 40)
}

func (hdr IP6Hdr) un1() uint32 {
	return binary.BigEndian.Uint32(hdr[0:4])
}

func (hdr IP6Hdr) setUn1(n uint32) {
	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b, n)
	copy(hdr[0:4], b)
}

func (hdr IP6Hdr) un2() uint32 {
	return binary.BigEndian.Uint32(hdr[4:8])
}

func (hdr IP6Hdr) setUn2(n uint32) {
	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b, n)
	copy(hdr[4:8], b)
}

func (hdr IP6Hdr) Version() uint8 {
	return uint8((hdr.un1() & 0xF0000000) >> 28)
}

func (hdr IP6Hdr) SetVersion(n uint8) {
	un1 := hdr.un1()
	un1 = uint32(n)<<28 | un1&^0xF0000000
	hdr.setUn1(un1)
}

func (hdr IP6Hdr) TrafficClass() uint8 {
	return uint8((hdr.un1() & 0x0FF00000) >> 20)
}

func (hdr IP6Hdr) SetTrafficClass(n uint8) {
	un1 := hdr.un1()
	un1 = uint32(n)<<20 | un1&^0x0FF00000
	hdr.setUn1(un1)
}

func (hdr IP6Hdr) Flow() uint32 {
	return hdr.un1() & 0x000FFFFF
}

func (hdr IP6Hdr) SetFlow(n uint32) {
	un1 := hdr.un1()
	un1 = n | un1&^0x000FFFFF
	hdr.setUn1(un1)
}

func (hdr IP6Hdr) PayloadLen() uint16 {
	return uint16((hdr.un2() & 0xFFFF0000) >> 16)
}

func (hdr IP6Hdr) SetPayloadLen(n uint16) {
	un2 := hdr.un2()
	un2 = uint32(n)<<16 | un2&^0xFFFF0000
	hdr.setUn2(un2)
}

func (hdr IP6Hdr) NextHeader() uint8 {
	return uint8((hdr.un2() & 0x0000FF00) >> 8)
}

func (hdr IP6Hdr) SetNextHeader(n uint8) {
	un2 := hdr.un2()
	un2 = uint32(n)<<8 | un2&^0x0000FF00
	hdr.setUn2(un2)
}

func (hdr IP6Hdr) HopLimit() uint8 {
	return uint8((hdr.un2() & 0x000000FF))
}

func (hdr IP6Hdr) SetHopLimit(n uint8) {
	un2 := hdr.un2()
	un2 = uint32(n) | un2&^0x000000FF
	hdr.setUn2(un2)
}

func (hdr IP6Hdr) Src() net.IP {
	return net.IP(hdr[8:24])
}

func (hdr IP6Hdr) SetSrc(ip net.IP) {
	copy(hdr[8:24], ip.To16())
}

func (hdr IP6Hdr) Dst() net.IP {
	return net.IP(hdr[24:40])
}

func (hdr IP6Hdr) SetDst(ip net.IP) {
	copy(hdr[24:40], ip.To16())
}
