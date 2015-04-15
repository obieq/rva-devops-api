package models

import goar "github.com/obieq/goar"

type Automobile struct {
	BaseModel
	Year  int    `json:"year,omitempty"`
	Make  string `json:"make,omitempty"`
	Model string `json:"model,omitempty"`
}

func (model Automobile) ToActiveRecord() *Automobile {
	return goar.ToAR(&model).(*Automobile)
}

func (m *Automobile) Validate() {
	//m.Validation.Required("ID", m.ID)
	m.Validation.Required("Year", m.Year)
	m.Validation.Required("Make", m.Make)
	m.Validation.Required("Model", m.Model)
}
