package main

import (
	"log"
	"net/http"
	"net/http/httputil"

	"github.com/go-martini/martini"
	"github.com/martini-contrib/binding"
	"github.com/martini-contrib/render"
	controllers "github.com/obieq/rva-devops-api/controllers"
	"github.com/obieq/rva-devops-api/resources"
	"github.com/twinj/uuid"
)

func main() {
	// switch the uuid format
	uuid.SwitchFormat(uuid.CleanHyphen)

	m := martini.Classic()

	// use render contrib library within controllers
	m.Use(render.Renderer())

	m.Use(func(req *http.Request) {
		log.Println("CALLING HTTP REQUEST INTERCEPTOR!")
		dump, _ := httputil.DumpRequest(req, true)
		log.Println(string(dump))
	})

	m.Get("/", func() string {
		return "Hello world!"
	})

	// quote intent routes
	m.Get("/api/v1/automobiles", controllers.HandleGetAutomobiles)
	m.Get("/api/v1/automobiles/:id", controllers.HandleGetAutomobile)
	m.Post("/api/v1/automobiles", binding.Json(resources.AutomobileJsonApiRequest{}), controllers.HandleCreateAutomobile)
	m.Put("/api/v1/automobiles/:id", binding.Json(resources.AutomobileJsonApiRequest{}), controllers.HandleUpdateAutomobile)
	m.Delete("/api/v1/automobiles/:id", controllers.HandleDeleteAutomobile)

	m.RunOnAddr(":5000")
}
