package main

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"strings"
)

type Writer struct {
	OutputDir string
}

func NewWriter(outputDir string) Writer {
	trimmedOutputDir := strings.Trim(outputDir, "/")
	return Writer{
		OutputDir: trimmedOutputDir,
	}
}

func (w Writer) WriteRoutesFile(data *bytes.Buffer) error {
	err := w.createConfigDirIfNotExists()
	if err != nil {
		return err
	}
	return writeFile(w.OutputDir+"/config/routes.rb", data)
}

func (w Writer) WriteActionFiles(actions []ActionDefinition) error {
	err := w.createActionsDirIfNotExists()
	if err != nil {
		return err
	}

	for _, actionDefinition := range actions {
		actionDirectory := fmt.Sprintf("%s/actions/%s/", w.OutputDir, toSnake(actionDefinition.ModuleName))
		err = os.MkdirAll(actionDirectory, os.ModePerm)
		if err != nil {
			return fmt.Errorf("error creating action directory %s: %w", actionDirectory, err)
		}

		actionFilePath := fmt.Sprintf("%s%s.rb", actionDirectory, toSnake(actionDefinition.ActionName))
		err = writeFile(actionFilePath, actionDefinition.GeneratedCode)
		if err != nil {
			return fmt.Errorf("error writing action file %s: %w", actionFilePath, err)
		}
	}
	return nil
}

func (w Writer) WriteServiceFiles(services []ServiceDefinition) error {
	err := w.createActionsDirIfNotExists()
	if err != nil {
		return err
	}

	for _, serviceDefinition := range services {
		parentDir := fmt.Sprintf("%s/actions/%s/", w.OutputDir, toSnake(serviceDefinition.ModuleName))
		err = os.MkdirAll(parentDir, os.ModePerm)
		if err != nil {
			return fmt.Errorf("error creating parent directory %s: %w", parentDir, err)
		}

		serviceFilePath := fmt.Sprintf("%s%s.rb", parentDir, toSnake(serviceDefinition.ServiceName))

		fileExists := doesFileExist(serviceFilePath)
		if fileExists {
			// don't write the thing, we don't want to overwrite service files if they already exist
			continue
		}

		err = writeFile(serviceFilePath, serviceDefinition.GeneratedCode)
		if err != nil {
			return fmt.Errorf("error writing service file %s: %w", serviceFilePath, err)
		}
	}
	return nil
}

func (w Writer) WriteContractsFile(data *bytes.Buffer) error {
	err := w.createActionsDirIfNotExists()
	if err != nil {
		return err
	}

	return writeFile(w.OutputDir+"/actions/contracts.rb", data)
}

func (w Writer) WriteSchemasFile(data *bytes.Buffer) error {
	err := w.createActionsDirIfNotExists()
	if err != nil {
		return err
	}

	return writeFile(w.OutputDir+"/actions/schemas.rb", data)
}

func (w Writer) createConfigDirIfNotExists() error {
	err := os.MkdirAll(w.OutputDir+"/config/", os.ModePerm)
	if err != nil {
		return fmt.Errorf("error creating directory %s/config: %w", w.OutputDir, err)
	}
	return nil
}

func (w Writer) createActionsDirIfNotExists() error {
	err := os.MkdirAll(w.OutputDir+"/actions/", os.ModePerm)
	if err != nil {
		return fmt.Errorf("error creating directory %s/actions: %w", w.OutputDir, err)
	}
	return nil
}

func writeFile(filePath string, data *bytes.Buffer) error {
	err := os.WriteFile(filePath, data.Bytes(), 0644)
	if err != nil {
		return fmt.Errorf("error writing to file %s: %w", filePath, err)
	}

	return nil
}

func doesFileExist(filePath string) bool {
	_, err := os.Stat(filePath)

	if errors.Is(err, os.ErrNotExist) {
		return false
	}

	return true
}
