package storage

import (
	"app/models"
	"crypto/sha256"
)

type Memory struct {
	runs map[[sha256.Size]byte]*models.Run
}

func NewMemory() Storage {
	return &Memory{}
}

func (m *Memory) AddRun(run *models.Run) {
	m.runs[run.Query.Hash] = run
}

func (m *Memory) FindRun(hash [sha256.Size]byte) *models.Run {
	return m.runs[hash]
}
