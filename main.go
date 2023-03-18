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

	routesFileBuf, err := g.GenerateRoutesFile()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to generate routes file: %s\n", err)
		return exitError
	}

	actionFileBufs, err := g.GenerateActionFiles()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to generate action files: %s\n", err)
		return exitError
	}

	serviceFileBufs, err := g.GenerateServiceFiles()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to generate service files: %s\n", err)
		return exitError
	}

	contractsFileBuf, err := g.GenerateContractsFile()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to generate contracts file: %s\n", err)
		return exitError
	}

	schemasFile, err := g.GenerateSchemasFile()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to generate schemas file: %s\n", err)
		return exitError
	}

	w := NewWriter(config.outputDir)

	err = w.WriteRoutesFile(routesFileBuf)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to write routes file: %s\n", err)
		return exitError
	}

	err = w.WriteActionFiles(actionFileBufs)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to write action files: %s\n", err)
		return exitError
	}

	err = w.WriteServiceFiles(serviceFileBufs)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to write service files: %s\n", err)
		return exitError
	}

	err = w.WriteContractsFile(contractsFileBuf)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to write contracts file: %s\n", err)
		return exitError
	}

	err = w.WriteSchemasFile(schemasFile)
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
