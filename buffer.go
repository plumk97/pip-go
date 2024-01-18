package pipgo

type Buffer struct {
	payload  []byte
	totalLen int
	next     *Buffer
}

func NewBuffer(payload []byte) *Buffer {
	buf := &Buffer{}
	buf.payload = payload
	buf.totalLen = len(payload)
	return buf
}

func (buf *Buffer) SetNext(nextBuf *Buffer) {
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

func (buf *Buffer) Next() *Buffer {
	return buf.next
}

func (buf *Buffer) TotalLen() int {
	return buf.totalLen
}

func (buf *Buffer) Payload() []byte {
	return buf.payload
}
