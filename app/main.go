package main

import (
	"app/driver"
	"app/models"
	"app/storage"
	"encoding/json"
	"fmt"
	_ "github.com/ClickHouse/clickhouse-go"
	"github.com/rs/cors"
	"github.com/streadway/amqp"
	"go.uber.org/zap"
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

func sendToQueue(conn *amqp.Connection) error {
	ch, err := conn.Channel()
	if err != nil {
		return err
	}

	defer ch.Close()

	q, err := ch.QueueDeclare(
		"hello",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return err
	}

	body := "Hello World!"

	for i := 0; i < 500; i++ {
		err = ch.Publish(
			"",     // exchange
			q.Name, // routing key
			false,  // mandatory
			false,  // immediate
			amqp.Publishing{
				ContentType: "text/plain",
				Body:        []byte(body),
			},
		)

		if err != nil {
			return err
		}
	}

	return nil
}

func main() {
	logger, err := zap.NewDevelopment()
	if err != nil {
		log.Fatal(err)
	}

	sugar := logger.Sugar()
	sugar.Info("starting")

	s := storage.NewMemory()
	//if err != nil {
	//	log.Fatal(err)
	//}

	driver, err := driver.NewExecutor("amqp://user:DFxoFGS2i3@my-release-rabbitmq:5672")
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

	fmt.Println("read your query")

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

		fmt.Println("sending your query to driver")

		result, err = s.driver.Exec(input.QueryStr)
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
