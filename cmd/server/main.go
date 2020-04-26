package main

import (
	"encoding/json"
	"github.com/Azatik1000/clickhouse-playground/internal/pkg/clickhouse"
	_ "github.com/ClickHouse/clickhouse-go"
	"go.uber.org/zap"
	"log"
	"net/http"
	"net/url"
)

type pgServer struct {
	chDriver clickhouse.Driver
}

func newPgServer(driver clickhouse.Driver) *pgServer {
	return &pgServer{chDriver: driver}
}

func main() {
	logger, err := zap.NewDevelopment()
	if err != nil {
		log.Fatal(err)
	}

	sugar := logger.Sugar()
	sugar.Info("starting")

	endpoint, err := url.Parse("http://172.19.0.2:8123")
	if err != nil {
		log.Fatal(err)
	}

	var driver clickhouse.Driver = clickhouse.NewHTTPDriver(endpoint)
	driver.HealthCheck()

	server := newPgServer(driver)

	http.HandleFunc("/exec", server.handleExec)
	_ = http.ListenAndServe(":8080", nil)
}

type execInput struct {
	Query string `json:"query"`
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

	result, err := s.chDriver.Exec(input.Query)
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
