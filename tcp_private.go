package pipgo

import (
	"encoding/binary"
	"sync"
	"syscall"
	"time"

	"github.com/plumk97/pip-go/types"
)

var tcps map[uint32]*TCP = make(map[uint32]*TCP)
var tcpsLock sync.Mutex

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
func (tcp *TCP) release(mutex *sync.Mutex) {
	if tcp.status == TCPStatusReleased {
		return
	}
	tcp.status = TCPStatusReleased

	arg := tcp.Arg
	tcp.Arg = nil

	if tcp.ClosedCallback != nil {

		if mutex != nil {
			mutex.Unlock()
		}
		tcp.ClosedCallback(tcp, arg)
		if mutex != nil {
			mutex.Lock()
		}

		tcp.ClosedCallback = nil
	}
}

// 发送数据包
func (tcp *TCP) sendPacket(packet *TCPPacket) {
	packet.sended()
	hdr := packet.hdr()
	datalen := packet.payloadLen

	if tcp.ipHeader.Version == 4 {
		output4(packet.headBuf, syscall.IPPROTO_TCP, tcp.ipHeader.Dst, tcp.ipHeader.Src)
	} else {
		output6(packet.headBuf, syscall.IPPROTO_TCP, tcp.ipHeader.Dst, tcp.ipHeader.Src)
	}

	tcp.seq = tcpIncreaseSeq(tcp.seq, hdr.Flags(), uint16(datalen))
}

// 重新发送数据包
func (tcp *TCP) resendPacket(packet *TCPPacket) {
	packet.sended()
	if tcp.ipHeader.Version == 4 {
		output4(packet.headBuf, syscall.IPPROTO_TCP, tcp.ipHeader.Dst, tcp.ipHeader.Src)
	} else {
		output6(packet.headBuf, syscall.IPPROTO_TCP, tcp.ipHeader.Dst, tcp.ipHeader.Src)
	}
}

// 发送确认ACK
func (tcp *TCP) sendAck() {
	packet := newTCPPacket(tcp, types.TH_ACK, nil, nil)
	tcp.sendPacket(packet)
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

	optionBuf := NewBuffer(make([]byte, 8))
	offset := 0
	{
		// mss
		optionBuf.payload[offset] = 2   // kind
		optionBuf.payload[offset+1] = 4 // len

		value := make([]byte, 2)
		binary.BigEndian.PutUint16(value, tcp.mss)
		copy(optionBuf.payload[offset+2:offset+4], value) // value

		offset += 4
	}

	{
		// window scale
		optionBuf.payload[offset] = 3   // kind
		optionBuf.payload[offset+1] = 3 // len
		optionBuf.payload[offset+2] = 0 // value
		offset += 3
	}

	packet := newTCPPacket(tcp, types.TH_SYN|types.TH_ACK, optionBuf, nil)
	tcp.packetQueue.Push(packet)
	tcp.sendPacket(packet)
}

// 处理断开连接
func (tcp *TCP) handleFin(mutex *sync.Mutex) {

	switch tcp.status {
	case TCPStatusFinWait2:
		packet := newTCPPacket(tcp, types.TH_ACK, nil, nil)
		tcp.sendPacket(packet)
		tcp.release(mutex)

	case TCPStatusEstablished:
		tcp.status = TCPStatusCloseWait

		packet := newTCPPacket(tcp, types.TH_FIN|types.TH_ACK, nil, nil)
		tcp.packetQueue.Push(packet)
		tcp.sendPacket(packet)
	}

}

// 处理ACK确认
func (tcp *TCP) handleAck(ack uint32, isUpdateWind bool, mutex *sync.Mutex) {

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

		hasSyn = hdr.Flags()&types.TH_SYN > 0
		hasFin = hdr.Flags()&types.TH_FIN > 0

		if pkt.payloadLen > 0 {
			writtenLen += pkt.payloadLen

			if hdr.Flags()&types.TH_PUSH > 0 {
				hasPush = true
				tcp.isWaitPushAck = false
			}
		}
	}

	if hasSyn {
		tcp.status = TCPStatusEstablished
		if tcp.ConnectedCallback != nil {
			mutex.Unlock()
			tcp.ConnectedCallback(tcp)
			mutex.Lock()
		}
	}

	if writtenLen > 0 || isUpdateWind {
		if tcp.WrittenCallback != nil {
			mutex.Unlock()
			tcp.WrittenCallback(tcp, writtenLen, hasPush, false)
			mutex.Lock()
		}
	}

	if hasFin {
		if tcp.status == TCPStatusFinWait1 {
			/// 主动关闭 改变状态
			tcp.status = TCPStatusFinWait2
			tcp.finTime = time.Now()

		} else if tcp.status == TCPStatusCloseWait {
			/// 被动关闭 清理资源
			tcp.release(mutex)
		}
	}

}

// 处理数据接收
func (tcp *TCP) handleReceive(data []byte, mutex *sync.Mutex) {
	tcp.wind -= uint16(len(data))

	if tcp.ReceivedCallback != nil {
		mutex.Unlock()
		tcp.ReceivedCallback(tcp, data)
		mutex.Lock()
	}
}

