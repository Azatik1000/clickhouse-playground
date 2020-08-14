package worker

import (
	"app/docker"
	"app/driver"
	"fmt"
)

type Worker struct {
	id       int
	name     string
	alias    string
	manager  *docker.Manager
	dbDriver driver.Driver
}

func New(manager *docker.Manager) (*Worker, error) {
	var worker Worker
	worker.name = fmt.Sprintf("db%d", worker.id)
	worker.alias = worker.name
	worker.manager = manager

	dbDriver, err := driver.NewExecutor(manager)
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
