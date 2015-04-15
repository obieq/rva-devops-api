package controllers

import (
	"log"

	"github.com/go-martini/martini"
	"github.com/martini-contrib/render"
	goar "github.com/obieq/goar"
	models "github.com/obieq/rva-devops-api/models"
	resources "github.com/obieq/rva-devops-api/resources"
)

func HandleGetAutomobiles(r render.Render) {
	var automobiles []resources.Automobile

	dbModels := make([]models.Automobile, 0)
	err := models.Automobile{}.ToActiveRecord().All(&dbModels, nil)

	log.Println("Err:", err)
	// map the models to resources
	if err == nil {
		log.Println("DB Models:", len(dbModels))
		automobiles = make([]resources.Automobile, len(dbModels))
		for i, m := range dbModels {
			automobile := resources.Automobile{}
			automobile.MapFromModel(&m)
			automobiles[i] = automobile
		}
	}
	HandleIndexResponse(err, resources.Link{}, automobiles, r)
}

func HandleGetAutomobile(args martini.Params, r render.Render) {
	var automobile resources.Automobile

	dbAutomobile, err := goar.ToAR(&models.Automobile{}).Find(args["id"])

	// map the model to the resource
	if err == nil {
		automobile = resources.Automobile{}
		automobile.MapFromModel(dbAutomobile)
	}
	HandleGetResponse(err, automobile, r)
}

func HandleCreateAutomobile(request resources.AutomobileJsonApiRequest, r render.Render) {
	var resource resources.Automobile

	// map the resource to the model
	m := &models.Automobile{}
	request.MapToModel(m)

	// persist the model
	success, err := m.Save() // TODO: implement patch?

	// map the model to the resource
	if err == nil {
		resource = resources.Automobile{}
		resource.MapFromModel(m)
	}

	// process result
	HandlePostResponse(success, err, &resource, r)
}

func HandleUpdateAutomobile(args martini.Params, request resources.AutomobileJsonApiRequest, r render.Render) {
	var resource resources.Automobile

	result, err := models.Automobile{}.ToActiveRecord().Find(args["id"])

	if err == nil {
		dbModel := result.(*models.Automobile)

		// update properties
		request.MapToModel(dbModel)

		// persist changes
		success, err := dbModel.Save()

		// map the model to the resource
		if err == nil {
			resource = resources.Automobile{}
			resource.MapFromModel(dbModel)
		}

		// process result
		HandlePostResponse(success, err, &resource, r)
	} else { // get failed, so re-use the get response method, which properly handles the error condition
		HandleGetResponse(err, result, r)
	}
}

func HandleDeleteAutomobile(args martini.Params, r render.Render) {
	model := models.Automobile{}
	model.ID = args["id"]

	HandleDeleteResponse(&model, r)
}
