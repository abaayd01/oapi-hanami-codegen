package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
)

type exitCode int

const (
	exitOK    exitCode = 0
	exitError exitCode = 1
)

func main() {
	code := mainRun()
	os.Exit(int(code))
}

func mainRun() exitCode {
	config, err := parseArgs()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error parsing args: %s\n", err)
		return exitError
	}

	g, err := NewGenerator(config.inputFilePath, config.appName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create generator: %s\n", err)
		return exitError
	}

	routesFileTemplateModel, err := g.GenerateRoutesFileTemplateModel()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to generate routes file template model: %s\n", err)
		return exitError
	}

	actionTemplateModels, err := g.GenerateActionTemplateModels()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to generate action template models: %s\n", err)
		return exitError
	}

	serviceTemplateModels, err := g.GenerateServiceTemplateModels()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to generate service template models: %s\n", err)
		return exitError
	}

	contractsFileTemplateModel, err := g.GenerateContractsFileTemplateModel()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to generate contracts file template model: %s\n", err)
		return exitError
	}

	schemasFileTemplateModel, err := g.GenerateSchemasFileTemplateModel()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to generate schemas file template model: %s\n", err)
		return exitError
	}

	w, err := NewWriter(config.outputDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create a new writer: %s\n", err)
	}

	err = w.WriteRoutesFileFromModel(routesFileTemplateModel)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to write routes file: %s\n", err)
		return exitError
	}

	err = w.WriteActionFilesFromModels(actionTemplateModels)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to write action files: %s\n", err)
		return exitError
	}

	err = w.WriteServiceFilesFromModels(serviceTemplateModels)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to write service files: %s\n", err)
		return exitError
	}

	err = w.WriteContractsFileFromModel(contractsFileTemplateModel)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to write contracts file: %s\n", err)
		return exitError
	}

	err = w.WriteSchemasFileFromModel(schemasFileTemplateModel)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to write schemas file: %s\n", err)
		return exitError
	}

	return exitOK
}

type args struct {
	inputFilePath string
	appName       string
	outputDir     string
}

func parseArgs() (*args, error) {
	inputFilePtr := flag.String("inputFile", "", "file path of OpenAPI spec")
	appNamePtr := flag.String("appName", "HanamiApp", "name of the top-level Hanami app module")
	outputDirPtr := flag.String("outputDir", "gen", "path to output directory")

	flag.Parse()

	if *inputFilePtr == "" {
		return nil, errors.New("must provide an inputFile")
	}

	return &args{
		inputFilePath: *inputFilePtr,
		appName:       *appNamePtr,
		outputDir:     *outputDirPtr,
	}, nil
}
