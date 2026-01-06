package main

import (
	"log"
	"os/exec"
	"strconv"

	"github.com/labulakalia/water"
	pipgo "github.com/plumk97/pip-go"
	"github.com/plumk97/pip-go/examples/transfer"
	"github.com/plumk97/pip-go/lib/chainbuf"
)

func main() {
	var err error

	// 建立utun网卡
	ifce, err := water.New(water.Config{
		DeviceType: water.TUN,
	})
	if err != nil {
		log.Fatalln(err)
	}
	defer ifce.Close()

	// 设置网关地址
	cmd := exec.Command("ifconfig",
		ifce.Name(),
		"192.168.33.1", "192.168.33.1",
		"netmask", "255.255.255.255",
		"mtu", strconv.Itoa(int(pipgo.MTU)),
		"up")

	err = cmd.Run()
	if err != nil {
		log.Fatalln(err)
	}

	// 设置路由
	if err := exec.Command("route", "-n", "add", "-net", "1.1.1.1/32", "-interface", ifce.Name()).Run(); err != nil {
		log.Fatalln(err)
	}

	log.Println("Ifce Name:", ifce.Name())
	log.Println("Ifce MTU:", pipgo.MTU)
	log.Println("Ifce Addr: 192.168.33.1")
	log.Println("Ifce Router: 1.1.1.1/32")

	netif := pipgo.NewNetif()
	defer netif.Close()

	transfer.Bind(netif, ifce.Name())

	// 初始化回调函数
	netif.OutputIPData = func(netif *pipgo.Netif, buf *chainbuf.ChainBuffer) {
		b := make([]byte, buf.TotalLen())

		offset := 0
		for q := buf; q != nil; q = q.Next() {
			copy(b[offset:], q.Payload())
			offset += len(q.Payload())
		}

		ifce.Write(b)
	}

	// 读取网卡数据并输入到netif中
	packet := make([]byte, pipgo.MTU)
	for {
		n, err := ifce.Read(packet)
		if err != nil {
			log.Fatal(err)
		}
		netif.Input(packet[:n])
	}
}
