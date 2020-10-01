package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v12 "k8s.io/client-go/kubernetes/typed/core/v1"
	"log"
	"os/exec"
	"strings"
)

func newWorker(podManager v12.PodInterface, imageVersion string) (*v1.Pod, error) {
	return podManager.Create(&v1.Pod{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			// TODO: fix generatename
			GenerateName: "kek-",
			Namespace:    "default",
			Labels: map[string]string{
				"app":       "clickhouse-playground",
				"component": "worker",
			},
		},
		Spec: v1.PodSpec{
			Hostname: "lollol",
			Containers: []v1.Container{
				{
					Name:                     "clickhouse-server",
					Image:                    fmt.Sprintf("yandex/clickhouse-server:%s", imageVersion),
					Ports:                    nil,
					EnvFrom:                  nil,
					Env:                      nil,
					Resources:                v1.ResourceRequirements{},
					LivenessProbe:            nil,
					ReadinessProbe:           nil,
					Lifecycle:                nil,
					TerminationMessagePath:   "",
					TerminationMessagePolicy: "",
					ImagePullPolicy:          "",
					SecurityContext:          nil,
					Stdin:                    false,
					StdinOnce:                false,
					TTY:                      false,
				},
			},
			RestartPolicy:                 "",
			TerminationGracePeriodSeconds: nil,
			ActiveDeadlineSeconds:         nil,
			EnableServiceLinks:            nil,
			PreemptionPolicy:              nil,
		},
		Status: v1.PodStatus{},
	})
}

func stopWorker(podManager v12.PodInterface, name string) error {
	return podManager.Delete(name, &metav1.DeleteOptions{})
}

type executor struct {
	ready        chan struct{}
	podManager   v12.PodInterface
	worker       *v1.Pod
	imageVersion string
}

func NewExecutor(podManager v12.PodInterface, imageVersion string) (*executor, error) {
	executor := executor{
		// TODO: maybe make buffered?
		ready:        make(chan struct{}),
		podManager:   podManager,
		worker:       nil,
		imageVersion: imageVersion,
	}

	go func() {
		err := executor.setNewWorker()
		if err != nil {
			log.Println(err)
		}
	}()

	return &executor, nil
}

func (e *executor) setNewWorker() error {
	newWorker, err := newWorker(e.podManager, e.imageVersion)
	if err != nil {
		return err
	}

	e.worker = newWorker
	e.ready <- struct{}{}
	return nil
}

func (e *executor) runProcess(query string) (string, error) {
	// TODO: mutexes are incorrect
	//e.mu.Lock()
	//defer e.mu.Unlock()

	// clientset updatestatus
	pod, _ := e.podManager.Get(e.worker.Name, metav1.GetOptions{})

	fmt.Printf("%+v\n", pod.Status)

	cmd := exec.Command(
		"clickhouse-client",
		fmt.Sprintf("--host=%s", pod.Status.PodIP),
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

	go func(podManager v12.PodInterface, name string) {
		err := stopWorker(podManager, name)
		if err != nil {
			log.Println(err)
		}
		//stopWorker(e.podManager, e.)
	}(e.podManager, e.worker.Name)

	go func(e *executor) {
		err := e.setNewWorker()
		if err != nil {
			log.Println(err)
			return
		}

	}(e)
	//go func() {
	//	processes, err := process.Processes()
	//	if err != nil {
	//		log.Fatal(err)
	//	}
	//
	//	log.Printf("current processes: %+v\n", processes)
	//
	//	for _, p := range processes {
	//		cmd, err := p.Cmdline()
	//		if err != nil {
	//			log.Fatal(err)
	//		}
	//
	//		if cmd == "/pause" || cmd == "/clickhouse-executor" {
	//			continue
	//		} else {
	//			log.Println("gonna kill", p.Pid, cmd)
	//			p.Kill() // TODO: maybe more gracefully?
	//		}
	//	}
	//
	//	// TODO: something smarter
	//	time.Sleep(20 * time.Second)
	//	e.ready <- struct{}{}
	//}()

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
