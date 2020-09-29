package main

import (
	"gopkg.in/yaml.v2"
	"log"
	"os"
	"path"
	"text/template"
)

type Version struct {
	Id           string `yaml:"id"`
	ImageVersion string `yaml:"imageVersion"`
}

var executorDeploymentTemplateFileName = "../executor-deployment-template.yaml"
var executorDeploymentsFileName = "../executor-deployments.yaml"

var autoscalerTemplateFileName = "../autoscaler-template.yaml"
var autoscalersFileName = "../autoscalers.yaml"

func main() {
	configFile, err := os.OpenFile("../config/versions.yaml", os.O_RDONLY, 0)
	if err != nil {
		log.Fatal(err)
	}

	decoder := yaml.NewDecoder(configFile)

	data := struct {
		Versions []Version `yaml:"versions"`
	}{}

	if err := decoder.Decode(&data); err != nil {
		log.Fatal(err)
	}

	executorDeploymentTemplate, err := template.New(path.Base(executorDeploymentTemplateFileName)).
		ParseFiles(executorDeploymentTemplateFileName)
	executorDeploymentsFile, err := os.OpenFile(
		executorDeploymentsFileName,
		os.O_WRONLY|os.O_TRUNC|os.O_CREATE,
		0644,
	)
	if err := executorDeploymentTemplate.Execute(executorDeploymentsFile, data); err != nil {
		log.Fatal(err)
	}

	autoscalerTemplate, err := template.New(path.Base(autoscalerTemplateFileName)).
		ParseFiles(autoscalerTemplateFileName)
	autoscalersFile, err := os.OpenFile(
		autoscalersFileName,
		os.O_WRONLY|os.O_TRUNC|os.O_CREATE,
		0644,
	)
	if err := autoscalerTemplate.Execute(autoscalersFile, data); err != nil {
		log.Fatal(err)
	}
}
