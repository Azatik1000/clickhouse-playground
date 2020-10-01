package main

import (
	"app/driver"
	"app/models"
	"app/storage"
	"encoding/json"
	"fmt"
	_ "github.com/ClickHouse/clickhouse-go"
	"github.com/rs/cors"
	"log"
	"net/http"
	"strings"
)

type pgServer struct {
	driver  driver.Driver
	storage storage.Storage
}

func newPgServer(driver driver.Driver, s storage.Storage) *pgServer {
	return &pgServer{driver: driver, storage: s}
}

func main() {
	s, err := storage.NewDatabase()
	if err != nil {
		log.Fatal(err)
	}

	driver, err := driver.NewExecutor(
		"amqp://test:test@my-release-rabbitmq:5672",
		"results",
	)
	if err != nil {
		log.Fatal(err)
	}

	server := newPgServer(driver, s)

	c := cors.Default()

	mux := http.NewServeMux()
	mux.HandleFunc("/exec", server.handleExec)

	handler := c.Handler(mux)

	_ = http.ListenAndServe(":8080", handler)
}

type execInput struct {
	QueryStr  string `json:"query"`
	VersionID string `json:"versionID"`
}

type execOutput struct {
	Link   string                   `json:"link"`
	Cached bool                     `json:"cached"`
	Result []map[string]interface{} `json:"result"`
}

// TODO: add logging to requests through wrappers
func (s *pgServer) handleExec(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)

	var input execInput
	err := decoder.Decode(&input)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	log.Println("read your query")

	var output execOutput

	query := models.NewQuery(input.QueryStr)
	// TODO: move to getlink method
	output.Link = "/runs/" + query.Hash.Hex()

	// TODO: check version
	found, err := s.storage.FindRun(query)
	//var found *models.Run = nil

	var result *models.Result

	// This query 's been already run
	if found != nil {
		output.Cached = true
		result = &found.Result
	} else {
		output.Cached = false

		fmt.Println("sending your query to driver")

		result, err = s.driver.Exec(input.QueryStr, input.VersionID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// TODO: pointers ok?
		s.storage.AddRun(&models.Run{
			Query:  *query,
			Result: *result,
		})
	}

	d := json.NewDecoder(strings.NewReader(string(*result)))
	d.Decode(&output.Result)

	encoder := json.NewEncoder(w)
	encoder.Encode(output)
}
