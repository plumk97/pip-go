package pipgo

import (
	"net"
	"sync"

	"github.com/plumk97/pip-go/lib/chainbuf"
	"github.com/plumk97/pip-go/lib/checksum"
	"github.com/plumk97/pip-go/types"
)

// OutputIPDataCallback 输出IP包数据
type OutputIPDataCallback func(netif *Netif, buf *chainbuf.ChainBuffer)

// NewTCPConnectCallback 接受到一个新的TCP连接
type NewTCPConnectCallback func(netif *Netif, tcp *TCP, handshakeData []byte)

// ReceiveUDPDataCallback 接受到UDP数据
type ReceiveUDPDataCallback func(netif *Netif, data []byte, srcIP net.IP, srcPort uint16, dstIP net.IP, dstPort uint16)

// ReceiveICMPDataCallback 接受到ICMP数据
type ReceiveICMPDataCallback func(netif *Netif, data []byte, srcIP, dstIP net.IP, ttl uint8)

// MTU 最大传输单元
const MTU uint16 = 9000

// Netif 网络接口
type Netif struct {
	OutputIPData    OutputIPDataCallback    // 输出IP包数据
	NewTCPConnect   NewTCPConnectCallback   // 接受到一个新的TCP连接
	ReceiveUDPData  ReceiveUDPDataCallback  // 接收到UDP数据
	ReceiveICMPData ReceiveICMPDataCallback // 接收到ICMP数据

	locker    sync.Mutex
	identifer uint16          // IP包标识符
	tcps      map[uint32]*TCP // 已建立的TCP连接
	stopTimer chan struct{}   // 停止定时器
}

// NewNetif 创建一个网络接口
func NewNetif() *Netif {
	netif := &Netif{
		tcps:      make(map[uint32]*TCP),
		identifer: 0,
		stopTimer: make(chan struct{}),
	}

	netif.startTCPTimer()
	return netif
}

// Close 关闭网络接口
func (n *Netif) Close() {
	close(n.stopTimer)
}

// 输入IP包数据
func (n *Netif) Input(bytes []byte) {
	ipHeader := types.NewIPHeader(bytes)
	if ipHeader.Version == 4 {
		// 检测是否有options 不支持options
		if ipHeader.HasOptions {
			return
		}
	}

	data := bytes[ipHeader.Headerlen:]
	switch ipHeader.Protocol {
	case types.IPPROTO_ICMP:
		n.icmpInput(data, ipHeader)

	case types.IPPROTO_TCP:
		n.tcpInput(data, ipHeader)

	case types.IPPROTO_UDP:
		n.udpInput(data, ipHeader)
	}
}

// 输出IP包数据 IPv4
func (n *Netif) output4(buf *chainbuf.ChainBuffer, proto uint8, src, dst net.IP) {

	n.locker.Lock()
	identifer := n.identifer
	n.identifer += 1
	n.locker.Unlock()

	ipHeadBuf := chainbuf.NewChainBuffer(types.NewIPHdr())
	ipHeadBuf.SetNext(buf)

	hdr := types.IPHdr(ipHeadBuf.Payload())
	hdr.SetVersion(4)
	hdr.SetIHL(5)
	hdr.SetTos(0)
	hdr.SetLen(uint16(ipHeadBuf.TotalLen()))
	hdr.SetID(identifer)
	hdr.SetOff(0x4000) // dont fragment flag
	hdr.SetTTL(64)
	hdr.SetProtocol(proto)
	hdr.SetSrc(src)
	hdr.SetDst(dst)
	hdr.SetSum(checksum.IPChecksum(hdr))

	if n.OutputIPData != nil {
		n.OutputIPData(n, ipHeadBuf)
	}

	ipHeadBuf.SetNext(nil)
}

// 输出IP包数据 IPv6
func (n *Netif) output6(buf *chainbuf.ChainBuffer, proto uint8, src, dst net.IP) {
	ipHeadBuf := chainbuf.NewChainBuffer(types.NewIP6Hdr())
	ipHeadBuf.SetNext(buf)

	hdr := types.IP6Hdr(ipHeadBuf.Payload())
	hdr.SetVersion(6)
	hdr.SetTrafficClass(0)
	hdr.SetFlow(0)
	hdr.SetPayloadLen(uint16(buf.TotalLen()))
	hdr.SetNextHeader(proto)
	hdr.SetHopLimit(64)
	hdr.SetSrc(src)
	hdr.SetDst(dst)

	if n.OutputIPData != nil {
		n.OutputIPData(n, ipHeadBuf)
	}

	ipHeadBuf.SetNext(nil)
}
