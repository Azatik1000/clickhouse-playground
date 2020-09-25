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

	driver.requests = make(map[uuid.UUID]*request)

	go func() {
		err := driver.handleResponses()
		if err != nil {
			log.Println(err)
		}
	}()

	return &driver, nil
}

func (d *executorDriver) handleResponseMsg(responseMsg amqp.Delivery) error {
	log.Printf("got correlationId=%s\n", responseMsg.CorrelationId)
	id, err := uuid.Parse(responseMsg.CorrelationId)
	if err != nil {
		return err
	}

	log.Printf("got response with id=%s\n", id.String())

	d.requests[id].result <- models.Result(responseMsg.Body)
	// TODO: remove from map

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
		err := d.handleResponseMsg(responseMsg)
		if err != nil {
			return err
		}
	}

	return nil
}

func (d *executorDriver) newRequest(query string) *request {
	request := &request{
		id: uuid.New(),
		query: query,
		result: make(chan models.Result),
	}

	d.requests[request.id] = request
	return request
}

func (d *executorDriver) sendRequest(request *request) error {
	log.Printf("sending with id=%s\n", request.id.String())

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

	log.Println("waiting on channel")
	result := <-request.result
	log.Println("got result from channel")
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
