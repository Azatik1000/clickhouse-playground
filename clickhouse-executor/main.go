package main

import (
	"github.com/streadway/amqp"
	"log"
)

type ExecServer struct {
	executor         *executor
	mqConn           *amqp.Connection
	mqChan           *amqp.Channel
	receiveQueueName string
}

func NewExecServer(
	host string,
	mqHost string,
	receiveQueueName string,
) (*ExecServer, error) {
	var server ExecServer

	executor, err := NewExecutor(host)
	if err != nil {
		return nil, err
	}
	server.executor = executor

	mqConn, err := amqp.Dial(mqHost)
	if err != nil {
		log.Fatal(err)
	}
	server.mqConn = mqConn

	ch, err := mqConn.Channel()
	if err != nil {
		return nil, err
	}
	server.mqChan = ch

	_, err = ch.QueueDeclare(
		receiveQueueName,
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return nil, err
	}
	server.receiveQueueName = receiveQueueName

	return &server, nil
}

func (s *ExecServer) handleRequestMsg(requestMsg amqp.Delivery) error {
	// TODO: handle error
	result, _ := s.executor.exec(string(requestMsg.Body))

	err := s.mqChan.Publish(
		"",                 // exchange
		requestMsg.ReplyTo, // routing key
		false,              // mandatory
		false,              // immediate
		amqp.Publishing{
			ContentType:   "application/json",
			CorrelationId: requestMsg.CorrelationId,
			Body:          []byte(result),
		},
	)
	if err != nil {
		return err
	}

	requestMsg.Ack(false)
	return nil
}

func (s *ExecServer) handleRequests() error {
	err := s.mqChan.Qos(
		1,     // prefetch count
		0,     // prefetch size
		false, // global
	)
	if err != nil {
		return err
	}

	requestMsgs, err := s.mqChan.Consume(
		s.receiveQueueName,
		"",    // consumer
		false, // auto-ack
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,   // args
	)

	for requestMsg := range requestMsgs {
		s.handleRequestMsg(requestMsg)
	}

	return nil
}

func main() {
	server, err := NewExecServer(
		"localhost",
		"amqp://user:DFxoFGS2i3@my-release-rabbitmq:5672",
		"hello",
	)

	if err != nil {
		log.Fatal(err)
	}

	// infinite
	server.handleRequests()
}
