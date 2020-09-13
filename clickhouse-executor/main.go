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
	host string
	mu   sync.Mutex
}

func NewExecServer(host string) (*ExecServer, error) {
	var server ExecServer
	server.host = host
	return &server, nil
}

func (s *ExecServer) exec(query string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	fmt.Println("in exec")

	cmd := exec.Command(
		"clickhouse-client",
		fmt.Sprintf("--host=%s", s.host),
		"-nm",
		"-f",
		"JSON",
	)

	var outb, errb bytes.Buffer
	cmd.Stdout = &outb
	cmd.Stderr = &errb

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return "", err
	}

	fmt.Println("starting clickhouse-client")

	err = cmd.Start()
	if err != nil {
		return "", err
	}

	_, err = io.WriteString(stdin, query)
	if err != nil {
		return "", err
	}

	// TODO: handle errors
	stdin.Close()

	err = cmd.Wait()
	//if err != nil {
	//	return "", err
	//}

	fmt.Println("clickhouse-client ended")

	if err == nil {
		return outb.String(), nil
	}

	if _, ok := err.(*exec.ExitError); ok {
		//stderr := string(errExit.Stderr)
		return "", errors.New(errb.String())
	}

	return "", err
}

type execInput struct {
	QueryStr string `json:"query"`
}

func (s *ExecServer) handleExec(w http.ResponseWriter, r *http.Request) {
	queryBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	query := string(queryBytes)

	fmt.Println("read your query")

	result, err := s.exec(query)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	go func() {
		processes, err := process.Processes()
		if err != nil {
			log.Fatal(err)
		}

		for _, p := range processes {
			cmd, err := p.Cmdline()
			if err != nil {
				log.Fatal(err)
			}

			if cmd == "/pause" || cmd == "/clickhouse-executor" {
				continue
			} else {
				fmt.Println("gonna kill", p.Pid, cmd)
				p.Kill() // TODO: maybe more gracefully?
			}
		}
	}()

	list := make([]map[string]interface{}, 0)

	data := make(map[string]interface{})
	d := json.NewDecoder(strings.NewReader(result))

	for {
		err = d.Decode(&data)
		if err != nil {
			// TODO: handle EOF and others
			fmt.Println(err)
			break
			//http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		list = append(list, data)
	}

	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "    ")
	if err = encoder.Encode(list); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func readFromQueue(conn *amqp.Connection) error {
	ch, err := conn.Channel()
	if err != nil {
		return err
	}

	defer ch.Close()

	q, err := ch.QueueDeclare(
		"hello", // name
		false,   // durable
		false,   // delete when unused
		false,   // exclusive
		false,   // no-wait
		nil,     // arguments
	)
	if err != nil {
		return err
	}

	err = ch.Qos(
		1,     // prefetch count
		0,     // prefetch size
		false, // global
	)

	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		false,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	if err != nil {
		return err
	}

	for d := range msgs {
		log.Printf("Received a message: %s", d.Body)
		time.Sleep(10 * time.Second)
		d.Ack(false)
	}

	return nil
}

func main() {
	//config, err := rest.InClusterConfig()
	//if err != nil {
	//	log.Fatal(err)
	//}
	//
	//_, err = kubernetes.NewForConfig(config)
	//if err != nil {
	//	log.Fatal(err)
	//}

	//for {
	//	pods, err := clientset.CoreV1().Pods("").List(context.TODO(), metav1.ListOptions{})
	//	if err != nil {
	//		log.Fatal(err)
	//	}
	//
	//	fmt.Printf("%+v\n", pods)
	//
	//	time.Sleep(10 * time.Second)
	//}

	s := "amqp://user:DFxoFGS2i3@my-release-rabbitmq:5672"
	conn, err := amqp.Dial(s)
	if err != nil {
		log.Fatal(err)
	}

	defer conn.Close()

	go readFromQueue(conn)

	//host := os.Getenv("HOST")
	host := "localhost"

	server, err := NewExecServer(host)
	if err != nil {
		log.Fatal(err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/exec", server.handleExec)

	_ = http.ListenAndServe(":8080", mux)
}
