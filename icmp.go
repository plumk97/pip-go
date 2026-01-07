package pipgo

import (
	"net/netip"

	"github.com/plumk97/pip-go/lib/chainbuf"
	"github.com/plumk97/pip-go/types"
)

func (n *Netif) icmpInput(data []byte, ipHeader *types.IPHeader) {
	if n.OnICMPData != nil {
		n.OnICMPData(n, data, ipHeader.Src, ipHeader.Dst, ipHeader.TTL)
	}
}

func (n *Netif) ICMPOutput(data []byte, srcIP, dstIP netip.Addr) {

	dataBuf := chainbuf.NewChainBuffer(data)

	if dstIP.Is4() {
		n.output4(dataBuf, types.IPPROTO_ICMP, srcIP, dstIP)
	} else {
		n.output6(dataBuf, types.IPPROTO_ICMP, srcIP, dstIP)
	}

}
