package pipgo

import (
	"net"

	"github.com/plumk97/pip-go/lib/chainbuf"
	"github.com/plumk97/pip-go/types"
)

func (n *Netif) icmpInput(data []byte, ipHeader *types.IPHeader) {
	if n.ReceiveICMPData != nil {
		n.ReceiveICMPData(n, data, ipHeader.Src, ipHeader.Dst, ipHeader.TTL)
	}
}

func (n *Netif) ICMPOutput(data []byte, srcIP net.IP, dstIP net.IP) {

	dataBuf := chainbuf.NewChainBuffer(data)

	if dstIP.To4() != nil {
		n.output4(dataBuf, types.IPPROTO_ICMP, srcIP, dstIP)
	} else {
		n.output6(dataBuf, types.IPPROTO_ICMP, srcIP, dstIP)
	}

}
