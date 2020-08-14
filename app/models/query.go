package models

import "crypto/sha256"

type Query struct {
	Str    string
	Hash   [sha256.Size]byte
}

func NewQuery(str string) *Query {
	return &Query{Str: str, Hash: sha256.Sum256([]byte(str))}
}
