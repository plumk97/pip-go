package pipgo

type tcpNewEvent struct {
	head []byte
}

type tcpConnectedEvent struct {
}

type tcpWrittenEvent struct {
	writtenLen int
	hasPush    bool
	isDrop     bool
}

type tcpReceivedEvent struct {
	data []byte
}

type tcpClosedEvent struct {
	arg any
}

// 处理事件
func (tcp *TCP) processEvents() {
	events := tcp.events
	tcp.events = nil

	for _, event := range events {
		switch ev := event.(type) {
		case *tcpNewEvent:
			if tcp.netif.NewTCPConnect != nil {
				tcp.netif.NewTCPConnect(tcp.netif, tcp, ev.head)
			}

		case *tcpConnectedEvent:
			if tcp.ConnectedCallback != nil {
				tcp.ConnectedCallback(tcp)
			}

		case *tcpWrittenEvent:
			if tcp.WrittenCallback != nil {
				tcp.WrittenCallback(tcp, ev.writtenLen, ev.hasPush, ev.isDrop)
			}

		case *tcpReceivedEvent:
			if tcp.ReceivedCallback != nil {
				tcp.ReceivedCallback(tcp, ev.data)
			}

		case *tcpClosedEvent:
			if tcp.ClosedCallback != nil {
				tcp.ClosedCallback(tcp, ev.arg)
			}
		}
	}
}
