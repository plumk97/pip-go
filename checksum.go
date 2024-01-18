package pipgo

import (
	"encoding/binary"
	"net"
)

func fold(n uint32) uint32 {
	return (n & 0x0000FFFF) + (n >> 16)
}

func Checksum(payload []byte, sum uint32) uint32 {
	i := 0
	len := len(payload)

	for i < len {
		if i+1 >= len {
			break
		}
		sum += uint32(binary.BigEndian.Uint16(payload[i : i+2]))
		i += 2
	}

	if i < len {
		sum += uint32(payload[i])<<8 | 0
	}

	sum = fold(sum)
	sum = fold(sum)
	return sum
}

func IPChecksum(payload []byte) uint16 {
	sum := Checksum(payload, 0)
	return uint16(^sum)
}

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

func InetChecksumBuf(buf *Buffer, proto uint8, srcIP, dstIP net.IP) uint16 {

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

	len := uint32(buf.totalLen)
	sum += (len & 0xFFFF0000) >> 16
	sum += (len & 0x0000FFFF)

	for q := buf; q != nil; q = q.next {
		sum = Checksum(q.payload, sum)
	}

	return uint16(^sum)
}
