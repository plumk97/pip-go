package checksum

import (
	"encoding/binary"
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
