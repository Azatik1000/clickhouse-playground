package main

import (
	"app/models"
	"app/storage"
	"app/worker"
	"encoding/json"
	_ "github.com/ClickHouse/clickhouse-go"
	"github.com/rs/cors"
	"go.uber.org/zap"
	"log"
	"net/http"
	"strings"
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

	s, err := storage.NewDatabase()
	if err != nil {
		log.Fatal(err)
	}

	pool, err := worker.NewPool(2)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("created pool successfully")

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

	var output execOutput

	query := models.NewQuery(input.QueryStr)
	// TODO: move to getlink method
	output.Link = "/runs/" + query.Hash.Hex()
	found, err := s.storage.FindRun(query)

	var result models.Result

	// This query 's been already run
	if found != nil {
		output.Cached = true
		result = found.Result
	} else {
		output.Cached = false

		result, err = s.pool.Execute(input.QueryStr)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// TODO: pointers ok?
		s.storage.AddRun(&models.Run{
			Query:  *query,
			Result: result,
		})
	}

	d := json.NewDecoder(strings.NewReader(string(result)))
	d.Decode(&output.Result)

	encoder := json.NewEncoder(w)
	encoder.Encode(output)
	//if _, err = io.WriteString(w, string(result)); err != nil {
	//	http.Error(w, err.Error(), http.StatusInternalServerError)
	//}

	//encoder := json.NewEncoder(w)
	//
	//if err = encoder.Encode(execOutput{Result: result}); err != nil {
	//	http.Error(w, err.Error(), http.StatusInternalServerError)
	//	return
	//}
}
