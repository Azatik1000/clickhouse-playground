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
)

type executor struct {
	host string
	//mu   sync.Mutex
}

func NewExecutor(host string) (*executor, error) {
	return &executor{
		host: host,
	}, nil
}

func (e *executor) runProcess(query string) (string, error) {
	// TODO: mutexes are incorrect
	//e.mu.Lock()
	//defer e.mu.Unlock()

	fmt.Println("in exec")

	cmd := exec.Command(
		"clickhouse-client",
		fmt.Sprintf("--host=%e", e.host),
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

func (e *executor) exec(query string) (string, error) {
	processOutput, err := e.runProcess(query)
	if err != nil {
		return "", err
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
	d := json.NewDecoder(strings.NewReader(processOutput))

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

	result, err := json.MarshalIndent(list, "", "    ")
	if err != nil {
		return "", err
	}

	return string(result), nil
}

