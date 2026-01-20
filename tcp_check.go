package pipgo

import (
	"maps"
	"time"
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

func (n *Netif) tcpCheck(tcp *TCP, curTime time.Time) {
	tcp.locker.Lock()
	defer tcp.locker.Unlock()

	if tcp.status == TCPStatusReleased {
		return
	}

	if (tcp.status == TCPStatusFinWait1 || tcp.status == TCPStatusFinWait2 || tcp.status == TCPStatusCloseWait) && (curTime.Sub(tcp.finTime) > 20*time.Second) {
		// 处于等待关闭状态 并且等待时间已经大于20秒 直接关闭
		tcp.release()
		return
	}

	if tcp.packetQueue.Empty() {
		return
	}

	packet := tcp.packetQueue.Front()
	if curTime.Sub(packet.sendTime) >= 1*time.Second {
		// 1 秒等待确认
		return
	}

	if packet.sendCount > 5 {
		// 重传超过5次 认为连接已经断开 发送RST报文
		tcp.sendReset()
	} else {
		// 重传数据包
		tcp.resendPacket(packet)
	}

}
