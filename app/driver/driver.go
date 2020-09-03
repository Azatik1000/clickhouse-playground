package driver

import "app/models"

type Driver interface {
	Exec(query string) (models.Result, error)
	HealthCheck() error
	Close() error
}



