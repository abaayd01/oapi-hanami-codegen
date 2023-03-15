package main

import (
	"bufio"
	"bytes"
	"embed"
	"errors"
	"flag"
	"fmt"
	"github.com/deepmap/oapi-codegen/pkg/codegen"
	"github.com/deepmap/oapi-codegen/pkg/util"
	"github.com/getkin/kin-openapi/openapi3"
	"log"
	"os"
	"regexp"
	"text/template"
)

type args struct {
	openAPISpecFilePath string
	appName             string
}

//go:embed templates/* templates/fragments/*
var templatesFS embed.FS

func parseArgs() args {
	inputFilePtr := flag.String("inputFile", "", "file path of OpenAPI spec")
	appNamePtr := flag.String("appName", "HanamiApp", "name of the top-level Hanami app module")

	flag.Parse()

	return args{
		openAPISpecFilePath: *inputFilePtr,
		appName:             *appNamePtr,
	}
}

// hmm inject these into generator?
var templatesFilePath = "templates"
var routesTemplateFileName = "routes.rb.tmpl"
var actionsTemplateFileName = "action.rb.tmpl"

func main() {
	var err error
	defer func() {
		if err != nil {
			log.Fatalln(err)
		}
	}()

	config := parseArgs()

	swagger, err := LoadSwagger(config.openAPISpecFilePath)
	if err != nil {
		return
	}

	operationDefinitions, err := codegen.OperationDefinitions(swagger)
	if err != nil {
		return
	}

	// todo: write a func to load these templates
	templates, err := template.New("routes").Funcs(TemplateFunctions).ParseFS(
		templatesFS,
		templatesFilePath+"/"+routesTemplateFileName,
		templatesFilePath+"/"+actionsTemplateFileName,
	)
	if err != nil {
		return
	}

	g := Generator{
		AppName:              config.appName,
		OperationDefinitions: operationDefinitions,
		Swagger:              swagger,
		Templates:            templates,
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

type Generator struct {
	// put extra config and stuff in here
	AppName              string
	OperationDefinitions []codegen.OperationDefinition
	Swagger              *openapi3.T
	Templates            *template.Template
}

func LoadSwagger(filePath string) (*openapi3.T, error) {
	swagger, err := util.LoadSwagger(filePath)
	if err != nil {
		return nil, fmt.Errorf("error loading swagger spec: %w", err)
	}

	config := codegen.Configuration{
		PackageName: "main",
		Generate: codegen.GenerateOptions{
			EchoServer:   true,
			Client:       true,
			Models:       true,
			EmbeddedSpec: true,
		},
	}

	// todo fix this dependency, don't want to have to call this to get things to work
	_, err = codegen.Generate(swagger, config)
	if err != nil {
		return nil, fmt.Errorf("error executing codegen.Generate: %w", err)
	}

	return swagger, nil
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

var TemplateFunctions = merge(codegen.TemplateFunctions, template.FuncMap{
	"toSnake": toSnake,
})

type RoutesFileTemplateModel struct {
	AppName string
	Routes  []RouteTemplateModel
}

type RouteTemplateModel struct {
	Path                string
	OperationDefinition codegen.OperationDefinition
}

// toRackPath converts a path definition as given by OpenAPI spec to something Rack understands.
// For example "/users/{user_id}" -> "/users/:user_id"
func toRackPath(codegenPath string) string {
	re := regexp.MustCompile("{(.*?)}")
	out := re.ReplaceAllString(codegenPath, ":$1")
	return out
}

func NewRoutesFileTemplateModel(appName string, operationDefinitions []codegen.OperationDefinition) (*RoutesFileTemplateModel, error) {
	var routeTemplateModels []RouteTemplateModel
	for _, op := range operationDefinitions {
		routeTemplateModels = append(routeTemplateModels, RouteTemplateModel{
			Path:                toRackPath(op.Path),
			OperationDefinition: op,
		})
	}

	return &RoutesFileTemplateModel{
		AppName: appName,
		Routes:  routeTemplateModels,
	}, nil
}

func (g Generator) GenerateRoutesFile() (*bytes.Buffer, error) {
	routesFileTemplateModel, err := NewRoutesFileTemplateModel(g.AppName, g.OperationDefinitions)
	if err != nil {
		return nil, fmt.Errorf("error generating routes file template model: %w", err)
	}
	return g.ExecuteRoutesFileTemplate(routesFileTemplateModel)
}

func (g Generator) ExecuteRoutesFileTemplate(model *RoutesFileTemplateModel) (*bytes.Buffer, error) {
	return ExecuteTemplate(g.Templates, routesTemplateFileName, model)
}

func ExecuteTemplate(tmpl *template.Template, filePath string, model any) (*bytes.Buffer, error) {
	var buf bytes.Buffer
	w := bufio.NewWriter(&buf)
	if err := tmpl.ExecuteTemplate(w, filePath, model); err != nil {
		return nil, fmt.Errorf("error executing %s template: %w", filePath, err)
	}
	if err := w.Flush(); err != nil {
		return nil, fmt.Errorf("error flushing output buffer: %w", err)
	}

	return &buf, nil
}

type ActionTemplateModel struct {
	AppName    string
	ActionName string
	ModuleName string
}

func NewActionTemplateModel(appName string, operationDefinition codegen.OperationDefinition) ActionTemplateModel {
	return ActionTemplateModel{
		AppName:    appName,
		ActionName: operationDefinition.OperationId,
		ModuleName: operationDefinition.Spec.Tags[0],
	}
}

type ActionDefinition struct {
	ActionTemplateModel
	GeneratedCode *bytes.Buffer
}

func NewActionDefinition(appName string, operationDefinition codegen.OperationDefinition, generatedCode *bytes.Buffer) ActionDefinition {
	return ActionDefinition{
		ActionTemplateModel: NewActionTemplateModel(appName, operationDefinition),
		GeneratedCode:       generatedCode,
	}
}

func (g Generator) GenerateActionFiles() ([]ActionDefinition, error) {
	var actionDefinitions []ActionDefinition
	for _, operationDefinition := range g.OperationDefinitions {
		actionTemplateModel := NewActionTemplateModel(g.AppName, operationDefinition)
		out, err := g.ExecuteActionFileTemplate(actionTemplateModel)
		if err != nil {
			return nil, err
		}
		actionDefinitions = append(actionDefinitions, NewActionDefinition(g.AppName, operationDefinition, out))
	}

	return actionDefinitions, nil
}

func (g Generator) ExecuteActionFileTemplate(model ActionTemplateModel) (*bytes.Buffer, error) {
	return ExecuteTemplate(g.Templates, actionsTemplateFileName, model)
}

type ServiceTemplateModel struct {
	AppName     string
	ServiceName string
	ModuleName  string
}

type ServiceDefinition struct {
	ServiceTemplateModel
	GeneratedCode *bytes.Buffer
}

func NewServiceTemplateModel(appName string, operationDefinition codegen.OperationDefinition) ServiceTemplateModel {
	return ServiceTemplateModel{
		AppName:     appName,
		ServiceName: fmt.Sprintf("%sService", operationDefinition.OperationId),
		ModuleName:  operationDefinition.Spec.Tags[0],
	}
}

func NewServiceDefinition(serviceTemplateModel ServiceTemplateModel, generatedCode *bytes.Buffer) ServiceDefinition {
	return ServiceDefinition{
		ServiceTemplateModel: serviceTemplateModel,
		GeneratedCode:        generatedCode,
	}
}

func (g Generator) GenerateServiceFiles() ([]ServiceDefinition, error) {
	tmpl, err := template.New("hanami-service").Funcs(TemplateFunctions).ParseFS(templatesFS, "templates/service.rb.tmpl")
	if err != nil {
		return nil, fmt.Errorf("error parsing template files: %w", err)
	}

	var serviceDefinitions []ServiceDefinition
	for _, operationDefinition := range g.OperationDefinitions {
		serviceTemplateModel := NewServiceTemplateModel(g.AppName, operationDefinition)

		var buf bytes.Buffer
		w := bufio.NewWriter(&buf)
		if err = tmpl.ExecuteTemplate(w, "service.rb.tmpl", serviceTemplateModel); err != nil {
			return nil, fmt.Errorf("error executing service template: %w", err)
		}
		if err = w.Flush(); err != nil {
			return nil, fmt.Errorf("error flushing output buffer: %w", err)
		}

		serviceDefinitions = append(serviceDefinitions, NewServiceDefinition(serviceTemplateModel, &buf))
	}

	return serviceDefinitions, nil
}

func (g Generator) GenerateContractsFile() (*bytes.Buffer, error) {
	model, err := NewContractsFileTemplateModel(g.AppName, g.OperationDefinitions)
	if err != nil {
		return nil, fmt.Errorf("error generating contracts file template model: %w", err)
	}
	tmpl, err := template.New("hanami-contracts").Funcs(TemplateFunctions).ParseFS(templatesFS, "templates/contracts.rb.tmpl", "templates/fragments/attribute.rb.tmpl")
	if err != nil {
		return nil, fmt.Errorf("error parsing template files: %w", err)
	}

	var buf bytes.Buffer
	w := bufio.NewWriter(&buf)
	if err = tmpl.ExecuteTemplate(w, "contracts.rb.tmpl", model); err != nil {
		return nil, fmt.Errorf("error executing hanami-contracts template: %w", err)
	}
	if err = w.Flush(); err != nil {
		return nil, fmt.Errorf("error flushing output buffer: %w", err)
	}

	return &buf, nil
}

type ContractTemplateModel struct {
	ContractName string
	Attributes   []AttributeDefinition
}

type ContractsFileTemplateModel struct {
	AppName   string
	Contracts []ContractTemplateModel
}

func NewContractsFileTemplateModel(appName string, operationDefinitions []codegen.OperationDefinition) (*ContractsFileTemplateModel, error) {
	var contracts []ContractTemplateModel
	for _, operationDefinition := range operationDefinitions {
		requestContract := ContractTemplateModel{
			ContractName: fmt.Sprintf("%sRequestContract", operationDefinition.OperationId),
		}

		// injecting the request body attributes
		if operationDefinition.Spec.RequestBody != nil {
			requestContract.Attributes = GenerateAttributeDefinitions(operationDefinition.Spec.RequestBody.Value.Content["application/json"].Schema)
		}

		// injecting the query & path params
		for _, pathParam := range operationDefinition.Spec.Parameters {
			requestContract.Attributes = append(requestContract.Attributes, GenerateAttributeDefinition(pathParam.Value.Name, pathParam.Value.Schema, pathParam.Value.Required))
		}

		responseContract := ContractTemplateModel{
			ContractName: fmt.Sprintf("%sResponseContract", operationDefinition.OperationId),
			Attributes:   GenerateAttributeDefinitions(operationDefinition.Spec.Responses["200"].Value.Content["application/json"].Schema),
		}

		contracts = append(contracts, requestContract, responseContract)
	}

	return &ContractsFileTemplateModel{
		AppName:   appName,
		Contracts: contracts,
	}, nil
}

type SchemaTemplateModel struct {
	SchemaName string
	Attributes []AttributeDefinition
}

type SchemasFileTemplateModel struct {
	AppName string
	Schemas []SchemaTemplateModel
}

func NewSchemasFileTemplateModel(appName string, swagger *openapi3.T) (SchemasFileTemplateModel, error) {
	var schemas []SchemaTemplateModel

	for key, value := range swagger.Components.Schemas {
		schemaTemplateModel := SchemaTemplateModel{
			SchemaName: key,
			Attributes: GenerateAttributeDefinitions(value),
		}

		schemas = append(schemas, schemaTemplateModel)
	}

	return SchemasFileTemplateModel{AppName: appName, Schemas: schemas}, nil
}

func (g Generator) GenerateSchemasFile() (*bytes.Buffer, error) {
	model, err := NewSchemasFileTemplateModel(g.AppName, g.Swagger)
	if err != nil {
		return nil, fmt.Errorf("error generating schemas file template model: %w", err)
	}
	tmpl, err := template.New("hanami-schemas").Funcs(TemplateFunctions).ParseFS(templatesFS, "templates/schemas.rb.tmpl", "templates/fragments/attribute.rb.tmpl")
	if err != nil {
		return nil, fmt.Errorf("error parsing template files: %w", err)
	}

	var buf bytes.Buffer
	w := bufio.NewWriter(&buf)
	if err = tmpl.ExecuteTemplate(w, "schemas.rb.tmpl", model); err != nil {
		return nil, fmt.Errorf("error executing hanami-schemas template: %w", err)
	}
	if err = w.Flush(); err != nil {
		return nil, fmt.Errorf("error flushing output buffer: %w", err)
	}

	return &buf, nil
}

type AttributeDefinition struct {
	AttributeName    string
	AttributeType    string
	Verb             string
	HasChildren      bool
	NestedAttributes []AttributeDefinition
	Required         bool
}

func IsInArray(arr []string, val string) bool {
	for _, el := range arr {
		if el == val {
			return true
		}
	}

	return false
}

func GenerateAttributeDefinitions(schemaRef *openapi3.SchemaRef) []AttributeDefinition {
	if schemaRef == nil {
		return nil
	}
	var attributeDefinitions []AttributeDefinition
	for propertyKey, propertyValue := range schemaRef.Value.Properties {
		attributeDefinition := GenerateAttributeDefinition(propertyKey, propertyValue, IsInArray(schemaRef.Value.Required, propertyKey))
		attributeDefinitions = append(attributeDefinitions, attributeDefinition)
	}

	return attributeDefinitions
}

func GenerateAttributeDefinition(key string, schemaRef *openapi3.SchemaRef, required bool) AttributeDefinition {
	attributeDefinition := AttributeDefinition{
		AttributeName:    key,
		AttributeType:    "",
		Verb:             "",
		HasChildren:      false,
		NestedAttributes: nil,
		Required:         required,
	}

	if isRef(schemaRef) {
		attributeDefinition.AttributeType = GenerateReferencedSchemaType(schemaRef)
		return attributeDefinition
	}

	propertyType := schemaRef.Value.Type
	switch propertyType {
	case "string":
		attributeDefinition.AttributeType = ":string"
		attributeDefinition.Verb = "value"
	case "integer":
		attributeDefinition.AttributeType = ":integer"
		attributeDefinition.Verb = "value"
	case "array":
		attributeDefinition.Verb = "array"
		itemsAttributeDefinition := GenerateAttributeDefinition("", schemaRef.Value.Items, IsInArray(schemaRef.Value.Required, key)) // todo: don't hardcode
		attributeDefinition.AttributeType = itemsAttributeDefinition.AttributeType
		attributeDefinition.NestedAttributes = itemsAttributeDefinition.NestedAttributes
		attributeDefinition.HasChildren = len(itemsAttributeDefinition.NestedAttributes) > 0
	case "object":
		attributeDefinition.AttributeType = ":hash"
		attributeDefinition.Verb = "value"
		attributeDefinition.HasChildren = true
		attributeDefinition.NestedAttributes = GenerateAttributeDefinitions(schemaRef)
	}

	return attributeDefinition
}

func isRef(propertyValue *openapi3.SchemaRef) bool {
	return propertyValue.Ref != ""
}

func GenerateReferencedSchemaType(schemaRef *openapi3.SchemaRef) string {
	return fmt.Sprintf("Schemas::%s", schemaRef.Value.Title)
}
