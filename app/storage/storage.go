package storage

import (
	"app/models"
)

type Storage interface {
	AddRun(run *models.Run)
	FindRun(query *models.Query) *models.Run
}
