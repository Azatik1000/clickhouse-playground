package worker

import (
	"app/docker"
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
		i := i
		go func() {
			w, err := New(i, pool.manager)
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

//func (p *Pool) Start() error {
//	var multiErr error
//	var wg sync.WaitGroup
//
//	fmt.Println(p.queue.Len())
//	for i := 0; i < int(p.queue.Len()); i++ {
//		wg.Add(1)
//		go func() {
//			defer wg.Done()
//			var err error
//			// TODO: handle error
//			first, _ := p.queue.Get(1)
//			worker := first[0].(*Worker)
//			err = worker.StartDb()
//			multierr.AppendInto(&multiErr, err)
//			p.queue.Put(worker)
//		}()
//	}
//	wg.Wait()
//
//	if len(multierr.Errors(multiErr)) > 0 {
//		return multiErr
//	}
//
//	return nil
//}

func (p *Pool) Execute(query string) (string, error) {
	// TODO: handle error
	first, _ := p.queue.Get(1)
	worker := first[0].(*Worker)

	result, err := worker.Execute(query)
	// TODO: handle error
	go func() {
		_ = worker.Restart()
		p.queue.Put(worker)
	}()

	return result, err
}
