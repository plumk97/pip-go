package checksum

func IPChecksum(payload []byte) uint16 {
	sum := Checksum(payload, 0)
	return uint16(^sum)
}
