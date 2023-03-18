package main

import (
	"bufio"
	"bytes"
	"embed"
	"fmt"
	"github.com/deepmap/oapi-codegen/pkg/codegen"
	"github.com/deepmap/oapi-codegen/pkg/util"
	"github.com/getkin/kin-openapi/openapi3"
	"regexp"
	"text/template"
)

//go:embed templates/* templates/fragments/*
var templatesFS embed.FS
var templatesFilePath = "templates"
var routesTemplateFileName = "routes.rb.tmpl"
var actionTemplateFileName = "action.rb.tmpl"
var serviceTemplateFileName = "service.rb.tmpl"
var contractsTemplateFileName = "contracts.rb.tmpl"
var schemasTemplateFileName = "schemas.rb.tmpl"

type Generator struct {
	AppName              string
	OperationDefinitions []codegen.OperationDefinition
	Swagger              *openapi3.T
	Templates            *template.Template
}

func NewGenerator(inputFilePath string, appName string) (*Generator, error) {
	swagger, err := LoadSwagger(inputFilePath)
	if err != nil {
		return nil, err
	}

	operationDefinitions, err := codegen.OperationDefinitions(swagger)
	if err != nil {
		return nil, err
	}

	templates, err := LoadTemplates()
	if err != nil {
		return nil, err
	}

	return &Generator{
		AppName:              appName,
		OperationDefinitions: operationDefinitions,
		Swagger:              swagger,
		Templates:            templates,
	}, nil
}

func LoadSwagger(filePath string) (*openapi3.T, error) {
	swagger, err := util.LoadSwagger(filePath)
	if err != nil {
		return nil, fmt.Errorf("error loading swagger spec: %w", err)
	}

	// todo fix this, don't want to have to call this to get things to work
	_, err = codegen.Generate(swagger, codegen.Configuration{
		PackageName: "main",
		Generate:    codegen.GenerateOptions{},
	})
	if err != nil {
		return nil, fmt.Errorf("error executing codegen.Generate: %w", err)
	}

	return swagger, nil
}

var TemplateFunctions = merge(codegen.TemplateFunctions, template.FuncMap{
	"toSnake": toSnake,
})

func LoadTemplates() (*template.Template, error) {
	return template.New("templates").Funcs(TemplateFunctions).ParseFS(
		templatesFS,
		templatesFilePath+"/*.tmpl",
		templatesFilePath+"/fragments/*.tmpl",
	)
}

type RoutesFileTemplateModel struct {
	AppName string
	Routes  []RouteTemplateModel
}

type RouteTemplateModel struct {
	Method        string
	ModuleName    string
	OperationName string
	Path          string
}

// toRackPath converts a path definition as given by OpenAPI spec to something Rack understands.
// For example "/users/{user_id}" -> "/users/:user_id"
func toRackPath(codegenPath string) string {
	re := regexp.MustCompile("{(.*?)}")
	out := re.ReplaceAllString(codegenPath, ":$1")
	return out
}

