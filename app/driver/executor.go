package driver

import (
	"app/models"
	"github.com/google/uuid"
	"github.com/streadway/amqp"
	"log"
)

type request struct {
	id     uuid.UUID
	query  string
	result chan models.Result
}

type executorDriver struct {
	mqConn           *amqp.Connection
	mqChan           *amqp.Channel
	sendQueueName    string
	receiveQueueName string
	requests         map[uuid.UUID]*request
}

func NewExecutor(mqHost string, sendQueueName string) (Driver, error) {
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

	_, err = ch.QueueDeclare(
		sendQueueName,
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return nil, err
	}
	driver.sendQueueName = sendQueueName

	receiveQueue, err := ch.QueueDeclare(
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return nil, err
	}
	driver.receiveQueueName = receiveQueue.Name

	go driver.handleResponses()

	return &driver, nil
}

func (d *executorDriver) handleResponseMsg(responseMsg amqp.Delivery) error {
	id, err := uuid.FromBytes([]byte(responseMsg.CorrelationId))
	if err != nil {
		return err
	}

	d.requests[id].result <- models.Result(responseMsg.Body)
	return nil
}

func (d *executorDriver) handleResponses() error {
	responseMsgs, err := d.mqChan.Consume(
		d.receiveQueueName, // queue
		"",                 // consumer
		true,               // auto-ack
		false,              // exclusive
		false,              // no-local
		false,              // no-wait
		nil,                // args
	)
	if err != nil {
		return err
	}

	for responseMsg := range responseMsgs {
		d.handleResponseMsg(responseMsg)
	}

	return nil
}

func (d *executorDriver) newRequest(query string) *request {
	request := &request{query: query, result: make(chan models.Result)}

	id := uuid.New()
	d.requests[id] = request

	return request
}

func (d *executorDriver) sendRequest(request *request) error {
	return d.mqChan.Publish(
		"",              // exchange
		d.sendQueueName, // routing key
		false,           // mandatory
		false,           // immediate
		amqp.Publishing{
			ContentType:   "text/plain",
			CorrelationId: request.id.String(),
			ReplyTo:       d.receiveQueueName,
			Body:          []byte(request.query),
		},
	)
}

func (d *executorDriver) Exec(query string) (models.Result, error) {
	request := d.newRequest(query)

	err := d.sendRequest(request)
	if err != nil {
		return "", err
	}

	result := <-request.result
	return result, nil
}

func (d *executorDriver) HealthCheck() error {
	// TODO: change
	return nil
}

func (d *executorDriver) Close() error {
	d.mqChan.Close()
	return d.mqConn.Close()
}
