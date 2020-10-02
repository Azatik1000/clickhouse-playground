package main

import (
	"app/driver"
	"app/models"
	"app/storage"
	"encoding/hex"
	"encoding/json"
	"fmt"
	_ "github.com/ClickHouse/clickhouse-go"
	mux "github.com/gorilla/mux"
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

	r := mux.NewRouter()
	r.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		log.Println("in default handler")
		writer.WriteHeader(http.StatusOK)
	})

	r.HandleFunc("/api/exec", server.handleExec)
	r.HandleFunc("/api/run/{ID:[a-zA-Z0-9]+}", server.handleRun)

	handler := c.Handler(r)

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
	log.Println("in exec handler")

	decoder := json.NewDecoder(r.Body)

	var input execInput
	err := decoder.Decode(&input)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	log.Println("read your query")

	var output execOutput

	query := models.NewQuery(input.VersionID, input.QueryStr)
	// TODO: move to getlink method
	output.Link = "/run/" + query.Hash.Hex()

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

func (s *pgServer) handleRun(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	runID := vars["ID"]

	log.Printf("runID: %s\n", runID)

	var hashBytes = make([]byte, 100)
	var hash [32]byte

	_, err := hex.Decode(hashBytes, []byte(runID))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	copy(hash[:], hashBytes[:32])

	query := models.QueryFromHash(hash)

	found, err := s.storage.FindRun(query)
	//var found *models.Run = nil
	if err != nil {
		http.Error(
			w,
			fmt.Sprintf("error querying your run: %s", err.Error()),
			http.StatusInternalServerError,
		)
		return
	}

	if found == nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	var output execOutput
	output.Cached = true

	d := json.NewDecoder(strings.NewReader(string(found.Result)))
	d.Decode(&output.Result)

	encoder := json.NewEncoder(w)
	encoder.Encode(output)
}
