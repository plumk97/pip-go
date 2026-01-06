package pipgo

import (
	"net"
	"syscall"

	"github.com/plumk97/pip-go/lib/chainbuf"
	"github.com/plumk97/pip-go/lib/checksum"
	"github.com/plumk97/pip-go/types"
)

func (n *Netif) udpInput(data []byte, ipHeader *types.IPHeader) {
	hdr := types.UDPHdr(data)
	data = data[8:]
	if n.ReceiveUDPData != nil {
		n.ReceiveUDPData(n, data, ipHeader.Src, hdr.SrcPort(), ipHeader.Dst, hdr.DstPort())
	}
}

func (n *Netif) UDPOutput(data []byte, srcIP net.IP, srcPort uint16, dstIP net.IP, dstPort uint16) {

	dataBuf := chainbuf.NewChainBuffer(data)
	udpHeadBuf := chainbuf.NewChainBuffer(types.NewUDPHdr())
	udpHeadBuf.SetNext(dataBuf)

	hdr := types.UDPHdr(udpHeadBuf.Payload())
	hdr.SetSrcPort(srcPort)
	hdr.SetDstPort(dstPort)
	hdr.SetLen(uint16(udpHeadBuf.TotalLen()))
	hdr.SetSum(checksum.InetChecksumBuf(udpHeadBuf, syscall.IPPROTO_UDP, srcIP, dstIP))
	if dstIP.To4() != nil {
		n.output4(udpHeadBuf, syscall.IPPROTO_UDP, srcIP, dstIP)
	} else {
		n.output6(udpHeadBuf, syscall.IPPROTO_UDP, srcIP, dstIP)
	}

}
