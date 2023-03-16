package main

import (
	"flag"
	"log"
)

type args struct {
	inputFilePath string
	appName       string
}

func parseArgs() args {
	inputFilePtr := flag.String("inputFile", "", "file path of OpenAPI spec")
	appNamePtr := flag.String("appName", "HanamiApp", "name of the top-level Hanami app module")

	flag.Parse()

	return args{
		inputFilePath: *inputFilePtr,
		appName:       *appNamePtr,
	}
}

func main() {
	var err error
	defer func() {
		if err != nil {
			log.Fatalln(err)
		}
	}()

	config := parseArgs()

	g, err := NewGenerator(config.inputFilePath, config.appName)
	if err != nil {
		return
	}

	routesFileBuf, err := g.GenerateRoutesFile()
	if err != nil {
		return
	}

	err = WriteRoutesFile(routesFileBuf)
	if err != nil {
		return
	}

	actionFileBufs, err := g.GenerateActionFiles()
	if err != nil {
		return
	}
	err = WriteActionFiles(actionFileBufs)
	if err != nil {
		return
	}

	serviceFileBufs, err := g.GenerateServiceFiles()
	if err != nil {
		return
	}
	err = WriteServiceFiles(serviceFileBufs)

	contractsFileBuf, err := g.GenerateContractsFile()
	if err != nil {
		return
	}
	err = WriteContractsFile(contractsFileBuf)

	schemasFile, err := g.GenerateSchemasFile()
	if err != nil {
		return
	}
	err = WriteSchemasFile(schemasFile)
}
