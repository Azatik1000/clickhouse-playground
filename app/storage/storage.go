package storage

import (
	"app/models"
	"crypto/sha256"
)

type Storage interface {
	AddRun(run *models.Run)
	FindRun(hash [sha256.Size]byte) *models.Run
}
