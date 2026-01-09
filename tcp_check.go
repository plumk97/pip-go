package pipgo

import (
	"maps"
	"time"

	"github.com/plumk97/pip-go/types"
)

func (n *Netif) startTCPTimer() {
	go func() {
		ticker := time.NewTicker(250 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				n.tcpTimerTick()
			case <-n.stopTimer:
				return
			}
		}
	}()
}

func (n *Netif) tcpTimerTick() {

	// 复制一份连接列表 避免锁住太久
	n.locker.Lock()
	if len(n.tcps) <= 0 {
		n.locker.Unlock()
		return
	}

	tcps := map[TCPKey]*TCP{}
	maps.Copy(tcps, n.tcps)
	n.locker.Unlock()

	curTime := time.Now()
	for _, tcp := range tcps {
		n.tcpCheck(tcp, curTime)
		tcp.processEvents()
	}
}

func (n *Netif) tcpCheck(tcp *TCP, curTime time.Time) bool {
	tcp.locker.Lock()
	defer tcp.locker.Unlock()

	if tcp.status == TCPStatusReleased {
		return true
	}

	if (tcp.status == TCPStatusFinWait1 || tcp.status == TCPStatusFinWait2 || tcp.status == TCPStatusCloseWait) && (curTime.Sub(tcp.finTime) > 20*time.Second) {
		// 处于等待关闭状态 并且等待时间已经大于20秒 直接关闭
		tcp.release()
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

					tcp.events = append(tcp.events, &tcpWrittenEvent{
						writtenLen: packet.payloadLen,
						hasPush:    hasPush,
						isDrop:     true,
					})
				}

			} else {
				// 小于2次的重发
				tcp.resendPacket(packet)
			}
		}
	}

	return false
}
