package main

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"github.com/deepmap/oapi-codegen/pkg/codegen"
	"github.com/deepmap/oapi-codegen/pkg/util"
	"github.com/getkin/kin-openapi/openapi3"
	"log"
	"os"
	"regexp"
	"text/template"
)

func main() {
	var err error
	defer func() {
		if err != nil {
			log.Fatalln(err)
		}
	}()

	// todo, can pass in the file name as a command line arg
	swagger, err := LoadSwagger("./petstore_simple.yaml")
	if err != nil {
		return
	}

	// todo, pass in the AppName as a command line arg too
	g := Generator{
		AppName: "PetstoreApp",
	}

	routesBuf, err := g.GenerateRoutes(swagger)
	if err != nil {
		return
	}
	err = WriteRoutesFile(routesBuf)
	if err != nil {
		return
	}

	actions, err := g.GenerateActions(swagger)
	if err != nil {
		return
	}
	err = WriteActionFiles(actions)
	if err != nil {
		return
	}

	services, err := g.GenerateServices(swagger)
	if err != nil {
		return
	}
	err = WriteServiceFiles(services)

	contracts, err := g.GenerateContracts(swagger)
	if err != nil {
		return
	}
	err = WriteContractsFile(contracts)

	schemas, err := g.GenerateSchemas(swagger)
	if err != nil {
		return
	}
	err = WriteSchemasFile(schemas)
}

