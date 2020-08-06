package worker

import (
	"app/docker"
	"go.uber.org/multierr"
	"sync"
)

type Pool struct {
	manager *docker.Manager
	workers []*Worker
}

func NewPool(workersNum int) (*Pool, error) {
	var pool Pool

	manager, err := docker.NewManager()
	if err != nil {
		return nil, err
	}
	pool.manager = manager

	pool.workers = make([]*Worker, workersNum)
	for i := 0; i < workersNum; i++ {
		pool.workers[i], err = New(i, pool.manager)
		if err != nil {
			return nil, err
		}
	}

	return &pool, nil
}

func (p *Pool) Start() error {
	var multiErr error
	var wg sync.WaitGroup
	for i := 0; i < len(p.workers); i++ {
		i := i
		wg.Add(1)
		go func() {
			defer wg.Done()
			var err error
			err = p.workers[i].Run()
			multierr.AppendInto(&multiErr, err)
		}()
	}

	wg.Wait()
	if len(multierr.Errors(multiErr)) > 0 {
		return multiErr
	}

	return nil
}

func (p *Pool) Execute(query string) (string, error) {
	return p.workers[0].Execute(query)
}