func NewRoutesFileTemplateModel(appName string, operationDefinitions []codegen.OperationDefinition) (RoutesFileTemplateModel, error) {
	var routeTemplateModels []RouteTemplateModel
	for _, op := range operationDefinitions {
		routeTemplateModels = append(routeTemplateModels, RouteTemplateModel{
			Method:        op.Method,
			ModuleName:    op.Spec.Tags[0],
			OperationName: op.OperationId,
			Path:          toRackPath(op.Path),
		})
	}

	return RoutesFileTemplateModel{
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

func (g Generator) ExecuteRoutesFileTemplate(model RoutesFileTemplateModel) (*bytes.Buffer, error) {
	return executeTemplate(g.Templates, routesTemplateFileName, model)
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

func (g Generator) GenerateActionDefinitions() ([]ActionDefinition, error) {
	var actionDefinitions []ActionDefinition
	for _, operationDefinition := range g.OperationDefinitions {
		actionTemplateModel := NewActionTemplateModel(g.AppName, operationDefinition)
		actionFileBuf, err := g.ExecuteActionFileTemplate(actionTemplateModel)
		if err != nil {
			return nil, err
		}
		actionDefinitions = append(actionDefinitions, NewActionDefinition(g.AppName, operationDefinition, actionFileBuf))
	}

	return actionDefinitions, nil
}

func (g Generator) ExecuteActionFileTemplate(model ActionTemplateModel) (*bytes.Buffer, error) {
	return executeTemplate(g.Templates, actionTemplateFileName, model)
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

func (g Generator) GenerateServiceDefinitions() ([]ServiceDefinition, error) {
	var serviceDefinitions []ServiceDefinition
	for _, operationDefinition := range g.OperationDefinitions {
		serviceTemplateModel := NewServiceTemplateModel(g.AppName, operationDefinition)
		serviceFileBuf, err := g.ExecuteServiceFileTemplate(serviceTemplateModel)
		if err != nil {
			return nil, err
		}
		serviceDefinitions = append(serviceDefinitions, NewServiceDefinition(serviceTemplateModel, serviceFileBuf))
	}

	return serviceDefinitions, nil
}

func (g Generator) ExecuteServiceFileTemplate(model ServiceTemplateModel) (*bytes.Buffer, error) {
	return executeTemplate(g.Templates, serviceTemplateFileName, model)
}

func (g Generator) GenerateContractsFile() (*bytes.Buffer, error) {
	model, err := NewContractsFileTemplateModel(g.AppName, g.OperationDefinitions)
	if err != nil {
		return nil, fmt.Errorf("error generating contracts file template model: %w", err)
	}

	return g.ExecuteContractsFileTemplate(model)
}

func (g Generator) ExecuteContractsFileTemplate(model ContractsFileTemplateModel) (*bytes.Buffer, error) {
	return executeTemplate(g.Templates, contractsTemplateFileName, model)
}

type ContractTemplateModel struct {
	ContractName string
	Attributes   []AttributeDefinition
}

type ContractsFileTemplateModel struct {
	AppName   string
	Contracts []ContractTemplateModel
}

func NewContractsFileTemplateModel(appName string, operationDefinitions []codegen.OperationDefinition) (ContractsFileTemplateModel, error) {
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

	return ContractsFileTemplateModel{
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
	return g.ExecuteSchemasFileTemplate(model)
}

func (g Generator) ExecuteSchemasFileTemplate(model SchemasFileTemplateModel) (*bytes.Buffer, error) {
	return executeTemplate(g.Templates, schemasTemplateFileName, model)
}

func executeTemplate(tmpl *template.Template, filePath string, model any) (*bytes.Buffer, error) {
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

type AttributeDefinition struct {
	AttributeName    string
	AttributeType    string
	Verb             string
	HasChildren      bool
	NestedAttributes []AttributeDefinition
	Required         bool
}

func GenerateAttributeDefinitions(schemaRef *openapi3.SchemaRef) []AttributeDefinition {
	if schemaRef == nil {
		return nil
	}
	var attributeDefinitions []AttributeDefinition
	for propertyKey, propertyValue := range schemaRef.Value.Properties {
		attributeDefinition := GenerateAttributeDefinition(propertyKey, propertyValue, isInArray(schemaRef.Value.Required, propertyKey))
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
		itemsAttributeDefinition := GenerateAttributeDefinition("", schemaRef.Value.Items, isInArray(schemaRef.Value.Required, key))
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

func GenerateReferencedSchemaType(schemaRef *openapi3.SchemaRef) string {
	return fmt.Sprintf("Schemas::%s", schemaRef.Value.Title)
}

func isRef(propertyValue *openapi3.SchemaRef) bool {
	return propertyValue.Ref != ""
}

func isInArray(arr []string, val string) bool {
	for _, el := range arr {
		if el == val {
			return true
		}
	}

	return false
}
