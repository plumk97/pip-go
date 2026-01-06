package checksum

import (
	"encoding/binary"
	"net"

	"github.com/plumk97/pip-go/lib/chainbuf"
)

func InetChecksum(payload []byte, proto uint8, srcIP, dstIP net.IP) uint16 {

	var sum uint32 = 0
	addr := binary.BigEndian.Uint32(srcIP.To4())
	sum += (addr & 0xFFFF0000) >> 16
	sum += (addr & 0x0000FFFF)

	addr = binary.BigEndian.Uint32(dstIP.To4())
	sum += (addr & 0xFFFF0000) >> 16
	sum += (addr & 0x0000FFFF)

	sum += uint32(proto) & 0x0000FFFF

	len := uint32(len(payload))
	sum += (len & 0xFFFF0000) >> 16
	sum += (len & 0x0000FFFF)

	return uint16(^Checksum(payload, sum))
}

func InetChecksumBuf(buf *chainbuf.ChainBuffer, proto uint8, srcIP, dstIP net.IP) uint16 {

	var sum uint32 = 0

	for i := 0; i < len(srcIP); i += 4 {
		addr := binary.BigEndian.Uint32(srcIP[i : i+4])
		sum += (addr & 0xFFFF0000) >> 16
		sum += (addr & 0x0000FFFF)
	}

	for i := 0; i < len(dstIP); i += 4 {
		addr := binary.BigEndian.Uint32(dstIP[i : i+4])
		sum += (addr & 0xFFFF0000) >> 16
		sum += (addr & 0x0000FFFF)
	}

	sum += uint32(proto) & 0x0000FFFF

	len := uint32(buf.TotalLen())
	sum += (len & 0xFFFF0000) >> 16
	sum += (len & 0x0000FFFF)

	for q := buf; q != nil; q = q.Next() {
		sum = Checksum(q.Payload(), sum)
	}

	return uint16(^sum)
}
