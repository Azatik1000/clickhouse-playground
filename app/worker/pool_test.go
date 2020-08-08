package worker

import (
	"github.com/Workiva/go-datastructures/queue"
	"testing"
)

func TestPoolSimple(t *testing.T) {
	q := queue.New(0)

	go func() {
		q.Get(1)
	}()

	q.Put(282)
}
