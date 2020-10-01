package main

import (
	"bytes"
	"encoding/json"
	"github.com/streadway/amqp"
	"k8s.io/client-go/kubernetes"
	v12 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
	"log"
	"os"
)

type ExecServer struct {
	executor         *executor
	mqConn           *amqp.Connection
	mqChan           *amqp.Channel
	receiveQueueName string
	podManager       v12.PodInterface
}

func NewExecServer(
	mqHost string,
	receiveQueueName string,
	podManager v12.PodInterface,
	imageVersion string,
) (*ExecServer, error) {
	var server ExecServer

	executor, err := NewExecutor(podManager, imageVersion)
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

	server.podManager = podManager

	return &server, nil
}

type responseData struct {
	Success bool                     `json:"success"`
	Error   *string                  `json:"error"`
	Result  []map[string]interface{} `json:"result"`
}

func (s *ExecServer) handleRequestMsg(requestMsg amqp.Delivery) error {
	result, err := s.executor.exec(string(requestMsg.Body))

	var response bytes.Buffer

	encoder := json.NewEncoder(&response)

	if err != nil {
		log.Printf("got exec error: %s", err)
		errStr := err.Error()
		encoder.Encode(responseData{
			Success: false,
			Error:   &errStr,
		})
	} else {
		log.Printf("got result: %s", result)
		encoder.Encode(responseData{
			Success: true,
			Result:  result,
		})
	}

	log.Printf("gonna send %s\n", string(response.Bytes()))

	err = s.mqChan.Publish(
		"",                 // exchange
		requestMsg.ReplyTo, // routing key
		false,              // mandatory
		false,              // immediate
		amqp.Publishing{
			ContentType:   "application/json",
			CorrelationId: requestMsg.CorrelationId,
			Body:          response.Bytes(),
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

	for ; ; {
		<-s.executor.ready
		requestMsg := <-requestMsgs
		s.handleRequestMsg(requestMsg)
	}

	return nil
}

func main() {
	receiveQueueName := os.Getenv("RECEIVE_QUEUE_NAME")
	if receiveQueueName == "" {
		log.Fatal("didn't specify RECEIVE_QUEUE_NAME")
	}

	log.Printf("RECEIVE_QUEUE_NAME=%s\n", receiveQueueName)

	imageVersion := os.Getenv("IMAGE_VERSION")
	if imageVersion == "" {
		log.Fatal("didn't specify IMAGE_VERSION")
	}

	// creates the in-cluster config
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}
	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	// TODO: move user/pass to config
	server, err := NewExecServer(
		"amqp://test:test@my-release-rabbitmq:5672",
		receiveQueueName,
		clientset.CoreV1().Pods("default"),
		imageVersion,
	)

	if err != nil {
		log.Fatal(err)
	}

	// TODO: change to smth smarter
	//time.Sleep(15 * time.Second)
	// infinite
	server.handleRequests()
}
