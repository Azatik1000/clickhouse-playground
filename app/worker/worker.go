package worker

import (
	"app/docker"
	"app/driver"
	"fmt"
	"net/url"
)

type Worker struct {
	id        int
	port      int
	alias     string
	manager   *docker.Manager
	dbDriver  driver.Driver
	container *docker.Container
}

func (w *Worker) resetContainer() error {
	container, err := docker.NewContainer(w.port, w.manager, w.alias)
	if err != nil {
		return err
	}

	err = container.Run(w.handleExit)
	if err != nil {
		return err
	}

	w.container = container
	return nil
}

func (w *Worker) handleExit() {
	// TODO: handle error
	container, _ := docker.NewContainer(w.port, w.manager, w.alias)
	w.container = container
}

func New(id int, manager *docker.Manager) (*Worker, error) {
	var worker Worker
	worker.id = id
	worker.port = 8123 + id
	worker.alias = fmt.Sprintf("db%d", worker.id)
	worker.manager = manager
	return &worker, nil
}

func (w *Worker) Run() error {
	err := w.resetContainer()
	if err != nil {
		return err
	}

	endpoint, _ := url.Parse(fmt.Sprintf("http://%s:%d", w.alias, 8123))
	w.dbDriver = driver.NewHTTPDriver(endpoint)
	err = w.dbDriver.HealthCheck()
	if err != nil {
		return err
	}

	return err
}

func (w *Worker) Execute(query string) (string, error) {
	return w.dbDriver.Exec(query)
}
