package pipgo

import (
	"net/netip"
	"syscall"

	"github.com/plumk97/pip-go/lib/chainbuf"
	"github.com/plumk97/pip-go/lib/checksum"
	"github.com/plumk97/pip-go/types"
)

func (n *Netif) udpInput(data []byte, ipHeader *types.IPHeader) {
	hdr := types.UDPHdr(data)
	data = data[8:]
	if n.OnUDPData != nil {
		n.OnUDPData(n, data, ipHeader.Src, hdr.SrcPort(), ipHeader.Dst, hdr.DstPort())
	}
}

func (n *Netif) UDPOutput(data []byte, srcIP netip.Addr, srcPort uint16, dstIP netip.Addr, dstPort uint16) {

	dataBuf := chainbuf.NewChainBuffer(data)
	udpHeadBuf := chainbuf.NewChainBuffer(types.NewUDPHdr())
	udpHeadBuf.SetNext(dataBuf)

	hdr := types.UDPHdr(udpHeadBuf.Payload())
	hdr.SetSrcPort(srcPort)
	hdr.SetDstPort(dstPort)
	hdr.SetLen(uint16(udpHeadBuf.TotalLen()))
	hdr.SetSum(checksum.InetChecksumBuf(udpHeadBuf, syscall.IPPROTO_UDP, srcIP, dstIP))
	if dstIP.Is4() {
		n.output4(udpHeadBuf, syscall.IPPROTO_UDP, srcIP, dstIP)
	} else {
		n.output6(udpHeadBuf, syscall.IPPROTO_UDP, srcIP, dstIP)
	}

}
