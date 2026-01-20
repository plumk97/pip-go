package pipgo

type tcpNewEvent struct {
	head []byte
}

type tcpConnectedEvent struct {
}

type tcpWrittenEvent struct {
	writtenLen int
	hasPush    bool
}

type tcpReceivedEvent struct {
	data []byte
}

type tcpClosedEvent struct {
	arg any
}

// 处理事件
func (tcp *TCP) processEvents() {
	tcp.locker.Lock()
	events := tcp.events
	tcp.events = nil
	tcp.locker.Unlock()

	for _, event := range events {
		switch ev := event.(type) {
		case *tcpNewEvent:
			if tcp.netif.OnTCPConnect != nil {
				tcp.netif.OnTCPConnect(tcp.netif, tcp, ev.head)
			}

		case *tcpConnectedEvent:
			if tcp.OnConnected != nil {
				tcp.OnConnected(tcp)
			}

		case *tcpWrittenEvent:
			if tcp.OnWritten != nil {
				tcp.OnWritten(tcp, ev.writtenLen, ev.hasPush)
			}

		case *tcpReceivedEvent:
			if tcp.OnReceived != nil {
				tcp.OnReceived(tcp, ev.data)
			}

		case *tcpClosedEvent:
			// 删除连接
			tcp.netif.locker.Lock()
			delete(tcp.netif.tcps, tcp.key)
			tcp.netif.locker.Unlock()

			if tcp.OnClosed != nil {
				tcp.OnClosed(tcp, ev.arg)
			}
		}
	}
}
