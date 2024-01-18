package types

import "encoding/binary"

type TCPHdr []byte

const (
	TH_FIN  = 0x01
	TH_SYN  = 0x02
	TH_RST  = 0x04
	TH_PUSH = 0x08
	TH_ACK  = 0x10
	TH_URG  = 0x20
	TH_ECE  = 0x40
	TH_CWR  = 0x80
	TH_AE   = 0x100
)

func NewTCPHdr() TCPHdr {
	return make(TCPHdr, 20)
}

func (hdr TCPHdr) SrcPort() uint16 {
	return binary.BigEndian.Uint16(hdr[0:2])
}

func (hdr TCPHdr) SetSrcPort(n uint16) {
	b := make([]byte, 2)
	binary.BigEndian.PutUint16(b, n)
	copy(hdr[0:2], b)
}

func (hdr TCPHdr) DstPort() uint16 {
	return binary.BigEndian.Uint16(hdr[2:4])
}

func (hdr TCPHdr) SetDstPort(n uint16) {
	b := make([]byte, 2)
	binary.BigEndian.PutUint16(b, n)
	copy(hdr[2:4], b)
}

func (hdr TCPHdr) Seq() uint32 {
	return binary.BigEndian.Uint32(hdr[4:8])
}

func (hdr TCPHdr) SetSeq(n uint32) {
	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b, n)
	copy(hdr[4:8], b)
}

func (hdr TCPHdr) Ack() uint32 {
	return binary.BigEndian.Uint32(hdr[8:12])
}

func (hdr TCPHdr) SetAck(n uint32) {
	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b, n)
	copy(hdr[8:12], b)
}

func (hdr TCPHdr) Off() uint8 {
	return (hdr[12] & 0xF0) >> 4
}

func (hdr TCPHdr) SetOff(n uint8) {
	hdr[12] = n << 4
}

func (hdr TCPHdr) Flags() uint8 {
	return hdr[13]
}

func (hdr TCPHdr) SetFlags(n uint8) {
	hdr[13] = n
}

func (hdr TCPHdr) Win() uint16 {
	return binary.BigEndian.Uint16(hdr[14:16])
}

func (hdr TCPHdr) SetWin(n uint16) {
	b := make([]byte, 4)
	binary.BigEndian.PutUint16(b, n)
	copy(hdr[14:16], b)
}

func (hdr TCPHdr) Sum() uint16 {
	return binary.BigEndian.Uint16(hdr[16:18])
}

func (hdr TCPHdr) SetSum(n uint16) {
	b := make([]byte, 4)
	binary.BigEndian.PutUint16(b, n)
	copy(hdr[16:18], b)
}

func (hdr TCPHdr) Urp() uint16 {
	return binary.BigEndian.Uint16(hdr[18:20])
}

func (hdr TCPHdr) SetUrp(n uint16) {
	b := make([]byte, 4)
	binary.BigEndian.PutUint16(b, n)
	copy(hdr[18:20], b)
}
