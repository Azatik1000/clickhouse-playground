package driver

import (
	"app/docker"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"sync"
)

type executorDriver struct {
	executor *docker.ExecutorContainer
	endpoint *url.URL
}

var executorCount = 0
var executorMu sync.Mutex

func NewExecutor(manager *docker.Manager) (Driver, error) {
	var driver executorDriver

	executorMu.Lock()
	id := executorCount
	executorCount++
	executorMu.Unlock()

	alias := fmt.Sprintf("client%d", id)

	executor, err := docker.NewExecutor(
		manager,
		alias,
	)
	if err != nil {
		return nil, err
	}

	if _, err = (*docker.Container)(executor).Run(); err != nil {
		return nil, err
	}


	driver.executor = executor

	endpoint, _ := url.Parse(
		fmt.Sprintf("http://%s:%d/exec", alias, 8080),
	)

	driver.endpoint = endpoint

	return &driver, nil
}

func (c *executorDriver) Exec(query string) (string, error) {
	fmt.Println("Exec:", query)

	// TODO: change to json
	response, err := http.Post(
		c.endpoint.String(),
		"",
		strings.NewReader(query),
	)
	if err != nil {
		return "", err
	}
	defer func() {
		_ = response.Body.Close()
	}()

	data, err := ioutil.ReadAll(response.Body)
	if response.StatusCode != 200 {
		return "", errors.New(string(data))
	}

	if err != nil {
		return "", err
	}

	return string(data), nil
}

func (c *executorDriver) HealthCheck() error {
	// TODO: change
	return nil
}

func (c *executorDriver) Close() error {
	// TODO: remove container
	return nil
}
