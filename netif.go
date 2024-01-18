package pipgo

import (
	"net"

	"github.com/plumk97/pip-go/types"
)

// 输出IP包数据
// @param buf IP包数据
var OutputIPDataCallback func(buf *Buffer)

// 接受到一个新的TCP连接
// @param tcp
// @param handshakeData 建立连接的握手数据 连接成功调用 connected 方法需要传入
var NewTCPConnectCallback func(tcp *TCP, handshakeData []byte)

// 接受到UDP数据
// @param data 数据
// @param srcIP 来源地址
// @param srcPort 来源端口
// @param dstIP 目的地址
// @param dstPort 目的端口
var ReceiveUDPDataCallback func(data []byte, srcIP net.IP, srcPort uint16, dstIP net.IP, dstPort uint16)

// 接受到ICMP数据
// @param srcIP 来源地址
// @param dstIP 目的地址
// @param ttl
var ReceiveICMPDataCallback func(data []byte, srcIP, dstIP net.IP, ttl uint8)

var identifer uint16 = 0

// 输入IP包数据
func Input(bytes []byte) {
	ipHeader := NewIPHeader(bytes)
	if ipHeader.Version == 4 {
		// 检测是否有options 不支持options
		if ipHeader.HasOptions {
			return
		}
	}

	data := bytes[ipHeader.Headerlen:]
	switch ipHeader.Protocol {
	case types.IPPROTO_ICMP:
		icmpInput(data, ipHeader)

	case types.IPPROTO_TCP:
		tcpInput(data, ipHeader)

	case types.IPPROTO_UDP:
		udpInput(data, ipHeader)
	}
}

func output4(buf *Buffer, proto uint8, src, dst net.IP) {

	ipHeadBuf := NewBuffer(types.NewIPHdr())
	ipHeadBuf.SetNext(buf)

	hdr := types.IPHdr(ipHeadBuf.payload)
	hdr.SetVersion(4)
	hdr.SetIHL(5)
	hdr.SetTos(0)
	hdr.SetLen(uint16(ipHeadBuf.totalLen))
	hdr.SetID(identifer)
	hdr.SetOff(0x4000) // dont fragment flag
	hdr.SetTTL(64)
	hdr.SetProtocol(proto)
	hdr.SetSrc(src)
	hdr.SetDst(dst)
	hdr.SetSum(IPChecksum(hdr))

	if OutputIPDataCallback != nil {
		OutputIPDataCallback(ipHeadBuf)
	}

	ipHeadBuf.SetNext(nil)
	identifer += 1
}

func output6(buf *Buffer, proto uint8, src, dst net.IP) {
	ipHeadBuf := NewBuffer(types.NewIP6Hdr())
	ipHeadBuf.SetNext(buf)

	hdr := types.IP6Hdr(ipHeadBuf.payload)
	hdr.SetVersion(6)
	hdr.SetTrafficClass(0)
	hdr.SetFlow(0)
	hdr.SetPayloadLen(uint16(buf.totalLen))
	hdr.SetNextHeader(proto)
	hdr.SetHopLimit(64)
	hdr.SetSrc(src)
	hdr.SetDst(dst)

	if OutputIPDataCallback != nil {
		OutputIPDataCallback(ipHeadBuf)
	}

	ipHeadBuf.SetNext(nil)
}
