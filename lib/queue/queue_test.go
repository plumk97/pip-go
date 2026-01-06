package queue_test

import (
	"log"
	"testing"

	"github.com/plumk97/pip-go/lib/queue"
)

func TestQueue(t *testing.T) {
	queue := queue.NewQueue[int]()

	queue.Push(1)
	queue.Push(2)
	queue.Push(3)
	log.Println(queue.Size())

	log.Println(queue.Front())
	queue.Pop()

	log.Println(queue.Front())
	queue.Pop()

	log.Println(queue.Front())
	queue.Pop()

	log.Println(queue.Size())
}
