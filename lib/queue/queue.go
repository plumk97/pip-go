package queue

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

	q.size++
}

func (q *Queue[T]) Pop() {
	if q.size == 0 {
		return
	}
	q.head = q.head.next
	if q.head == nil {
		q.foot = nil
	}
	q.size--
}

func (q *Queue[T]) Empty() bool {
	return q.size <= 0
}

func (q *Queue[T]) Size() int {
	return q.size
}
