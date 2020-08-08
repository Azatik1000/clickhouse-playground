package main

import (
	"app/models"
	"app/storage"
	"app/worker"
	"encoding/json"
	"fmt"
	_ "github.com/ClickHouse/clickhouse-go"
	"github.com/rs/cors"
	"go.uber.org/zap"
	"log"
	"net/http"
)

type pgServer struct {
	pool    *worker.Pool
	storage storage.Storage
}

func newPgServer(pool *worker.Pool, s storage.Storage) *pgServer {
	return &pgServer{pool: pool, storage: s}
}

func main() {
	logger, err := zap.NewDevelopment()
	if err != nil {
		log.Fatal(err)
	}

	sugar := logger.Sugar()
	sugar.Info("starting")

	s := storage.NewMemory()
	pool, err := worker.NewPool(2)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("created pool successfully")

	if err = pool.Start(); err != nil {
		log.Fatal(err)
	}

	log.Println("started pool successfully")

	server := newPgServer(pool, s)

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
	fmt.Println("exec request")

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

	result, err := s.pool.Execute(input.QueryStr)
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
