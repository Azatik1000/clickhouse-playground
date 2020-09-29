package driver

import (
	"app/clickhouse"
	"app/models"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/streadway/amqp"
	"log"
	"sync"
)

type request struct {
	id        uuid.UUID
	query     string
	versionID string
	result    chan responseData
}

type executorDriver struct {
	mqConn          *amqp.Connection
	mqChan          *amqp.Channel
	requestsMu      sync.Mutex
	resultQueueName string
	requests        map[uuid.UUID]*request
}

func sendQueueName(versionID string) string {
	return fmt.Sprintf("execute-%s", versionID)
}

func (d *executorDriver) declareQueue(name string) error {
	_, err := d.mqChan.QueueDeclare(
		name,
		false,
		false,
		false,
		false,
		nil,
	)
	return err
}

func NewExecutor(mqHost string, resultQueueName string) (Driver, error) {
	var driver executorDriver

	mqConn, err := amqp.Dial(mqHost)
	if err != nil {
		log.Fatal(err)
	}
	driver.mqConn = mqConn

	ch, err := mqConn.Channel()
	if err != nil {
		return nil, err
	}
	driver.mqChan = ch

	for _, version := range clickhouse.Versions {
		if err = driver.declareQueue(sendQueueName(version.Id)); err != nil {
			return nil, err
		}

		log.Printf("successfully declared queue %s\n", sendQueueName(version.Id))
	}

	if err = driver.declareQueue(resultQueueName); err != nil {
		return nil, err
	}
	driver.resultQueueName = resultQueueName

	driver.requests = make(map[uuid.UUID]*request)

	go func() {
		err := driver.handleResponses()
		if err != nil {
			log.Println(err)
		}
	}()

	return &driver, nil
}

type responseData struct {
	Success bool                     `json:"success"`
	Error   *string                  `json:"error"`
	Result  []map[string]interface{} `json:"result"`
}

func (d *executorDriver) handleResponseMsg(responseMsg amqp.Delivery) error {
	log.Printf("got correlationId=%s\n", responseMsg.CorrelationId)

	id, err := uuid.Parse(responseMsg.CorrelationId)
	if err != nil {
		return err
	}

	log.Printf("got response with id=%s\n", id.String())

	d.requestsMu.Lock()
	var result chan responseData
	if _, ok := d.requests[id]; ok {
		result = d.requests[id].result
	}
	d.requestsMu.Unlock()

	if result != nil {
		decoder := json.NewDecoder(bytes.NewReader(responseMsg.Body))

		var response responseData
		decoder.Decode(&response)

		result <- response
	}
	// TODO: remove from map

	return nil
}

func (d *executorDriver) handleResponses() error {
	responseMsgs, err := d.mqChan.Consume(
		d.resultQueueName, // queue
		"",                          // consumer
		true,                        // auto-ack
		false,                       // exclusive
		false,                       // no-local
		false,                       // no-wait
		nil,                         // args
	)
	if err != nil {
		return err
	}

	for responseMsg := range responseMsgs {
		err := d.handleResponseMsg(responseMsg)
		if err != nil {
			return err
		}
	}

	return nil
}

func (d *executorDriver) newRequest(query string, versionID string) *request {
	request := &request{
		id:        uuid.New(),
		query:     query,
		versionID: versionID,
		result:    make(chan responseData, 1),
	}

	d.requestsMu.Lock()
	d.requests[request.id] = request
	d.requestsMu.Unlock()

	return request
}

func (d *executorDriver) sendRequest(request *request) error {
	log.Printf("sending with id=%s\n", request.id.String())

	defer log.Printf("published message to queue %s\n", sendQueueName(request.versionID))
	return d.mqChan.Publish(
		"",                               // exchange
		sendQueueName(request.versionID), // routing key
		false,                            // mandatory
		false,                            // immediate
		amqp.Publishing{
			ContentType:   "text/plain",
			CorrelationId: request.id.String(),
			ReplyTo:       d.resultQueueName,
			Body:          []byte(request.query),
		},
	)
}

func (d *executorDriver) Exec(query string, versionID string) (*models.Result, error) {
	request := d.newRequest(query, versionID)

	err := d.sendRequest(request)
	if err != nil {
		return nil, err
	}

	log.Println("waiting on channel")
	result := <-request.result
	log.Println("got result from channel")

	if !result.Success {
		return nil, errors.New(*result.Error)
	}

	var jsonResult bytes.Buffer
	encoder := json.NewEncoder(&jsonResult)
	encoder.Encode(result.Result)

	jsonString := jsonResult.String()

	return (*models.Result)(&jsonString), nil
}

func (d *executorDriver) HealthCheck() error {
	// TODO: change
	return nil
}

func (d *executorDriver) Close() error {
	d.mqChan.Close()
	return d.mqConn.Close()
}
