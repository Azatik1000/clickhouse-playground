package worker

import (
	"app/docker"
	"app/driver"
	"fmt"
	"time"
)

type Worker struct {
	id       int
	port     int
	name     string
	alias    string
	manager  *docker.Manager
	dbDriver driver.Driver
	dbServer *docker.DbServerContainer
	exited   chan struct{}
}

func (w *Worker) startDb() error {
	dbServer, err := docker.NewDbServer(w.port, w.manager, w.name, w.alias)
	if err != nil {
		return err
	}

	exited, err := (*docker.Container)(dbServer).Run()
	if err != nil {
		return err
	}

	w.dbServer = dbServer
	w.exited = exited

	return nil
}

func New(id int, manager *docker.Manager) (*Worker, error) {
	var worker Worker
	worker.id = id
	worker.port = 8123 + id
	worker.name = fmt.Sprintf("db%d", worker.id)
	worker.alias = worker.name
	worker.manager = manager

	if err := worker.startDb(); err != nil {
		return nil, err
	}

	// TODO: try to move to different threads

	if err := worker.startClient(); err != nil {
		return nil, err
	}

	return &worker, nil
}

func (w *Worker) startClient() error {
	dbDriver, err := driver.NewConsoleClientDriver(w.manager, w.name)
	if err != nil {
		return err
	}

	w.dbDriver = dbDriver
	return nil
}

func (w *Worker) checkConnection() error {
	var err error
	for i := 0; i < 10; i++ {
		err = w.dbDriver.HealthCheck()
		if err == nil {
			return nil
		}

		// TODO: change to smarter method
		time.Sleep(5 * time.Second)
	}

	return fmt.Errorf("couldn't connect: %e", err)
}

func (w *Worker) Restart() error {
	if err := w.Stop(); err != nil {
		return err
	}

	if err := w.startDb(); err != nil {
		return err
	}

	if err := w.checkConnection(); err != nil {
		return err
	}

	return nil
}

func (w *Worker) Stop() error {
	return (*docker.Container)(w.dbServer).Stop()
}

func (w *Worker) Execute(query string) (string, error) {
	return w.dbDriver.Exec(query)
}
