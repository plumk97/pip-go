package transfer

import (
	"log"
	"net"
	"runtime"
	"strings"

	pipgo "github.com/plumk97/pip-go"
)

var outboundIp net.IP

func Bind(netif *pipgo.Netif, ifceName string) {
	outboundIp = getOutboundIP(ifceName)

	netif.NewTCPConnect = newTCPConnectCallback
	netif.ReceiveUDPData = receiveUDPDataCallback
	netif.ReceiveICMPData = receiveICMPDataCallback
}

func getOutboundIP(ifceName string) net.IP {
	interfaces, err := net.Interfaces()
	if err != nil {
		log.Println(err)
		return nil
	}

	for _, iface := range interfaces {
		if iface.Name == ifceName {
			continue
		}

		if runtime.GOOS != "windows" && !strings.HasPrefix(iface.Name, "en") {
			continue
		}

		if addrs, err := iface.Addrs(); err == nil {
			for _, addr := range addrs {
				ipnet, ok := addr.(*net.IPNet)
				if ok && !ipnet.IP.IsLoopback() && ipnet.IP.To4() != nil {
					return ipnet.IP.To4()
				}
			}
		}
	}
	return nil
}
