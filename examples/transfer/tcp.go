package transfer

import (
	"log"
	"net"

	pipgo "github.com/plumk97/pip-go"
)

// tcpBridge 负责在 pipgo TCP 与本地 TCPConn 之间转发数据
type tcpBridge struct {
	tcp  *pipgo.TCP
	conn *net.TCPConn
}

func newTCPBridge(tcp *pipgo.TCP, conn *net.TCPConn) *tcpBridge {
	b := &tcpBridge{tcp: tcp, conn: conn}
	b.bindCallbacks()
	return b
}

func (b *tcpBridge) bindCallbacks() {
	tcp := b.tcp
	conn := b.conn

	tcp.ReceivedCallback = func(t *pipgo.TCP, data []byte) {
		conn.Write(data)
		t.Received(uint16(len(data)))
	}

	tcp.WrittenCallback = func(t *pipgo.TCP, writtenLen int, hasPush, isDrop bool) {
		if hasPush || writtenLen == 0 {
			go b.read()
		}
	}

	tcp.ClosedCallback = func(t *pipgo.TCP, arg any) {
		log.Println("TCP: closed")
		conn.Close()
	}

	tcp.ConnectedCallback = func(t *pipgo.TCP) {
		log.Println("TCP: connected")
		go b.read()
	}
}

func (b *tcpBridge) read() {
	buf := make([]byte, 65535<<b.tcp.OppWindShift())
	n, err := b.conn.Read(buf)
	if n <= 0 || err != nil {
		log.Println(err)
		b.tcp.Close()
		return
	}

	b.tcp.Write(buf[:n])
}

func newTCPConnectCallback(netif *pipgo.Netif, tcp *pipgo.TCP, handshakeData []byte) {
	log.Println("TCP: new connection", tcp.IPHeader().Src.String(), ":", tcp.SrcPort(), "->", tcp.IPHeader().Dst.String(), ":", tcp.DstPort())
	conn, err := net.DialTCP("tcp", &net.TCPAddr{
		// IP: net.IP{127, 0, 0, 1},
		IP: outboundIp,
	}, &net.TCPAddr{
		// IP:   tcp.IPHeader().Dst,
		IP:   net.IP{127, 0, 0, 1},
		Port: int(tcp.DstPort()),
	})

	if err != nil {
		tcp.Close()
		log.Println(err)
		return
	}

	bridge := newTCPBridge(tcp, conn)
	bridge.tcp.Connected(handshakeData)
}
