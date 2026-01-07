package transfer

import (
	"log"
	"net"
	"net/netip"

	pipgo "github.com/plumk97/pip-go"
)

func onUDPData(netif *pipgo.Netif, data []byte, srcIP netip.Addr, srcPort uint16, dstIP netip.Addr, dstPort uint16) {

	go func() {
		conn, err := net.DialUDP("udp",
			&net.UDPAddr{IP: outboundIp},
			&net.UDPAddr{IP: dstIP.AsSlice(), Port: int(dstPort)})
		if err != nil {
			log.Println(err)
			return
		}
		defer conn.Close()

		conn.Write(data)

		bytes := [2048]byte{}
		n, err := conn.Read(bytes[:])
		if err != nil {
			log.Println(err)
			return
		}

		netif.UDPOutput(bytes[:n], dstIP, dstPort, srcIP, srcPort)
	}()
}
