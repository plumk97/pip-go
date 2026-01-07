package pipgo

import (
	"github.com/plumk97/pip-go/types"
)

// Input
func (n *Netif) tcpInput(data []byte, ipHeader *types.IPHeader) {

	hdr := types.TCPHdr(data[:20])

	datalen := ipHeader.Datalen - uint16(hdr.Off())*4

	srcPort := hdr.SrcPort()
	dstPort := hdr.DstPort()

	key := TCPKey{
		SrcIP:   ipHeader.Src,
		DstIP:   ipHeader.Dst,
		SrcPort: srcPort,
		DstPort: dstPort,
	}

	// 查找对应的TCP连接
	n.locker.Lock()
	tcp, isOK := n.tcps[key]

	// 不存在的连接 如果是SYN包 则建立一个新的连接
	if !isOK && hdr.Flags()&types.TH_SYN > 0 {
		tcp = newTCP(n)
		tcp.key = key
		tcp.seq = 0
		tcp.ipHeader = ipHeader
		tcp.srcPort = srcPort
		tcp.dstPort = dstPort
		n.tcps[key] = tcp
	}
	n.locker.Unlock()

	// 不存在的连接 直接返回RST
	if tcp == nil {
		if hdr.Flags()&types.TH_RST <= 0 {
			// 不存在的连接 直接返回RST
			tcp = newTCP(n)
			tcp.key = key
			tcp.seq = 0
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

	tcp.input(hdr, data, data[20:], datalen)
}

// TCP 处理输入的TCP包
func (tcp *TCP) input(hdr types.TCPHdr, head []byte, data []byte, datalen uint16) {
	defer tcp.processEvents()

	tcp.locker.Lock()
	defer tcp.locker.Unlock()

	if tcp.status == TCPStatusReleased {
		// 连接已释放不处理
		return
	}

	if hdr.Flags()&types.TH_RST > 0 {
		// 处理RST包
		tcp.release()
		return
	}

	if hdr.Flags()&types.TH_ACK > 0 && hdr.Seq() == tcp.ack-1 {
		// keep-alive 包 直接回复
		tcp.sendAck()
		return
	}

	if hdr.Ack() > 0 && hdr.Seq() != tcp.ack {
		// 当前数据包seq与之前的ack对不上 产生了丢包 回复之前的ack 等待重传
		tcp.sendAck()
		return
	}

	// 更新对方的seq和ack
	tcp.oppSeq = hdr.Seq()
	tcp.ack = tcpIncreaseSeq(hdr.Seq(), hdr.Flags(), datalen)

	// 更新对方的窗口大小
	isUpdateWind := tcp.oppWind <= 0 && !tcp.isWaitPushAck
	tcp.oppWind = uint32(hdr.Win()) << uint32(tcp.oppWindShift)

	// 处理ACK和数据包
	if hdr.Flags()&types.TH_ACK > 0 {
		tcp.handleAck(hdr.Ack(), isUpdateWind)
	}

	// 处理收到的数据
	if hdr.Flags()&types.TH_PUSH > 0 || datalen > 0 {
		tcp.handleReceive(data)
	}

	if tcp.status == TCPStatusReleased {
		// 在handleAck里已经释放
		return
	}

	if hdr.Flags()&types.TH_SYN > 0 {
		tcp.status = TCPStatusWaitEstablishing
		tcp.events = append(tcp.events, &tcpNewEvent{
			head: head,
		})
	}

	// 处理FIN包
	if hdr.Flags()&types.TH_FIN > 0 {
		tcp.handleFin()
	}
}
