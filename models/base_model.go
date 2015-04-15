package models

import (
	aro "github.com/obieq/goar/db/orchestrate"
	"github.com/twinj/uuid"
)

type BaseModel struct {
	aro.ArOrchestrate
}

func (m *BaseModel) BeforeSave() error {
	var err error = nil

	// Generate an ID if one hasn't been specified AND the model hasn't been persisted yet
	if m.CreatedAt == nil && m.ID == "" {
		u := uuid.NewV4()
		m.ID = u.String()
	}

	return err
}
