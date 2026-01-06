package transfer

import (
	"log"
	"net"
	"runtime"
	"syscall"

	pipgo "github.com/plumk97/pip-go"
	"github.com/plumk97/pip-go/types"
)

func receiveICMPDataCallback(netif *pipgo.Netif, data []byte, srcIP, dstIP net.IP, ttl uint8) {

	conn, err := net.DialIP("ip:icmp", &net.IPAddr{IP: outboundIp}, &net.IPAddr{IP: dstIP})
	if err != nil {
		log.Println(err)
		return
	}
	defer conn.Close()

	syscall.Socket(syscall.AF_INET, syscall.SOCK_DGRAM, types.IPPROTO_ICMP)
	conn.Write(data)

	if runtime.GOOS == "windows" {
		b := make([]byte, 1024)
		n, _ := conn.Read(b)

		// 返回的是完整的IP包
		if n > 0 {
			netif.ICMPOutput(b[20:n], dstIP, srcIP)
		}
	}
}
