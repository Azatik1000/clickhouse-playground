package driver

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

type executorDriver struct {
	endpoint *url.URL
}

func NewExecutor(endpointStr string) (Driver, error) {
	var driver executorDriver

	// TODO: handle error
	endpoint, _ := url.Parse(endpointStr)

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
