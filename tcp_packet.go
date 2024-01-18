package pipgo

import (
	"syscall"
	"time"

	"github.com/plumk97/pip-go/types"
)

type TCPPacket struct {
	headBuf    *Buffer
	payloadLen int
	sendTime   time.Time
	sendCount  uint8
}

func newTCPPacket(tcp *TCP, flags uint8, optionBuf *Buffer, payloadBuf *Buffer) *TCPPacket {
	packet := &TCPPacket{}
	packet.headBuf = NewBuffer(types.NewTCPHdr())

	if optionBuf != nil {
		packet.headBuf.SetNext(optionBuf)
		optionBuf.SetNext(payloadBuf)
	} else {
		packet.headBuf.SetNext(payloadBuf)
	}

	if payloadBuf != nil {
		packet.payloadLen = payloadBuf.totalLen
	}

	hdr := packet.hdr()
	hdr.SetSrcPort(tcp.dstPort)
	hdr.SetDstPort(tcp.srcPort)
	hdr.SetSeq(tcp.seq)
	hdr.SetAck(tcp.ack)

	headlen := len(hdr)
	if optionBuf != nil {
		headlen += len(optionBuf.payload)
	}

	hdr.SetOff(uint8(headlen / 4))
	hdr.SetFlags(flags)
	hdr.SetWin(tcp.wind)
	hdr.SetSum(InetChecksumBuf(packet.headBuf, syscall.IPPROTO_TCP, tcp.ipHeader.Dst, tcp.ipHeader.Src))

	return packet
}

func (packet *TCPPacket) sended() {
	packet.sendTime = time.Now()
	packet.sendCount += 1
}

func (packet *TCPPacket) hdr() types.TCPHdr {
	return types.TCPHdr(packet.headBuf.payload)
}
