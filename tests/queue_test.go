package tests_test

import (
	"log"
	"testing"

	pipgo "github.com/plumk97/pip-go"
)

func TestQueue(t *testing.T) {
	queue := pipgo.NewQueue[int]()

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
