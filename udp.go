package pipgo

import (
	"net"
	"syscall"

	"github.com/plumk97/pip-go/types"
)

func udpInput(data []byte, ipHeader *IPHeader) {
	hdr := types.UDPHdr(data)
	data = data[8:]
	ReceiveUDPDataCallback(data, ipHeader.Src, hdr.SrcPort(), ipHeader.Dst, hdr.DstPort())
}

func UDPOutput(data []byte, srcIP net.IP, srcPort uint16, dstIP net.IP, dstPort uint16) {

	dataBuf := NewBuffer(data)
	udpHeadBuf := NewBuffer(types.NewUDPHdr())
	udpHeadBuf.SetNext(dataBuf)

	hdr := types.UDPHdr(udpHeadBuf.payload)
	hdr.SetSrcPort(srcPort)
	hdr.SetDstPort(dstPort)
	hdr.SetLen(uint16(udpHeadBuf.totalLen))
	hdr.SetSum(InetChecksumBuf(udpHeadBuf, syscall.IPPROTO_UDP, srcIP, dstIP))
	if dstIP.To4() != nil {
		output4(udpHeadBuf, syscall.IPPROTO_UDP, srcIP, dstIP)
	} else {
		output6(udpHeadBuf, syscall.IPPROTO_UDP, srcIP, dstIP)
	}

}
