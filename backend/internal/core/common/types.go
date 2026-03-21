package common

import "github.com/google/uuid"

type EntityRef struct {
	Type string    `json:"type"`
	ID   uuid.UUID `json:"id"`
}
