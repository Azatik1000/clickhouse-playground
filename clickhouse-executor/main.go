package main

import (
	"bytes"
	"github.com/shirou/gopsutil/process"
	"github.com/streadway/amqp"
	"time"

	//"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	//metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"log"
	"net/http"
	"os/exec"
	"strings"
	"sync"
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

func (s *ExecServer) handleRequests() error {
	
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


}
