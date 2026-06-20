package domain

import (
	"time"

	"github.com/google/uuid"
)

type CV struct {
	UserID      uuid.UUID
	FileName    string
	ContentType string
	StorageKey  string
	SizeBytes   int64
	UpdatedAt   time.Time
}

type CVObject struct {
	UserID      uuid.UUID
	StorageKey  string
	ContentType string
	Content     []byte
}
