package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
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

func WriteRoutesFile(data *bytes.Buffer) error {
	err := os.MkdirAll("gen/config/", os.ModePerm)
	if err != nil {
		return fmt.Errorf("error creating directory gen/config: %w", err)
	}
	return WriteFile("gen/config/routes.rb", data)
}

func WriteActionFiles(actions []ActionDefinition) error {
	// make sure the actions directory exists
	err := os.MkdirAll("gen/actions/", os.ModePerm)
	if err != nil {
		return fmt.Errorf("error creating directory gen/actions: %w", err)
	}
	for _, actionDefinition := range actions {
		actionDirectory := fmt.Sprintf("gen/actions/%s/", toSnake(actionDefinition.ModuleName))
		err = os.MkdirAll(actionDirectory, os.ModePerm)
		if err != nil {
			return fmt.Errorf("error creating action directory %s: %w", actionDirectory, err)
		}

		actionFilePath := fmt.Sprintf("%s%s.rb", actionDirectory, toSnake(actionDefinition.ActionName))
		err = WriteFile(actionFilePath, actionDefinition.GeneratedCode)
		if err != nil {
			return fmt.Errorf("error writing action file %s: %w", actionFilePath, err)
		}
	}
	return nil
}

func WriteServiceFiles(services []ServiceDefinition) error {
	// make sure the actions directory exists
	err := os.MkdirAll("gen/actions/", os.ModePerm)
	if err != nil {
		return fmt.Errorf("error creating directory gen/actions: %w", err)
	}
	for _, serviceDefinition := range services {
		parentDir := fmt.Sprintf("gen/actions/%s/", toSnake(serviceDefinition.ModuleName))
		err = os.MkdirAll(parentDir, os.ModePerm)
		if err != nil {
			return fmt.Errorf("error creating parent directory %s: %w", parentDir, err)
		}

		serviceFilePath := fmt.Sprintf("%s%s.rb", parentDir, toSnake(serviceDefinition.ServiceName))

		fileExists := DoesFileExist(serviceFilePath)
		if fileExists {
			continue // don't write the thing, we don't want to overwrite service files
		}

		err = WriteFile(serviceFilePath, serviceDefinition.GeneratedCode)
		if err != nil {
			return fmt.Errorf("error writing service file %s: %w", serviceFilePath, err)
		}
	}
	return nil
}

func WriteContractsFile(data *bytes.Buffer) error {
	err := os.MkdirAll("gen/actions/", os.ModePerm)
	if err != nil {
		return fmt.Errorf("error creating directory gen/actions: %w", err)
	}
	return WriteFile("gen/actions/contracts.rb", data)
}

func WriteSchemasFile(data *bytes.Buffer) error {
	err := os.MkdirAll("gen/actions/", os.ModePerm)
	if err != nil {
		return fmt.Errorf("error creating directory gen/actions: %w", err)
	}
	return WriteFile("gen/actions/schemas.rb", data)
}

func WriteFile(filePath string, data *bytes.Buffer) error {
	err := os.WriteFile(filePath, data.Bytes(), 0644)
	if err != nil {
		return fmt.Errorf("error writing to file %s: %w", filePath, err)
	}

	return nil
}

func DoesFileExist(filePath string) bool {
	_, err := os.Stat(filePath)

	if errors.Is(err, os.ErrNotExist) {
		return false
	}

	return true
}
