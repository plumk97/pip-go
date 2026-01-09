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
	tcp, exists := n.tcps[key]

	if !exists {
		if hdr.Flags()&types.TH_SYN == types.TH_SYN {
			// 不存在的连接 如果是SYN包 则建立一个新的连接
			tcp = newTCP(n)
			tcp.key = key
			tcp.seq = 0
			tcp.ipHeader = ipHeader
			tcp.srcPort = srcPort
			tcp.dstPort = dstPort
			n.tcps[key] = tcp

		} else {
			// 不存在的连接 并且不是SYN包 直接返回RST
			tmpTcp := newTCP(n)
			tmpTcp.key = key
			tmpTcp.seq = 0
			tmpTcp.ipHeader = ipHeader
			tmpTcp.srcPort = srcPort
			tmpTcp.dstPort = dstPort
			tmpTcp.seq = hdr.Ack()
			tmpTcp.ack = tcpIncreaseSeq(hdr.Seq(), hdr.Flags(), datalen)
			tmpTcp.sendReset()
		}

	}
	n.locker.Unlock()

	if tcp == nil {
		return
	}

	tcp.input(hdr, data, data[20:], datalen)
	tcp.processEvents()
}

// TCP 处理输入的TCP包
func (tcp *TCP) input(hdr types.TCPHdr, head []byte, data []byte, datalen uint16) {

	tcp.locker.Lock()
	defer tcp.locker.Unlock()

	if tcp.status == TCPStatusReleased {
		// 连接已释放不处理
		return
	}

	if hdr.Flags()&types.TH_RST == types.TH_RST {
		// 处理RST包
		tcp.release()
		return
	}

	if hdr.Flags()&types.TH_ACK == types.TH_ACK && hdr.Seq() == tcp.ack-1 {
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
	if hdr.Flags()&types.TH_ACK == types.TH_ACK {
		tcp.handleAck(hdr.Ack(), isUpdateWind)
	}

	// 处理收到的数据
	if hdr.Flags()&types.TH_PUSH == types.TH_PUSH || datalen > 0 {
		tcp.handleReceive(data)
	}

	if tcp.status == TCPStatusReleased {
		// 在handleAck里已经释放
		return
	}

	if hdr.Flags()&types.TH_SYN == types.TH_SYN {

		switch tcp.status {
		case TCPStatusNone:
			// 处理新的SYN包 建立连接
			tcp.status = TCPStatusWaitEstablishing
			tcp.events = append(tcp.events, &tcpNewEvent{
				head: head,
			})

		case TCPStatusWaitEstablishing, TCPStatusEstablishing:
			// 正在建立连接 收到重复的SYN包 忽略

		case TCPStatusEstablished:
			// 已经建立连接 收到重复的SYN包 重新发送SYN-ACK
			tcp.handleSyn(nil)

		default:
			// 其他状态下 收到SYN包 直接回复RST
			tcp.sendReset()

		}
	}

	// 处理FIN包
	if hdr.Flags()&types.TH_FIN == types.TH_FIN {
		tcp.handleFin()
	}
}
