package main

import (
	"app/clickhouse"
	"app/models"
	"app/storage"
	"encoding/json"
	_ "github.com/ClickHouse/clickhouse-go"
	"github.com/rs/cors"
	"go.uber.org/zap"
	"log"
	"net/http"
	"net/url"
)

type pgServer struct {
	chDriver clickhouse.Driver
	storage  storage.Storage
}

func newPgServer(driver clickhouse.Driver, s storage.Storage) *pgServer {
	return &pgServer{chDriver: driver, storage: s}
}

func main() {
	logger, err := zap.NewDevelopment()
	if err != nil {
		log.Fatal(err)
	}

	sugar := logger.Sugar()
	sugar.Info("starting")

	endpoint, err := url.Parse("http://db:8123")
	if err != nil {
		log.Fatal(err)
	}

	var driver clickhouse.Driver = clickhouse.NewHTTPDriver(endpoint)
	driver.HealthCheck()

	s := storage.NewMemory()
	server := newPgServer(driver, s)

	c := cors.Default()

	mux := http.NewServeMux()
	mux.HandleFunc("/exec", server.handleExec)

	handler := c.Handler(mux)

	_ = http.ListenAndServe(":8080", handler)
}

type execInput struct {
	QueryStr string `json:"query"`
}

type execOutput struct {
	Result string `json:"result"`
}

func (s *pgServer) handleExec(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)

	var input execInput
	err := decoder.Decode(&input)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	query := models.NewQuery(input.QueryStr)
	found := s.storage.FindRun(query.Hash)

	// This query was already run
	if found != nil {

	} else {
		
	}
	//sum := sha256.Sum256([]byte(input.QueryStr))

	result, err := s.chDriver.Exec(input.QueryStr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	encoder := json.NewEncoder(w)

	if err = encoder.Encode(execOutput{Result: result}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