// Input
func tcpInput(data []byte, ipHeader *IPHeader) {

	hdr := types.TCPHdr(data[:20])

	datalen := ipHeader.Datalen - uint16(hdr.Off())*4

	srcPort := hdr.SrcPort()
	dstPort := hdr.DstPort()

	iden := ipHeader.GenerateIden() ^ uint32(srcPort) ^ uint32(dstPort)

	tcpsLock.Lock()
	defer tcpsLock.Unlock()

	tcp, isOK := tcps[iden]
	if !isOK && hdr.Flags()&types.TH_SYN > 0 {
		tcp = newTCP()
		tcp.iden = iden
		tcp.seq = iden
		tcp.ipHeader = ipHeader
		tcp.srcPort = srcPort
		tcp.dstPort = dstPort
		tcps[iden] = tcp
	}

	if tcp == nil {
		if hdr.Flags()&types.TH_RST <= 0 {
			// 不存在的连接 直接返回RST
			tcp = newTCP()
			tcp.iden = iden
			tcp.seq = iden
			tcp.ipHeader = ipHeader
			tcp.srcPort = srcPort
			tcp.dstPort = dstPort
			tcp.seq = hdr.Ack()
			tcp.ack = tcpIncreaseSeq(hdr.Seq(), hdr.Flags(), datalen)

			packet := newTCPPacket(tcp, types.TH_RST|types.TH_ACK, nil, nil)
			tcp.sendPacket(packet)
		}
		return
	}

	tcp.mutex.Lock()
	defer tcp.mutex.Unlock()

	if tcp.status == TCPStatusReleased {
		return
	}

	if hdr.Flags()&types.TH_RST > 0 {
		// RST 标志直接释放
		tcp.release(&tcp.mutex)
		return
	}

	if hdr.Flags()&types.TH_ACK > 0 && hdr.Seq() == tcp.ack-1 {
		// keep-alive 包 直接回复
		tcp.sendAck()
		return
	}

	if hdr.Ack() > 0 {
		if hdr.Seq() != tcp.ack {
			// 当前数据包seq与之前的ack对不上 产生了丢包 回复之前的ack 等待重传
			tcp.sendAck()
			return
		}
	}

	tcp.oppSeq = hdr.Seq()
	tcp.ack = tcpIncreaseSeq(hdr.Seq(), hdr.Flags(), datalen)

	isUpdateWind := tcp.oppWind <= 0 && !tcp.isWaitPushAck
	tcp.oppWind = uint32(hdr.Win()) << uint32(tcp.oppWindShift)

	if hdr.Flags()&types.TH_PUSH > 0 || datalen > 0 {
		tcp.handleReceive(data[20:], &tcp.mutex)
	}

	if hdr.Flags()&types.TH_ACK > 0 {
		tcp.handleAck(hdr.Ack(), isUpdateWind, &tcp.mutex)
	}

	if tcp.status == TCPStatusReleased {
		// 在handleAck里已经释放
		return
	}

	if hdr.Flags()&types.TH_SYN > 0 {
		tcp.status = TCPStatusWaitEstablishing
		if NewTCPConnectCallback != nil {
			tcp.mutex.Unlock()
			NewTCPConnectCallback(tcp, data)
			tcp.mutex.Lock()
		}
	}

	if hdr.Flags()&types.TH_FIN > 0 {
		tcp.handleFin(&tcp.mutex)
	}
}

func tcpTimerTick() {
	tcpsLock.Lock()
	defer tcpsLock.Unlock()

	curTime := time.Now()
	if len(tcps) <= 0 {
		return
	}

	for key, tcp := range tcps {

		isRemove := tcpCheck(tcp, curTime)
		if isRemove {
			delete(tcps, key)
		}
	}
}

func tcpCheck(tcp *TCP, curTime time.Time) bool {
	tcp.mutex.Lock()
	defer tcp.mutex.Unlock()

	if tcp.status == TCPStatusReleased {
		return true
	}

	if (tcp.status == TCPStatusFinWait1 || tcp.status == TCPStatusFinWait2 || tcp.status == TCPStatusCloseWait) && (curTime.Sub(tcp.finTime) > 20*time.Second) {
		// 处于等待关闭状态 并且等待时间已经大于20秒 直接关闭
		tcp.release(&tcp.mutex)
		return true
	}

	if tcp.packetQueue.Empty() {
		return false
	}

	packet := tcp.packetQueue.Front()
	if packet != nil {
		if curTime.Sub(packet.sendTime) >= 2*time.Second {
			// 数据超过2秒没有确认

			if packet.sendCount > 2 {
				// 已经发送过2次的直接丢弃
				tcp.packetQueue.Pop()

				if packet.payloadLen > 0 {
					hasPush := packet.hdr().Flags()&types.TH_PUSH > 0
					if hasPush {
						tcp.isWaitPushAck = false
					}

					if tcp.WrittenCallback != nil {
						tcp.mutex.Unlock()
						tcp.WrittenCallback(tcp, packet.payloadLen, hasPush, true)
						tcp.mutex.Lock()
					}
				}

			} else {
				// 小于2次的重发
				tcp.resendPacket(packet)
			}
		}
	}

	return false
}

func init() {
	go func() {
		ticker := time.NewTicker(250 * time.Millisecond)
		defer ticker.Stop()

		for range ticker.C {
			tcpTimerTick()
		}
	}()
}
