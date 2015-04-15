package resources

import (
	"strings"
	"time"

	as "github.com/obieq/goar/active_support"
	resources "github.com/obieq/goar/validations"
)

const API_PATH string = "http://192.168.1.8:4000/api/v1"

type JsonApiResourcer interface {
	Resourcer
	BuildLinks(model interface{})
	SelfLink() string
}

type Resourcer interface {
	MapToModel(model interface{})
	MapFromModel(model interface{})
	Errors() map[string]string
	SetErrors(map[string]*resources.ValidationError)
}

type Link struct {
	Self    string `json:"self,omitempty"`
	Related string `json:"related,omitempty"`
}

// CollectionLink => JSON API links
type CollectionLink struct {
	Self    string    `json:"self,omitempty"`
	Related string    `json:"related,omitempty"`
	Linkage []Linkage `json:"linkage,omitempty"`
}

// Linkage => JSON API linkage
type Linkage struct {
	Type string `json:"type"`
	ID   string `json:"id"`
}

type BaseResource struct {
	ResourceType string     `json:"type,omitempty"`
	ID           string     `json:"id,omitempty"`
	CreatedAt    *time.Time `json:"created-at,omitempty"`
	UpdatedAt    *time.Time `json:"updated-at,omitempty"`
	errors       map[string]*resources.ValidationError
}

func (r *BaseResource) Errors() map[string]string {
	errors := make(map[string]string)

	for k, v := range r.errors {
		key := as.String(k).Dasherize()
		errors[key] = strings.ToLower(v.Message)
	}

	return errors
}

func (r *BaseResource) SetErrors(errors map[string]*resources.ValidationError) {
	r.errors = errors
}
