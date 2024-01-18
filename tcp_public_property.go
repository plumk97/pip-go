package pipgo

func (tcp *TCP) Iden() uint32 {
	return tcp.iden
}

func (tcp *TCP) Status() TCPStatus {
	return tcp.status
}

func (tcp *TCP) IPHeader() *IPHeader {
	return tcp.ipHeader
}

func (tcp *TCP) SrcPort() uint16 {
	return tcp.srcPort
}

func (tcp *TCP) DstPort() uint16 {
	return tcp.dstPort
}

func (tcp *TCP) OppWindShift() uint8 {
	return tcp.oppWindShift
}
