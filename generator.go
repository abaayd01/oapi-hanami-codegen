package main

import (
	"errors"
	"fmt"
	"github.com/deepmap/oapi-codegen/pkg/codegen"
	"github.com/deepmap/oapi-codegen/pkg/util"
	"github.com/getkin/kin-openapi/openapi3"
	"regexp"
	"sort"
)

type Generator struct {
	AppName              string
	SliceName            string // only support a single slice for now
	OperationDefinitions []OperationDefinition
	Swagger              *openapi3.T
}

func NewGenerator(inputFilePath string, appName string, sliceName string) (*Generator, error) {
	swagger, err := loadSwagger(inputFilePath)
	if err != nil {
		return nil, err
	}

	codegenOperationDefinitions, err := codegen.OperationDefinitions(swagger)
	if err != nil {
		return nil, err
	}

	var operationDefinitions []OperationDefinition
	for i, _ := range codegenOperationDefinitions {
		operationDefinition, err := NewOperationDefinition(codegenOperationDefinitions[i])
		if err != nil {
			return nil, err
		}
		operationDefinitions = append(operationDefinitions, *operationDefinition)
	}

	return &Generator{
		AppName:              appName,
		SliceName:            sliceName,
		OperationDefinitions: operationDefinitions,
		Swagger:              swagger,
	}, nil
}

