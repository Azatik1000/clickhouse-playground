package worker

import (
	"app/docker"
	"app/driver"
	"fmt"
	"net/url"
	"time"
)

type Worker struct {
	id        int
	port      int
	alias     string
	manager   *docker.Manager
	dbDriver  driver.Driver
	container *docker.Container
	exited    chan struct{}
}

func (w *Worker) setContainer() error {
	container, err := docker.NewContainer(w.port, w.manager, w.alias)
	if err != nil {
		return err
	}

	exited, err := container.Run()
	if err != nil {
		return err
	}

	w.container = container
	w.exited = exited

	return nil
}

//func (w *Worker) handleExit() {
//	// TODO: handle error
//	container, _ := docker.NewContainer(w.port, w.manager, w.alias)
//	w.container = container
//}

func New(id int, manager *docker.Manager) (*Worker, error) {
	var worker Worker
	worker.id = id
	worker.port = 8123 + id
	worker.alias = fmt.Sprintf("db%d", worker.id)
	worker.manager = manager

	endpoint, _ := url.Parse(fmt.Sprintf("http://%s:%d", worker.alias, 8123))
	worker.dbDriver = driver.NewHTTPDriver(endpoint)

	return &worker, nil
}

func (w *Worker) Start() error {
	err := w.setContainer()
	if err != nil {
		return err
	}

	for i := 0; i < 10; i++ {
		err = w.dbDriver.HealthCheck()
		if err == nil {
			break
		}
		time.Sleep(5 * time.Second)
	}

	if err != nil {
		return err
	}

	return nil
}

func (w *Worker) Stop() error {
	return w.container.Stop()
}

func (w *Worker) Execute(query string) (string, error) {
	return w.dbDriver.Exec(query)
}
