package pipgo

import (
	"net/netip"
	"sync"
	"sync/atomic"

	"github.com/plumk97/pip-go/lib/chainbuf"
	"github.com/plumk97/pip-go/lib/checksum"
	"github.com/plumk97/pip-go/types"
)

// OnIPData 网络接口输出IP包数据回调函数
type OnIPData func(netif *Netif, buf *chainbuf.ChainBuffer)

// OnTCPConnect 接受到一个新的TCP连接
type OnTCPConnect func(netif *Netif, tcp *TCP, handshakeData []byte)

// OnUDPData 接受到UDP数据
type OnUDPData func(netif *Netif, data []byte, srcIP netip.Addr, srcPort uint16, dstIP netip.Addr, dstPort uint16)

// OnICMPData 接受到ICMP数据
type OnICMPData func(netif *Netif, data []byte, srcIP, dstIP netip.Addr, ttl uint8)

// MTU 最大传输单元
const MTU uint16 = 9000

// Netif 网络接口
type Netif struct {
	OnIPData     OnIPData     // 输出IP包数据
	OnTCPConnect OnTCPConnect // 接受到一个新的TCP连接
	OnUDPData    OnUDPData    // 接收到UDP数据
	OnICMPData   OnICMPData   // 接收到ICMP数据

	locker    sync.Mutex
	identifer uint32          // IP包标识符
	tcps      map[TCPKey]*TCP // 已建立的TCP连接
	stopTimer chan struct{}   // 停止定时器
}

// NewNetif 创建一个网络接口
func NewNetif() *Netif {
	netif := &Netif{
		tcps:      make(map[TCPKey]*TCP),
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
func (n *Netif) output4(buf *chainbuf.ChainBuffer, proto uint8, src, dst netip.Addr) {

	identifer := atomic.AddUint32(&n.identifer, 1)
	folded := identifer ^ (identifer >> 16)

	ipHeadBuf := chainbuf.NewChainBuffer(types.NewIPHdr())
	ipHeadBuf.SetNext(buf)

	hdr := types.IPHdr(ipHeadBuf.Payload())
	hdr.SetVersion(4)
	hdr.SetIHL(5)
	hdr.SetTos(0)
	hdr.SetLen(uint16(ipHeadBuf.TotalLen()))
	hdr.SetID(uint16(folded))
	hdr.SetOff(0x4000) // dont fragment flag
	hdr.SetTTL(64)
	hdr.SetProtocol(proto)
	hdr.SetSrc(src.AsSlice())
	hdr.SetDst(dst.AsSlice())
	hdr.SetSum(checksum.IPChecksum(hdr))

	if n.OnIPData != nil {
		n.OnIPData(n, ipHeadBuf)
	}

	ipHeadBuf.SetNext(nil)
}

// 输出IP包数据 IPv6
func (n *Netif) output6(buf *chainbuf.ChainBuffer, proto uint8, src, dst netip.Addr) {
	ipHeadBuf := chainbuf.NewChainBuffer(types.NewIP6Hdr())
	ipHeadBuf.SetNext(buf)

	hdr := types.IP6Hdr(ipHeadBuf.Payload())
	hdr.SetVersion(6)
	hdr.SetTrafficClass(0)
	hdr.SetFlow(0)
	hdr.SetPayloadLen(uint16(buf.TotalLen()))
	hdr.SetNextHeader(proto)
	hdr.SetHopLimit(64)
	hdr.SetSrc(src.AsSlice())
	hdr.SetDst(dst.AsSlice())

	if n.OnIPData != nil {
		n.OnIPData(n, ipHeadBuf)
	}

	ipHeadBuf.SetNext(nil)
}
