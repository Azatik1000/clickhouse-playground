package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/streadway/amqp"
	"k8s.io/client-go/kubernetes"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	"log"
	"os"
	"time"
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

	for requestMsg := range requestMsgs {
		s.handleRequestMsg(requestMsg)
		<-s.executor.ready
	}

	return nil
}

func main() {
	receiveQueueName := os.Getenv("RECEIVE_QUEUE_NAME")
	if receiveQueueName == "" {
		log.Fatal("didn't specify RECEIVE_QUEUE_NAME")
	}

	log.Printf("RECEIVE_QUEUE_NAME=%s\n", receiveQueueName)

	log.Printf("MY_POD_NAME=%s\n", os.Getenv("MY_POD_NAME"))

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

	//clientset.CoreV1().Pods("").Get("kek")
	for {
		// get pods in all the namespaces by omitting namespace
		// Or specify namespace to get pods in particular namespace
		pods, err := clientset.CoreV1().Pods("").List(metav1.ListOptions{})
		if err != nil {
			panic(err.Error())
		}
		fmt.Printf("There are %d pods in the cluster\n", len(pods.Items))

		//// Examples for error handling:
		//// - Use helper functions e.g. errors.IsNotFound()
		//// - And/or cast to StatusError and use its properties like e.g. ErrStatus.Message
		//_, err = clientset.CoreV1().Pods("default").Get("example-xxxxx", metav1.GetOptions{})
		//if errors.IsNotFound(err) {
		//	fmt.Printf("Pod example-xxxxx not found in default namespace\n")
		//} else if statusError, isStatus := err.(*errors.StatusError); isStatus {
		//	fmt.Printf("Error getting pod %v\n", statusError.ErrStatus.Message)
		//} else if err != nil {
		//	panic(err.Error())
		//} else {
		//	fmt.Printf("Found example-xxxxx pod in default namespace\n")
		//}

		time.Sleep(10 * time.Second)
	}

	// TODO: move user/pass to config
	server, err := NewExecServer(
		"localhost",
		"amqp://user:4Pb4iaav1K@my-release-rabbitmq:5672",
		receiveQueueName,
	)

	if err != nil {
		log.Fatal(err)
	}

	// TODO: change to smth smarter
	time.Sleep(15 * time.Second)
	// infinite
	server.handleRequests()
}
