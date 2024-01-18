package pipgo

type QueueNode[T any] struct {
	value T
	next  *QueueNode[T]
}
type Queue[T any] struct {
	size int
	head *QueueNode[T]
	foot *QueueNode[T]
}

func NewQueue[T any]() *Queue[T] {
	return &Queue[T]{}
}

func (q *Queue[T]) Front() T {
	return q.head.value
}

func (q *Queue[T]) Push(value T) {

	node := &QueueNode[T]{
		value: value,
	}

	if q.head == nil {
		q.head = node
	}

	if q.foot == nil {
		q.foot = node
	} else {
		q.foot.next = node
		q.foot = node
	}

	q.size += 1
}

func (q *Queue[T]) Pop() {
	if q.size > 0 {
		node := q.head.next

		if q.head == q.foot {
			q.head = nil
			q.foot = nil
		} else {
			q.head = nil
		}

		q.head = node
		q.size -= 1
	}
}

func (q *Queue[T]) Empty() bool {
	return q.size <= 0
}

func (q *Queue[T]) Size() int {
	return q.size
}
