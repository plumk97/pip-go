package transfer

import (
	"log"
	"net"

	pipgo "github.com/plumk97/pip-go"
)

func receiveUDPDataCallback(netif *pipgo.Netif, data []byte, srcIP net.IP, srcPort uint16, dstIP net.IP, dstPort uint16) {

	if dstIP[0] >= 224 || dstIP[len(dstIP)-1] == 255 {
		// 过滤D类和E类地址 过滤广播
		return
	}

	go func() {
		conn, err := net.DialUDP("udp",
			&net.UDPAddr{IP: outboundIp},
			&net.UDPAddr{IP: dstIP, Port: int(dstPort)})
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
