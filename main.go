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

	g, err := NewGenerator(config.inputFilePath, config.appName, config.sliceName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create generator: %s\n", err)
		return exitError
	}

	templateModels, err := g.GenerateTemplateModels()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to generate template models: %s\n", err)
		return exitError
	}

	w, err := NewWriter(config.outputDir, config.appName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create a new writer: %s\n", err)
		return exitError
	}

	err = w.WriteFilesFromTemplateModels(templateModels)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to write files from template models: %s\n", err)
		return exitError
	}

	return exitOK
}

type args struct {
	inputFilePath string
	appName       string
	sliceName     string
	outputDir     string
}

func parseArgs() (*args, error) {
	inputFilePtr := flag.String("inputFile", "", "file path of OpenAPI spec")
	appNamePtr := flag.String("appName", "HanamiApp", "name of the top-level Hanami app module")
	sliceNamePtr := flag.String("sliceName", "API", "name of the slice you want to put your generated actions in")
	outputDirPtr := flag.String("outputDir", "gen", "path to output directory")

	flag.Parse()

	if *inputFilePtr == "" {
		return nil, errors.New("must provide an inputFile")
	}

	return &args{
		inputFilePath: *inputFilePtr,
		appName:       *appNamePtr,
		sliceName:     *sliceNamePtr,
		outputDir:     *outputDirPtr,
	}, nil
}
