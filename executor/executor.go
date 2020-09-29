package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/shirou/gopsutil/process"
	"io"
	"log"
	"os/exec"
	"strings"
	"time"
)

type executor struct {
	host  string
	ready chan struct{}
	//mu   sync.Mutex
}

func NewExecutor(host string) (*executor, error) {
	return &executor{
		host: host,
		ready: make(chan struct{}),
	}, nil
}

func (e *executor) runProcess(query string) (string, error) {
	// TODO: mutexes are incorrect
	//e.mu.Lock()
	//defer e.mu.Unlock()

	cmd := exec.Command(
		"clickhouse-client",
		fmt.Sprintf("--host=%s", e.host),
		"-nm",
		"-f",
		"JSON",
	)

	log.Printf("%+v\n", cmd.Args)

	var outb, errb bytes.Buffer
	cmd.Stdout = &outb
	cmd.Stderr = &errb

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return "", err
	}

	log.Println("starting clickhouse-client")

	err = cmd.Start()
	if err != nil {
		return "", err
	}

	log.Printf("wrote \"%s\" to client stdin\n", query)
	_, err = io.WriteString(stdin, query)
	if err != nil {
		return "", err
	}

	log.Println("close stdin")
	// TODO: handle errors
	stdin.Close()

	err = cmd.Wait()
	//if err != nil {
	//	return "", err
	//}

	log.Println("clickhouse-client ended")

	if err == nil {
		return outb.String(), nil
	}

	if _, ok := err.(*exec.ExitError); ok {
		//stderr := string(errExit.Stderr)
		return "", errors.New(errb.String())
	}

	return "", err
}

func (e *executor) exec(query string) ([]map[string]interface{}, error) {
	processOutput, err := e.runProcess(query)

	go func() {
		processes, err := process.Processes()
		if err != nil {
			log.Fatal(err)
		}

		log.Printf("current processes: %+v\n", processes)

		for _, p := range processes {
			cmd, err := p.Cmdline()
			if err != nil {
				log.Fatal(err)
			}

			if cmd == "/pause" || cmd == "/clickhouse-executor" {
				continue
			} else {
				log.Println("gonna kill", p.Pid, cmd)
				p.Kill() // TODO: maybe more gracefully?
			}
		}

		// TODO: something smarter
		time.Sleep(20 * time.Second)
		e.ready <- struct{}{}
	}()

	if err != nil {
		return nil, err
	}

	list := make([]map[string]interface{}, 0)

	data := make(map[string]interface{})
	d := json.NewDecoder(strings.NewReader(processOutput))

	for {
		err = d.Decode(&data)
		if err != nil {
			// TODO: handle EOF and others
			log.Println(err)
			break
			//http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		list = append(list, data)
	}

	//result, err := json.MarshalIndent(list, "", "    ")
	//if err != nil {
	//	return "", err
	//}

	//return string(result), nil

	return list, nil
}