type Generator struct {
	// put extra config and stuff in here I guess
	AppName string
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

func NewRoutesFileTemplateModel(appName string, swagger *openapi3.T) (*RoutesFileTemplateModel, error) {
	ops, err := codegen.OperationDefinitions(swagger)
	if err != nil {
		return nil, fmt.Errorf("error generating operation definitions: %w", err)
	}

	var routeTemplateModels []RouteTemplateModel
	for _, op := range ops {
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

func (g Generator) GenerateRoutes(swagger *openapi3.T) (*bytes.Buffer, error) {
	routesFileTemplateModel, err := NewRoutesFileTemplateModel(g.AppName, swagger)
	if err != nil {
		return nil, fmt.Errorf("error generating routes file template model: %w", err)
	}

	tmpl, err := template.New("hanami-codegen").Funcs(TemplateFunctions).ParseFiles("./templates/hanami_routes.rb.tmpl")
	if err != nil {
		return nil, fmt.Errorf("error parsing template files: %w", err)
	}

	var buf bytes.Buffer
	w := bufio.NewWriter(&buf)
	if err = tmpl.ExecuteTemplate(w, "hanami_routes.rb.tmpl", routesFileTemplateModel); err != nil {
		return nil, fmt.Errorf("error executing hanami_routes template: %w", err)
	}
	if err = w.Flush(); err != nil {
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

func (g Generator) GenerateActions(swagger *openapi3.T) ([]ActionDefinition, error) {
	ops, err := codegen.OperationDefinitions(swagger)
	if err != nil {
		return nil, fmt.Errorf("error generating operation definitions: %w", err)
	}

	tmpl, err := template.New("hanami-action").Funcs(TemplateFunctions).ParseFiles("./templates/hanami_action.rb.tmpl")
	if err != nil {
		return nil, fmt.Errorf("error parsing template files: %w", err)
	}

	var actionDefinitions []ActionDefinition
	for _, operationDefinition := range ops {
		actionTemplateModel := NewActionTemplateModel(g.AppName, operationDefinition)

		var buf bytes.Buffer
		w := bufio.NewWriter(&buf)
		if err = tmpl.ExecuteTemplate(w, "hanami_action.rb.tmpl", actionTemplateModel); err != nil {
			return nil, fmt.Errorf("error executing hanami_action template: %w", err)
		}
		if err = w.Flush(); err != nil {
			return nil, fmt.Errorf("error flushing output buffer: %w", err)
		}

		actionDefinitions = append(actionDefinitions, NewActionDefinition(g.AppName, operationDefinition, &buf))
	}

	return actionDefinitions, nil
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

func (g Generator) GenerateServices(swagger *openapi3.T) ([]ServiceDefinition, error) {
	ops, err := codegen.OperationDefinitions(swagger)
	if err != nil {
		return nil, fmt.Errorf("error generating operation definitions: %w", err)
	}

	tmpl, err := template.New("hanami-service").Funcs(TemplateFunctions).ParseFiles("./templates/service.rb.tmpl")
	if err != nil {
		return nil, fmt.Errorf("error parsing template files: %w", err)
	}

	var serviceDefinitions []ServiceDefinition
	for _, operationDefinition := range ops {
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

type ContractAttributeDefinition struct {
	AttributeName string
	AttributeType string
}

type ContractTemplateModel struct {
	ContractName string
	Attributes   []ContractAttributeDefinition
}

type ContractsFileTemplateModel struct {
	AppName   string
	Contracts []ContractTemplateModel
}

func GenerateContractAttributeDefinitions(swagger *openapi3.T, schemaRef *openapi3.SchemaRef) []ContractAttributeDefinition {
	var attributeDefinitions []ContractAttributeDefinition
	for propertyKey, propertyValue := range schemaRef.Value.Properties {
		attributeDefinition := GenerateContractAttributeDefinition(swagger, propertyKey, propertyValue)
		attributeDefinitions = append(attributeDefinitions, attributeDefinition)
	}

	return attributeDefinitions
}

func GenerateContractAttributeDefinition(swagger *openapi3.T, key string, value *openapi3.SchemaRef) ContractAttributeDefinition {
	dryType := GenerateDryType(swagger, value)
	return ContractAttributeDefinition{
		AttributeName: key,
		AttributeType: dryType,
	}
}

func NewContractsFileTemplateModel(appName string, swagger *openapi3.T) (*ContractsFileTemplateModel, error) {
	ops, err := codegen.OperationDefinitions(swagger)
	if err != nil {
		return nil, fmt.Errorf("error generating operation definitions: %w", err)
	}

	var contracts []ContractTemplateModel
	for _, operationDefinition := range ops {
		requestContract := ContractTemplateModel{
			ContractName: fmt.Sprintf("%sRequestContract", operationDefinition.OperationId),
			Attributes:   nil,
		}

		responseContract := ContractTemplateModel{
			ContractName: fmt.Sprintf("%sResponseContract", operationDefinition.OperationId),
			Attributes:   GenerateContractAttributeDefinitions(swagger, operationDefinition.Spec.Responses["200"].Value.Content["application/json"].Schema),
		}

		contracts = append(contracts, requestContract, responseContract)
	}

	return &ContractsFileTemplateModel{
		AppName:   appName,
		Contracts: contracts,
	}, nil
}

func (g Generator) GenerateContracts(swagger *openapi3.T) (*bytes.Buffer, error) {
	model, err := NewContractsFileTemplateModel(g.AppName, swagger)
	if err != nil {
		return nil, fmt.Errorf("error generating contracts file template model: %w", err)
	}
	tmpl, err := template.New("hanami-contracts").Funcs(TemplateFunctions).ParseFiles("./templates/contracts.rb.tmpl")
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

// TODO unify schema attribute definition and contract attribute definition

type SchemaAttributeDefinition struct {
	AttributeName string
	AttributeType string
}

func GenerateSchemaAttributeDefinitions(swagger *openapi3.T, schemaRef *openapi3.SchemaRef) []SchemaAttributeDefinition {
	var attributeDefinitions []SchemaAttributeDefinition
	for propertyKey, propertyValue := range schemaRef.Value.Properties {
		attributeDefinition := GenerateSchemaAttributeDefinition(swagger, propertyKey, propertyValue)
		attributeDefinitions = append(attributeDefinitions, attributeDefinition)
	}

	return attributeDefinitions
}

func GenerateSchemaAttributeDefinition(swagger *openapi3.T, key string, value *openapi3.SchemaRef) SchemaAttributeDefinition {
	dryType := GenerateDryType(swagger, value)
	return SchemaAttributeDefinition{
		AttributeName: key,
		AttributeType: dryType,
	}
}

type SchemaTemplateModel struct {
	SchemaName string
	Attributes []SchemaAttributeDefinition
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
			Attributes: GenerateSchemaAttributeDefinitions(swagger, value),
		}

		schemas = append(schemas, schemaTemplateModel)
	}

	return SchemasFileTemplateModel{AppName: appName, Schemas: schemas}, nil
}

func (g Generator) GenerateSchemas(swagger *openapi3.T) (*bytes.Buffer, error) {
	model, err := NewSchemasFileTemplateModel(g.AppName, swagger)
	if err != nil {
		return nil, fmt.Errorf("error generating schemas file template model: %w", err)
	}
	tmpl, err := template.New("hanami-schemas").Funcs(TemplateFunctions).ParseFiles("./templates/schemas.rb.tmpl")
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
	Name    string
	DryType string
}

type ModelDefinition struct {
	ClassName  string
	Attributes []AttributeDefinition
}

func GenerateModelDefinitions(swagger *openapi3.T) ([]ModelDefinition, error) {
	var modelDefinitions []ModelDefinition

	for componentKey, component := range swagger.Components.Schemas {
		modelDefinition := GenerateModelDefinition(swagger, componentKey, component)
		modelDefinitions = append(modelDefinitions, modelDefinition)
	}

	for responseKey, response := range swagger.Components.Responses {
		modelDefinition := GenerateResponseModelDefinition(swagger, responseKey, response)
		modelDefinitions = append(modelDefinitions, modelDefinition)
	}

	return modelDefinitions, nil
}

func GenerateResponseModelDefinition(swagger *openapi3.T, key string, response *openapi3.ResponseRef) ModelDefinition {
	schemaRef := response.Value.Content.Get("application/json").Schema
	return GenerateModelDefinition(swagger, key, schemaRef)
}

func GenerateModelDefinition(swagger *openapi3.T, key string, schemaRef *openapi3.SchemaRef) ModelDefinition {
	// assume the root is an object?

	attributes := GenerateAttributeDefinitions(swagger, schemaRef)

	return ModelDefinition{
		ClassName:  key,
		Attributes: attributes,
	}
}

func GenerateAttributeDefinitions(swagger *openapi3.T, schemaRef *openapi3.SchemaRef) []AttributeDefinition {
	var attributeDefinitions []AttributeDefinition
	for propertyKey, propertyValue := range schemaRef.Value.Properties {
		attributeDefinition := GenerateAttributeDefinition(swagger, propertyKey, propertyValue)
		attributeDefinitions = append(attributeDefinitions, attributeDefinition)
	}

	return attributeDefinitions
}

func GenerateDryTypeForArray(swagger *openapi3.T, propertyValue *openapi3.SchemaRef) string {
	innerType := GenerateDryType(swagger, propertyValue.Value.Items)
	return fmt.Sprintf("Types::Array.of(%s)", innerType)
}

func GenerateDryTypeForObject(swagger *openapi3.T, propertyValue *openapi3.SchemaRef) string {
	attributeDefinitions := GenerateAttributeDefinitions(swagger, propertyValue)
	tmpl, _ := template.New("dry-hash").Parse("{{range .}}{{.Name}}: {{.DryType}}, {{end}}")
	var buf bytes.Buffer
	w := bufio.NewWriter(&buf)
	err := tmpl.ExecuteTemplate(w, "dry-hash", attributeDefinitions)
	if err != nil {
		log.Println("could not execute template dry-hash")
	}

	if err = w.Flush(); err != nil {
		// todo, fix the error handling
		fmt.Errorf("error flushing output buffer")
	}

	innerTypes := buf.String()

	return fmt.Sprintf("Types::Hash.schema(%s)", innerTypes)
}

func isRef(propertyValue *openapi3.SchemaRef) bool {
	return propertyValue.Ref != ""
}

func GenerateReferencedDryType(swagger *openapi3.T, propertyValue *openapi3.SchemaRef) string {
	return fmt.Sprintf("Schemas::%s", propertyValue.Value.Title)
}

func GenerateDryType(swagger *openapi3.T, propertyValue *openapi3.SchemaRef) string {
	if isRef(propertyValue) {
		return GenerateReferencedDryType(swagger, propertyValue)
	}

	propertyType := propertyValue.Value.Type
	var dryType string
	switch propertyType {
	case "string":
		dryType = ":string"
	case "integer":
		dryType = ":integer"
	case "array":
		dryType = GenerateDryTypeForArray(swagger, propertyValue)
	case "object":
		dryType = GenerateDryTypeForObject(swagger, propertyValue)
	}

	return dryType
}

func GenerateAttributeDefinition(swagger *openapi3.T, key string, propertyValue *openapi3.SchemaRef) AttributeDefinition {
	dryType := GenerateDryType(swagger, propertyValue)
	return AttributeDefinition{
		Name:    key,
		DryType: dryType,
	}
}
