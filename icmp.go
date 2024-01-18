package pipgo

import (
	"net"

	"github.com/plumk97/pip-go/types"
)

func icmpInput(data []byte, ipHeader *IPHeader) {
	if ReceiveICMPDataCallback != nil {
		ReceiveICMPDataCallback(data, ipHeader.Src, ipHeader.Dst, ipHeader.TTL)
	}
}

func ICMPOutput(data []byte, srcIP net.IP, dstIP net.IP) {

	dataBuf := NewBuffer(data)

	if dstIP.To4() != nil {
		output4(dataBuf, types.IPPROTO_ICMP, srcIP, dstIP)
	} else {
		output6(dataBuf, types.IPPROTO_ICMP, srcIP, dstIP)
	}

}