func loadSwagger(filePath string) (*openapi3.T, error) {
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

type OperationDefinition struct {
	*codegen.OperationDefinition
	ModuleName            string
	RequestBodySchema     *openapi3.SchemaRef
	ResponseBody200Schema *openapi3.SchemaRef
}

func NewOperationDefinition(codegenOperationDefinition codegen.OperationDefinition) (*OperationDefinition, error) {
	if codegenOperationDefinition.Spec == nil {
		return nil, ErrSpecCannotBeNil
	}

	moduleName, err := safelyDigModuleName(codegenOperationDefinition)
	if err != nil {
		return nil, fmt.Errorf("error digging out module name from tags: %w", err)
	}

	requestBodySchema, err := safelyDigRequestBodySchema(codegenOperationDefinition)
	if err != nil {
		return nil, fmt.Errorf("error digging out request body schema: %w", err)
	}

	responseBody200Schema, err := safelyDigResponseBody200Schema(codegenOperationDefinition)
	if err != nil {
		return nil, fmt.Errorf("error digging out response body 200 schema: %w", err)
	}

	return &OperationDefinition{
		OperationDefinition:   &codegenOperationDefinition,
		ModuleName:            moduleName,
		RequestBodySchema:     requestBodySchema,
		ResponseBody200Schema: responseBody200Schema,
	}, nil
}

var MediaTypeJson = "application/json"

var ErrMissingTags = errors.New("operation definition must specify at least one tag")
var ErrSpecCannotBeNil = errors.New("operation definition Spec attribute cannot be nil")
var ErrMalformedSpec = errors.New("operation definition Spec attribute is malformed")
var ErrMalformedSpecNoRequestBodyJsonMediaType = errors.New("operation definition Spec RequestBody must define an application/json media type")
var Err200ResponseBodyMissing = errors.New("operation definition must define a 200 response body")
var Err200ResponseBodyNoJsonMediaType = errors.New("operation definition 200 response body is missing application/json response")

func safelyDigModuleName(codegenOperationDefinition codegen.OperationDefinition) (string, error) {
	tags := codegenOperationDefinition.Spec.Tags

	if len(tags) == 0 {
		return "", ErrMissingTags
	}
	return tags[0], nil
}

func safelyDigRequestBodySchema(codegenOperationDefinition codegen.OperationDefinition) (*openapi3.SchemaRef, error) {
	var requestBodySchema *openapi3.SchemaRef
	if codegenOperationDefinition.Spec.RequestBody != nil {
		if codegenOperationDefinition.Spec.RequestBody.Value == nil {
			return nil, ErrMalformedSpec
		}

		if codegenOperationDefinition.Spec.RequestBody.Value.GetMediaType(MediaTypeJson) == nil {
			return nil, ErrMalformedSpecNoRequestBodyJsonMediaType
		}

		if codegenOperationDefinition.Spec.RequestBody.Value.GetMediaType(MediaTypeJson).Schema == nil {
			return nil, ErrMalformedSpec
		}

		requestBodySchema = codegenOperationDefinition.Spec.RequestBody.Value.GetMediaType(MediaTypeJson).Schema
	}

	return requestBodySchema, nil
}

func safelyDigResponseBody200Schema(codegenOperationDefinition codegen.OperationDefinition) (*openapi3.SchemaRef, error) {
	response := codegenOperationDefinition.Spec.Responses.Get(200)

	if response == nil {
		response = codegenOperationDefinition.Spec.Responses.Get(201)
	}

	if response == nil {
		return nil, Err200ResponseBodyMissing
	}

	if response.Value == nil {
		return nil, ErrMalformedSpec
	}

	if response.Value.Content.Get(MediaTypeJson) == nil {
		return nil, Err200ResponseBodyNoJsonMediaType
	}

	return response.Value.Content.Get(MediaTypeJson).Schema, nil
}

type TemplateModels struct {
	RoutesFileTemplateModel    RoutesFileTemplateModel
	ActionTemplateModels       []ActionTemplateModel
	ServiceTemplateModels      []ServiceTemplateModel
	ContractsFileTemplateModel ContractsFileTemplateModel
	SchemasFileTemplateModel   SchemasFileTemplateModel
}

func (g Generator) GenerateTemplateModels() (*TemplateModels, error) {
	routesFileTemplateModel, err := g.GenerateRoutesFileTemplateModel()
	if err != nil {
		return nil, fmt.Errorf("failed to generate routes file template model: %w\n", err)
	}

	actionTemplateModels, err := g.GenerateActionTemplateModels()
	if err != nil {
		return nil, fmt.Errorf("failed to generate action template models: %w\n", err)
	}

	serviceTemplateModels, err := g.GenerateServiceTemplateModels()
	if err != nil {
		return nil, fmt.Errorf("failed to generate service template models: %w\n", err)
	}

	contractsFileTemplateModel, err := g.GenerateContractsFileTemplateModel()
	if err != nil {
		return nil, fmt.Errorf("failed to generate contracts file template models: %w\n", err)
	}

	schemasFileTemplateModel, err := g.GenerateSchemasFileTemplateModel()
	if err != nil {
		return nil, fmt.Errorf("failed to generate schemas file template models: %w\n", err)
	}

	return &TemplateModels{
		RoutesFileTemplateModel:    routesFileTemplateModel,
		ActionTemplateModels:       actionTemplateModels,
		ServiceTemplateModels:      serviceTemplateModels,
		ContractsFileTemplateModel: contractsFileTemplateModel,
		SchemasFileTemplateModel:   schemasFileTemplateModel,
	}, nil
}

type RoutesFileTemplateModel struct {
	AppName   string
	SliceName string
	Routes    []RouteTemplateModel
}

type RouteTemplateModel struct {
	Method        string
	ModuleName    string
	OperationName string
	Path          string
}

func (g Generator) GenerateRoutesFileTemplateModel() (RoutesFileTemplateModel, error) {
	var routeTemplateModels []RouteTemplateModel
	for _, operationDefinition := range g.OperationDefinitions {
		routeTemplateModels = append(routeTemplateModels, RouteTemplateModel{
			Method:        operationDefinition.Method,
			ModuleName:    operationDefinition.ModuleName,
			OperationName: operationDefinition.OperationId,
			Path:          toRackPath(operationDefinition.Path),
		})
	}

	return RoutesFileTemplateModel{
		AppName:   g.AppName,
		SliceName: g.SliceName,
		Routes:    routeTemplateModels,
	}, nil
}

// toRackPath converts a path definition as given by OpenAPI spec to something Rack understands.
// For example "/users/{user_id}" -> "/users/:user_id"
func toRackPath(codegenPath string) string {
	re := regexp.MustCompile("{(.*?)}")
	out := re.ReplaceAllString(codegenPath, ":$1")
	return out
}

type ActionTemplateModel struct {
	AppName    string
	SliceName  string
	ActionName string
	ModuleName string
}

func NewActionTemplateModel(appName string, sliceName string, operationDefinition OperationDefinition) ActionTemplateModel {
	return ActionTemplateModel{
		AppName:    appName,
		SliceName:  sliceName,
		ActionName: operationDefinition.OperationId,
		ModuleName: operationDefinition.ModuleName,
	}
}

func (g Generator) GenerateActionTemplateModels() ([]ActionTemplateModel, error) {
	var actionTemplateModels []ActionTemplateModel
	for _, operationDefinition := range g.OperationDefinitions {
		actionTemplateModels = append(actionTemplateModels, NewActionTemplateModel(g.AppName, g.SliceName, operationDefinition))
	}

	return actionTemplateModels, nil
}

type ServiceTemplateModel struct {
	AppName     string
	SliceName   string
	ServiceName string
	ModuleName  string
}

func NewServiceTemplateModel(appName string, sliceName string, operationDefinition OperationDefinition) ServiceTemplateModel {
	return ServiceTemplateModel{
		AppName:     appName,
		SliceName:   sliceName,
		ServiceName: fmt.Sprintf("%s", operationDefinition.OperationId),
		ModuleName:  operationDefinition.ModuleName,
	}
}

func (g Generator) GenerateServiceTemplateModels() ([]ServiceTemplateModel, error) {
	var serviceTemplateModels []ServiceTemplateModel
	for _, operationDefinition := range g.OperationDefinitions {
		serviceTemplateModels = append(serviceTemplateModels, NewServiceTemplateModel(g.AppName, g.SliceName, operationDefinition))
	}

	return serviceTemplateModels, nil
}

type ContractTemplateModel struct {
	ContractName string
	BaseClass    string
	Attributes   []AttributeDefinition
}

type ContractsFileTemplateModel struct {
	AppName   string
	SliceName string
	Contracts []ContractTemplateModel
}

func (g Generator) GenerateContractsFileTemplateModel() (ContractsFileTemplateModel, error) {
	var contracts []ContractTemplateModel
	for _, operationDefinition := range g.OperationDefinitions {
		requestContract := ContractTemplateModel{
			ContractName: fmt.Sprintf("%sRequestContract", operationDefinition.OperationId),
			BaseClass:    "Hanami::Action::Params",
		}

		// injecting the request body attributes
		if operationDefinition.Spec.RequestBody != nil {
			requestContract.Attributes = generateAttributeDefinitions(operationDefinition.RequestBodySchema)
		}

		// injecting the query & path params
		for _, pathParam := range operationDefinition.Spec.Parameters {
			requestContract.Attributes = append(requestContract.Attributes, generateAttributeDefinition(pathParam.Value.Name, pathParam.Value.Schema, pathParam.Value.Required))
		}

		responseContract := ContractTemplateModel{
			ContractName: fmt.Sprintf("%sResponseContract", operationDefinition.OperationId),
			BaseClass:    "Dry::Validation::Contract",
			Attributes:   generateAttributeDefinitions(operationDefinition.ResponseBody200Schema),
		}

		contracts = append(contracts, requestContract, responseContract)
	}

	return ContractsFileTemplateModel{
		AppName:   g.AppName,
		SliceName: g.SliceName,
		Contracts: contracts,
	}, nil
}

type SchemaTemplateModel struct {
	SchemaName string
	Attributes []AttributeDefinition
}

type SchemasFileTemplateModel struct {
	AppName   string
	SliceName string
	Schemas   []SchemaTemplateModel
}

func (g Generator) GenerateSchemasFileTemplateModel() (SchemasFileTemplateModel, error) {
	var schemas []SchemaTemplateModel

	for key, value := range g.Swagger.Components.Schemas {
		schemaTemplateModel := SchemaTemplateModel{
			SchemaName: key,
			Attributes: generateAttributeDefinitions(value),
		}

		schemas = append(schemas, schemaTemplateModel)
	}

	return SchemasFileTemplateModel{
		AppName:   g.AppName,
		SliceName: g.SliceName,
		Schemas:   schemas,
	}, nil
}

type AttributeDefinition struct {
	AttributeName    string
	AttributeType    string
	Verb             string
	HasChildren      bool
	NestedAttributes []AttributeDefinition
	Required         bool
}

func generateAttributeDefinitions(schemaRef *openapi3.SchemaRef) []AttributeDefinition {
	if schemaRef == nil {
		return nil
	}

	var attributeDefinitions []AttributeDefinition

	// Want to sort the keys to make sure they come out in a consistent order, because
	// I don't think openapi3 makes guarantees about the order they are in openapi3.Schemas map,
	// which causes flaky tests.
	//
	// Sorts alphabetically.
	sortedKeys := sortedSchemaRefPropertyKeys(schemaRef)

	for _, propertyKey := range sortedKeys {
		propertyValue := schemaRef.Value.Properties[propertyKey]
		attributeDefinition := generateAttributeDefinition(propertyKey, propertyValue, isInArray(schemaRef.Value.Required, propertyKey))
		attributeDefinitions = append(attributeDefinitions, attributeDefinition)
	}

	return attributeDefinitions
}

func sortedSchemaRefPropertyKeys(schemaRef *openapi3.SchemaRef) []string {
	properties := schemaRef.Value.Properties
	sortedKeys := make([]string, 0)
	for k, _ := range properties {
		sortedKeys = append(sortedKeys, k)
	}
	sort.Strings(sortedKeys)
	return sortedKeys
}

func generateAttributeDefinition(key string, schemaRef *openapi3.SchemaRef, required bool) AttributeDefinition {
	attributeDefinition := AttributeDefinition{
		AttributeName:    key,
		AttributeType:    "",
		Verb:             "",
		HasChildren:      false,
		NestedAttributes: nil,
		Required:         required,
	}

	if isRef(schemaRef) {
		attributeDefinition.AttributeType = referencedSchemaType(schemaRef)
		return attributeDefinition
	}

	propertyType := schemaRef.Value.Type
	switch propertyType {
	case "string":
		attributeDefinition.Verb = "value"

		if schemaRef.Value.Format == "uuid" {
			attributeDefinition.AttributeType = ":uuid_v4?"
		} else {
			attributeDefinition.AttributeType = ":string"
		}
	case "integer":
		attributeDefinition.AttributeType = ":integer"
		attributeDefinition.Verb = "value"
	case "array":
		attributeDefinition.Verb = "array"
		itemsAttributeDefinition := generateAttributeDefinition("", schemaRef.Value.Items, isInArray(schemaRef.Value.Required, key))
		attributeDefinition.AttributeType = itemsAttributeDefinition.AttributeType
		attributeDefinition.NestedAttributes = itemsAttributeDefinition.NestedAttributes
		attributeDefinition.HasChildren = len(itemsAttributeDefinition.NestedAttributes) > 0
	case "object":
		attributeDefinition.AttributeType = ":hash"
		attributeDefinition.Verb = "value"
		attributeDefinition.HasChildren = true
		attributeDefinition.NestedAttributes = generateAttributeDefinitions(schemaRef)
	}

	return attributeDefinition
}

func isRef(propertyValue *openapi3.SchemaRef) bool {
	return propertyValue.Ref != ""
}

func referencedSchemaType(schemaRef *openapi3.SchemaRef) string {
	return fmt.Sprintf("Schemas::%s", schemaRef.Value.Title)
}

func isInArray(arr []string, val string) bool {
	for _, el := range arr {
		if el == val {
			return true
		}
	}

	return false
}
