package experiments

import (
	"fmt"
	"strconv"
	"sync"
	"testing"
	"time"
)

type Queue struct {
	queue []string
	cond  *sync.Cond
}

func TestCondition(t *testing.T) {
	q := Queue{
		queue: []string{},
		cond:  sync.NewCond(&sync.Mutex{}),
	}

	go func() {
		for i := range 5 {
			q.Enqueue(strconv.Itoa(i))
			time.Sleep(time.Second * 2)
		}
	}()

	for range 5 {
		result := q.Dequeue()
		fmt.Printf("dequeued: %s\n", result)
	}
}

func (q *Queue) Enqueue(item string) {
	q.cond.L.Lock()
	defer q.cond.L.Unlock()

	q.queue = append(q.queue, item)
	fmt.Printf("putting #{item} to queue notify all\n")
	q.cond.Broadcast()
}

func (q *Queue) Dequeue() string {
	q.cond.L.Lock()
	defer q.cond.L.Unlock()

	if len(q.queue) == 0 {
		fmt.Println("no data available, waiting...")
		q.cond.Wait()
	}
	result := q.queue[0]
	q.queue = q.queue[1:]
	return result
}
