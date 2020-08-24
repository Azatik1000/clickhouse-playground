package worker

import (
	"app/docker"
	"app/models"
	"fmt"
	"github.com/Workiva/go-datastructures/queue"
	"sync"
)

type Pool struct {
	manager *docker.Manager
	mu      sync.Mutex
	queue   *queue.Queue
}

func NewPool(workersNum int) (*Pool, error) {
	var pool Pool

	manager, err := docker.NewManager()
	if err != nil {
		return nil, err
	}
	pool.manager = manager

	pool.queue = queue.New(int64(workersNum))

	type Result struct {
		*Worker
		error
	}

	results := make(chan Result)

	for i := 0; i < workersNum; i++ {
		go func() {
			w, err := New(pool.manager)
			results <- Result{w, err}
		}()
	}

	for i := 0; i < workersNum; i++ {
		result := <-results
		if result.error != nil {
			return nil, result.error
		}

		pool.queue.Put(result.Worker)
	}

	return &pool, nil
}

func (p *Pool) Execute(query string) (models.Result, error) {
	// TODO: handle error
	first, _ := p.queue.Get(1)
	worker := first[0].(*Worker)

	result, err := worker.Execute(query)
	// TODO: handle error

	go func() {
		err := worker.restartServer()
		if err != nil {
			fmt.Println(err)
		}

		p.queue.Put(worker)
	}()

	return models.Result(result), err
}
