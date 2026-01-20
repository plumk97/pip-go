package pipgo

import (
	"encoding/binary"
	"syscall"
	"time"

	"github.com/plumk97/pip-go/lib/chainbuf"
	"github.com/plumk97/pip-go/types"
)

func tcpIsBeforeSeq(seq, ack uint32) bool {
	return int32(seq-ack) <= 0
}

func tcpIncreaseSeq(seq uint32, flags uint8, datalen uint16) uint32 {
	if datalen > 0 {
		return seq + uint32(datalen)
	}

	if flags&types.TH_SYN > 0 || flags&types.TH_FIN > 0 {
		return seq + 1
	}

	return seq
}

// 释放资源
func (tcp *TCP) release() {
	if tcp.status == TCPStatusReleased {
		return
	}
	tcp.status = TCPStatusReleased
	tcp.events = append(tcp.events, &tcpClosedEvent{
		arg: tcp.Arg,
	})
	tcp.Arg = nil
}

// 发送数据包
func (tcp *TCP) sendPacket(packet *TCPPacket) {
	packet.sended()
	hdr := packet.hdr()
	datalen := packet.payloadLen

	if tcp.ipHeader.Version == 4 {
		tcp.netif.output4(packet.headBuf, syscall.IPPROTO_TCP, tcp.ipHeader.Dst, tcp.ipHeader.Src)
	} else {
		tcp.netif.output6(packet.headBuf, syscall.IPPROTO_TCP, tcp.ipHeader.Dst, tcp.ipHeader.Src)
	}

	tcp.seq = tcpIncreaseSeq(tcp.seq, hdr.Flags(), uint16(datalen))
}

// 重新发送数据包
func (tcp *TCP) resendPacket(packet *TCPPacket) {
	packet.sended()
	if tcp.ipHeader.Version == 4 {
		tcp.netif.output4(packet.headBuf, syscall.IPPROTO_TCP, tcp.ipHeader.Dst, tcp.ipHeader.Src)
	} else {
		tcp.netif.output6(packet.headBuf, syscall.IPPROTO_TCP, tcp.ipHeader.Dst, tcp.ipHeader.Src)
	}
}

// 发送确认ACK
func (tcp *TCP) sendAck() {
	packet := newTCPPacket(tcp, types.TH_ACK, nil, nil)
	tcp.sendPacket(packet)
}

// 发送重置RST并关闭连接
func (tcp *TCP) sendReset() {
	packet := newTCPPacket(tcp, types.TH_RST|types.TH_ACK, nil, nil)
	tcp.sendPacket(packet)
	tcp.release()
}

// 处理建立连接
func (tcp *TCP) handleSyn(options []byte) {
	tcp.status = TCPStatusEstablishing

	if options != nil {

		optionLen := len(options)
		var offset int = 0
		for offset < optionLen {

			kind := options[offset]
			offset += 1

			if kind == 0 || kind == 1 {
				continue
			}

			len := options[offset]
			offset += 1

			var valueLen uint8 = 0
			if len > 2 {
				valueLen = len - 2
			}

			switch kind {
			case 2:
				// mss
				mss := binary.BigEndian.Uint16(options[offset : offset+int(valueLen)])
				tcp.oppMss = mss

			case 3:
				// wind shift
				shift := options[offset]
				tcp.oppWindShift = shift
			}

			offset += int(valueLen)
		}
	}

	optionBuf := chainbuf.NewChainBuffer(make([]byte, 8))
	offset := 0
	{
		// mss
		optionBuf.Payload()[offset] = 2   // kind
		optionBuf.Payload()[offset+1] = 4 // len

		value := make([]byte, 2)
		binary.BigEndian.PutUint16(value, tcp.mss)
		copy(optionBuf.Payload()[offset+2:offset+4], value) // value
		offset += 4
	}

	{
		// window scale
		optionBuf.Payload()[offset] = 3   // kind
		optionBuf.Payload()[offset+1] = 3 // len
		optionBuf.Payload()[offset+2] = 0 // value
		offset += 3
	}

	packet := newTCPPacket(tcp, types.TH_SYN|types.TH_ACK, optionBuf, nil)
	tcp.packetQueue.Push(packet)
	tcp.sendPacket(packet)
}

// 处理断开连接
func (tcp *TCP) handleFin() {

	switch tcp.status {
	case TCPStatusFinWait2:
		packet := newTCPPacket(tcp, types.TH_ACK, nil, nil)
		tcp.sendPacket(packet)
		tcp.release()

	case TCPStatusEstablished:
		tcp.status = TCPStatusCloseWait

		packet := newTCPPacket(tcp, types.TH_FIN|types.TH_ACK, nil, nil)
		tcp.packetQueue.Push(packet)
		tcp.sendPacket(packet)
	}

}

// 处理ACK确认
func (tcp *TCP) handleAck(ack uint32, isUpdateWind bool) {

	hasSyn := false
	hasFin := false
	hasPush := false
	writtenLen := 0

	for !tcp.packetQueue.Empty() {
		pkt := tcp.packetQueue.Front()
		hdr := pkt.hdr()

		seq := hdr.Seq() + uint32(pkt.payloadLen)
		if hdr == nil || !tcpIsBeforeSeq(seq, ack) {
			break
		}
		tcp.packetQueue.Pop()

		hasSyn = hdr.Flags()&types.TH_SYN == types.TH_SYN
		hasFin = hdr.Flags()&types.TH_FIN == types.TH_FIN

		if pkt.payloadLen > 0 {
			writtenLen += pkt.payloadLen

			if hdr.Flags()&types.TH_PUSH == types.TH_PUSH {
				hasPush = true
				tcp.isWaitPushAck = false
			}
		}
	}

	if hasSyn {
		tcp.status = TCPStatusEstablished
		tcp.events = append(tcp.events, &tcpConnectedEvent{})
	}

	if writtenLen > 0 || isUpdateWind {
		tcp.events = append(tcp.events, &tcpWrittenEvent{
			writtenLen: writtenLen,
			hasPush:    hasPush,
		})
	}

	if hasFin {
		switch tcp.status {
		case TCPStatusFinWait1:
			/// 主动关闭 改变状态
			tcp.status = TCPStatusFinWait2
			tcp.finTime = time.Now()

		case TCPStatusCloseWait:
			/// 被动关闭 清理资源
			tcp.release()
		}
	}

}

// 处理数据接收
func (tcp *TCP) handleReceive(data []byte) {
	tcp.wind -= uint16(len(data))
	tcp.events = append(tcp.events, &tcpReceivedEvent{
		data: data,
	})
}
