package resources

import (
	"log"

	"github.com/obieq/goar"
	"github.com/obieq/rva-devops-api/models"
)

const AUTOMOBILE_RESOURCE_TYPE string = "automobiles"

// AutomobileLinks => JSON API links
type AutomobileLinks struct {
	Link
}

// Automobile model
type Automobile struct {
	BaseResource
	Year  int             `json:"year,omitempty"`
	Make  string          `json:"make,omitempty"`
	Model string          `json:"model,omitempty"`
	Links AutomobileLinks `json:"links,omitempty"`
}

// AutomobileJsonApiRequest => struct for receiving and processing a JSON API request
type AutomobileJsonApiRequest struct {
	Automobile `json:"data,omitempty"`
}

// BuildLinks => builds JSON API links
func (p *Automobile) BuildLinks(automobileModel interface{}) {
	root := API_PATH + "/" + AUTOMOBILE_RESOURCE_TYPE + "/" + p.ID
	p.Links = AutomobileLinks{
		Link: Link{Self: root}}
}

func (p *Automobile) SelfLink() string {
	return p.Links.Self
}

func (r *Automobile) MapFromModel(model interface{}) {
	m := model.(*models.Automobile)

	if !m.HasErrors() {
		r.ResourceType = AUTOMOBILE_RESOURCE_TYPE
		r.ID = m.ID
		r.CreatedAt = m.CreatedAt
		r.UpdatedAt = m.UpdatedAt

		r.Year = m.Year
		r.Make = m.Make
		r.Model = m.Model

		// build links
		r.BuildLinks(m)
	} else {
		r.SetErrors(m.ErrorMap())
	}
}

func (r *Automobile) MapToModel(model interface{}) {
	m := model.(*models.Automobile)

	m.Year = r.Year
	m.Make = r.Make
	m.Model = r.Model

	log.Println("Resource:", r)
	log.Println("Model:", m)

	if m.CreatedAt == nil { // we're inserting a new record
		m.ID = r.ID
	}

	// conver model to an active record model
	goar.ToAR(m)
}
