package main

import (
	"bytes"
	"fmt"
	"gopkg.in/yaml.v2"
	"log"
	"os"
	"text/template"
)

//go:generate go run .

const (
	pipeline = "pipeline.yml"

	pipelineTemplate  = "pipeline.yml.tpl"
	platformsTemplate = "platforms.yml.tpl"
	helpers           = "_helpers.tpl"

	pipelineSrc  = "pipeline.src.yml"
	platformsSrc = "platforms.src.yml"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	if err := executeFile(pipelineSrc, platformsSrc, platformsTemplate); err != nil {
		return fmt.Errorf("cannot generate pipeline template values: %w", err)
	}

	if err := executeFile(pipeline, pipelineSrc, pipelineTemplate, helpers); err != nil {
		return fmt.Errorf("cannot generate pipeline: %w", err)
	}

	return nil
}

func executeFile(targetFilename, valuesFilename string, goTplFilenames ...string) error {
	log.Printf("reading %v", valuesFilename)
	values, err := os.ReadFile(valuesFilename)
	if err != nil {
		return fmt.Errorf("failed to read %v: %w", valuesFilename, err)
	}

	result, err := execute(values, goTplFilenames...)
	if err != nil {
		return err
	}

	log.Printf("writing %v", targetFilename)
	if err := os.WriteFile(targetFilename, result, 0644); err != nil {
		return fmt.Errorf("failed to write %v: %w", targetFilename, err)
	}

	return nil
}

func execute(values []byte, goTplFiles ...string) ([]byte, error) {
	log.Printf("parsing %v", goTplFiles)
	tpl, err := template.ParseFiles(goTplFiles...)
	if err != nil {
		return nil, fmt.Errorf("failed to parse files: %w", err)
	}

	var data map[string]interface{}
	if err := yaml.Unmarshal(values, &data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal values: %w", err)
	}

	out := &bytes.Buffer{}
	if err := tpl.Execute(out, data); err != nil {
		return nil, fmt.Errorf("failed to eecute: %w", err)
	}

	return out.Bytes(), nil
}
