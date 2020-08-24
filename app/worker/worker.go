package worker

import (
	"app/docker"
	"app/driver"
	"fmt"
	"sync"
)

var workerCount = 0
var workerMu sync.Mutex

type Worker struct {
	id       int
	manager  *docker.Manager
	server   *docker.DbServerContainer
	executor *docker.ExecutorContainer
	dbDriver driver.Driver
}

func New(manager *docker.Manager) (*Worker, error) {
	var worker Worker
	worker.manager = manager

	workerMu.Lock()
	id := workerCount
	workerCount++
	workerMu.Unlock()

	serverAlias := fmt.Sprintf("db%d", id)
	executorAlias := fmt.Sprintf("client%d", id)

	server, err := docker.NewServer(manager, serverAlias)
	if err != nil {
		return nil, err
	}
	if _, err = (*docker.Container)(server).Run(); err != nil {
		return nil, err
	}
	worker.server = server

	executor, err := docker.NewExecutor(
		manager,
		executorAlias,
		serverAlias,
	)
	if err != nil {
		return nil, err
	}
	if _, err = (*docker.Container)(executor).Run(); err != nil {
		return nil, err
	}
	worker.executor = executor

	dbDriver, err := driver.NewExecutor(fmt.Sprintf("http://%s:%d/exec", executorAlias, 8080),)
	if err != nil {
		return nil, err
	}
	worker.dbDriver = dbDriver

	return &worker, nil
}

//func (w *Worker) checkConnection() error {
//	var err error
//	for i := 0; i < 10; i++ {
//		err = w.dbDriver.HealthCheck()
//		if err == nil {
//			return nil
//		}
//
//		// TODO: change to smarter method
//		time.Sleep(5 * time.Second)
//	}
//
//	return fmt.Errorf("couldn't connect: %e", err)
//}



func (w *Worker) Execute(query string) (string, error) {
	return w.dbDriver.Exec(query)
}

func (w *Worker) restartServer() error {
	return (*docker.Container)(w.server).Restart()
}
