package pipgo

import (
	"time"

	"github.com/plumk97/pip-go/types"
)

//	建立连接
//
// @param handshakeData 发起连接时的握手数据
func (tcp *TCP) Connected(handshakeData []byte) {
	tcp.mutex.Lock()
	defer tcp.mutex.Unlock()

	if tcp.status != TCPStatusWaitEstablishing {
		return
	}

	hdr := types.TCPHdr(handshakeData)
	if hdr.Off() > 5 {
		tcp.handleSyn(handshakeData[20:])
	} else {
		tcp.handleSyn(nil)
	}
}

// 关闭连接
func (tcp *TCP) Close() {
	tcp.mutex.Lock()
	defer tcp.mutex.Unlock()

	switch tcp.status {
	case TCPStatusWaitClosed:
		tcp.release(&tcp.mutex)

	case TCPStatusWaitEstablishing,
		TCPStatusEstablishing:
		tcp._reset()

	case TCPStatusEstablished:
		tcp.status = TCPStatusFinWait1
		tcp.finTime = time.Now()

		packet := newTCPPacket(tcp, types.TH_FIN|types.TH_ACK, nil, nil)
		tcp.packetQueue.Push(packet)
		tcp.sendPacket(packet)
	}

}

// 重置连接
func (tcp *TCP) Reset() {
	tcp.mutex.Lock()
	defer tcp.mutex.Unlock()
	tcp._reset()
}
func (tcp *TCP) _reset() {
	switch tcp.status {
	case TCPStatusWaitEstablishing,
		TCPStatusEstablishing,
		TCPStatusEstablished:
		packet := newTCPPacket(tcp, types.TH_RST|types.TH_ACK, nil, nil)
		tcp.sendPacket(packet)
	}
}

// 发送数据 返回发送的长度
// @param data 待发送数据
func (tcp *TCP) Write(data []byte) int {
	tcp.mutex.Lock()
	defer tcp.mutex.Unlock()

	if tcp.status != TCPStatusEstablished || !tcp._canWrite() {
		return 0
	}

	datalen := len(data)
	offset := 0
	for offset < len(data) && tcp.oppWind > 0 {

		writeLen := int(tcp.oppMss)

		/// 获取小于等于mss的数据长度
		if offset+writeLen > datalen {
			writeLen = datalen - offset
		}

		/// 获取小于等于对方的窗口长度
		if uint32(writeLen) > tcp.oppWind {
			writeLen = int(tcp.oppWind)
		}

		if writeLen <= 0 {
			break
		}

		/// 如果当前发送数据大于等于总数据长度 或者 对方窗口为0 则发送PUSH标签
		isPush := offset+writeLen >= datalen || writeLen >= int(tcp.oppWind)

		payloadBuf := NewBuffer(data[offset : offset+writeLen])

		var packet *TCPPacket
		if isPush {
			packet = newTCPPacket(tcp, types.TH_PUSH|types.TH_ACK, nil, payloadBuf)
			tcp.isWaitPushAck = true
		} else {
			packet = newTCPPacket(tcp, types.TH_ACK, nil, payloadBuf)
		}

		tcp.packetQueue.Push(packet)
		tcp.sendPacket(packet)

		offset += writeLen
		tcp.oppWind -= uint32(writeLen)
	}

	return offset
}

// 接受数据之后调用更新窗口
// @param len 接受的数据大小
func (tcp *TCP) Received(len uint16) {
	tcp.mutex.Lock()
	defer tcp.mutex.Unlock()

	if tcp.status != TCPStatusEstablished {
		return
	}

	tcp.wind += len
	if tcp.wind > _TCP_WIND {
		tcp.wind = _TCP_WIND
	}

	if tcp.ack-uint32(len) == tcp.oppSeq || tcp.wind-len <= 0 {
		tcp.sendAck()
	}

}

// 写之前调用该方法判断当前是否能写
func (tcp *TCP) CanWrite() bool {
	tcp.mutex.Lock()
	defer tcp.mutex.Unlock()
	return tcp._canWrite()
}

func (tcp *TCP) _canWrite() bool {
	return !tcp.isWaitPushAck
}
