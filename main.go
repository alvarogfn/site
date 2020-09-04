package main

import (
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/candidatos-info/site/db"
	"github.com/labstack/echo"
)

var (
	dbClient       *db.DataStoreClient
	candidateRoles = []string{"vereador", "prefeito"} // available candidate roles
)

type tmplt struct {
	templates *template.Template
}

func (t *tmplt) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

func homePageHandler(c echo.Context) error {
	states, err := dbClient.GetStates()
	if err != nil {
		log.Printf("failed to retrieve states from db, erro %v", err)
		return c.String(http.StatusInternalServerError, err.Error())
	}
	templateData := struct {
		States         []string
		CandidateTypes []string
	}{
		states,
		candidateRoles,
	}
	return c.Render(http.StatusOK, "main.html", templateData)
}

func profilesPageHandler(c echo.Context) error {
	city := c.QueryParam("city")
	state := c.QueryParam("state")
	role := c.QueryParam("role")
	year := c.Param("year")
	y, _ := strconv.Atoi(year)
	candidates, _ := dbClient.FindCandidatesWithParams(state, city, role, y)
	templateData := struct {
		State        string
		City         string
		Role         string
		Candidatures []*db.CandidateForDB
		Year         int
	}{
		state,
		city,
		role,
		candidates,
		y,
	}
	return c.Render(http.StatusOK, "profiles.html", templateData)
}

func candidatePageHandler(c echo.Context) error {
	year := c.Param("year")
	yearAsInt, err := strconv.Atoi(year)
	if err != nil {
		log.Printf("failed to parse given year [%s] to int, erro %v", year, err)
		return c.String(http.StatusInternalServerError, err.Error())
	}
	state := c.Param("state")
	city := c.Param("city")
	role := c.QueryParam("role")
	sequencialCandidate := c.Param("sequencialCandidate")
	candidate, _ := dbClient.GetCandidateBySequencialID(yearAsInt, state, city, sequencialCandidate)
	templateData := struct {
		State        string
		City         string
		Role         string
		PhotoURL     string
		Name         string
		Party        string
		Twitter      string
		Description  string
		BallotNumber int
	}{
		state,
		city,
		role,
		candidate.PhotoURL,
		candidate.BallotName,
		candidate.Party,
		candidate.Twitter,
		candidate.Description,
		candidate.BallotNumber,
	}
	return c.Render(http.StatusOK, "candidate.html", templateData)
}

func citiesOfState(c echo.Context) error {
	state := c.QueryParam("state")
	if state == "" {
		return c.String(http.StatusBadRequest, "estado inválido")
	}
	citesOfState, err := dbClient.GetCities(state)
	if err != nil {
		log.Printf("failed to retrieve cities of state [%s], erro %v", state, err)
		return c.String(http.StatusInternalServerError, fmt.Sprintf("erro ao buscar cidades do estado [%s], erro %v", state, err))
	}
	return c.JSON(http.StatusOK, citesOfState)
}

func main() {
	projectID := os.Getenv("PROJECT_ID")
	if projectID == "" {
		log.Fatal("missing PROJECT_ID environment variable")
	}
	dbClient = db.NewDataStoreClient(projectID)
	log.Println("connected to database")
	e := echo.New()
	e.Renderer = &tmplt{
		templates: template.Must(template.ParseGlob("templates/*.html")),
	}
	e.Static("/static", "templates/")
	e.GET("/", homePageHandler)
	e.POST("/profiles/:year", profilesPageHandler)
	e.GET("/candidato/:year/:state/:city/:sequencialCandidate", candidatePageHandler)
	e.GET("/api/v1/cities", citiesOfState) // return the cities of a given state passed as a query param
	port := os.Getenv("PORT")
	if port == "" {
		log.Fatal("missing PORT environment variable")
	}
	log.Fatal(e.Start(":" + port))
}
