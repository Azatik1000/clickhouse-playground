package storage

import (
	"app/models"
	"crypto/sha256"
)

type Memory struct {
	runs map[[sha256.Size]byte]*models.Run
}

func NewMemory() Storage {
	return &Memory{
		runs: make(map[[sha256.Size]byte]*models.Run),
	}
}

func (m *Memory) AddRun(run *models.Run) error {
	m.runs[run.Query.Hash] = run
	return nil
}

func (m *Memory) FindRun(query *models.Query) (*models.Run, error) {
	return m.runs[query.Hash], nil
}
