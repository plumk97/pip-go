package chainbuf

type ChainBuffer struct {
	payload  []byte
	totalLen int
	next     *ChainBuffer
}

func NewChainBuffer(payload []byte) *ChainBuffer {
	buf := &ChainBuffer{}
	buf.payload = payload
	buf.totalLen = len(payload)
	return buf
}

func (buf *ChainBuffer) SetNext(nextBuf *ChainBuffer) {
	if buf.next != nil {
		buf.totalLen -= buf.next.totalLen
	}

	if nextBuf == nil {
		buf.next = nil
	} else {
		buf.totalLen += nextBuf.totalLen
		buf.next = nextBuf
	}
}

func (buf *ChainBuffer) Next() *ChainBuffer {
	return buf.next
}

func (buf *ChainBuffer) TotalLen() int {
	return buf.totalLen
}

func (buf *ChainBuffer) Payload() []byte {
	return buf.payload
}
