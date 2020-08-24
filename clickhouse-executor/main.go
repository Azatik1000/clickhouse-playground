package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
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

	result, err := s.exec(query)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

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

func main() {
	host := os.Getenv("HOST")

	server, err := NewExecServer(host)
	if err != nil {
		log.Fatal(err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/exec", server.handleExec)

	_ = http.ListenAndServe(":8080", mux)
}
