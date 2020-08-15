package storage

import (
	"app/models"
)

type Storage interface {
	AddRun(run *models.Run) error
	FindRun(query *models.Query) (*models.Run, error)
}
