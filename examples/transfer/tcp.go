package transfer

import (
	"bytes"
	"errors"
	"io"
	"log"
	"net"

	pipgo "github.com/plumk97/pip-go"
)

// tcpBridge 负责在 pipgo TCP 与本地 TCPConn 之间转发数据
type tcpBridge struct {
	tcp  *pipgo.TCP
	conn *net.TCPConn

	buffer *bytes.Buffer
}

func newTCPBridge(tcp *pipgo.TCP, conn *net.TCPConn) *tcpBridge {
	b := &tcpBridge{tcp: tcp, conn: conn, buffer: &bytes.Buffer{}}
	b.bindCallbacks()
	return b
}

func (b *tcpBridge) bindCallbacks() {
	tcp := b.tcp
	conn := b.conn

	tcp.OnReceived = func(t *pipgo.TCP, data []byte) {
		conn.Write(data)
		t.Received(uint16(len(data)))
	}

	tcp.OnWritten = func(t *pipgo.TCP, writtenLen int, hasPush, isDrop bool) {
		if hasPush || writtenLen == 0 {
			b.handleBuffer()
		}
	}

	tcp.OnClosed = func(t *pipgo.TCP, arg any) {
		conn.Close()
	}

	tcp.OnConnected = func(t *pipgo.TCP) {
		go b.read()
	}
}

func (b *tcpBridge) handleBuffer() {
	for {
		if b.buffer.Len() == 0 {
			break
		}
		data := b.buffer.Bytes()
		n := b.tcp.Write(data, true)
		if n <= 0 {
			break
		}
		b.buffer.Next(n)
	}
}

func (b *tcpBridge) read() {
	defer b.tcp.Close()

	buf := make([]byte, 65535<<b.tcp.OppWindShift())
	for {
		n, err := b.conn.Read(buf)
		if n <= 0 || err != nil {
			if ne, ok := err.(net.Error); ok && ne.Timeout() {
				return
			}

			if errors.Is(err, io.EOF) {
				return
			}
			log.Println(err)
			return
		}

		b.buffer.Write(buf[:n])
		b.handleBuffer()
	}

}

func onTCPConnect(netif *pipgo.Netif, tcp *pipgo.TCP, handshakeData []byte) {
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
