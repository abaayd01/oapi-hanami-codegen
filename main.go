package main

import (
	"bufio"
	"bytes"
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

func WriteFile(filePath string, data *bytes.Buffer) error {
	err := os.WriteFile(filePath, data.Bytes(), 0644)
	if err != nil {
		return fmt.Errorf("error writing to file %s: %w", filePath, err)
	}

	return nil
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

func (g Generator) GenerateRoutesFileTemplateModel(swagger *openapi3.T) (*RoutesFileTemplateModel, error) {
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
		AppName: g.AppName,
		Routes:  routeTemplateModels,
	}, nil
}

func (g Generator) GenerateRoutes(swagger *openapi3.T) (*bytes.Buffer, error) {
	routesFileTemplateModel, err := g.GenerateRoutesFileTemplateModel(swagger)
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
	return fmt.Sprintf("Types::%s", propertyValue.Value.Title)
}

func GenerateDryType(swagger *openapi3.T, propertyValue *openapi3.SchemaRef) string {
	if isRef(propertyValue) {
		return GenerateReferencedDryType(swagger, propertyValue)
	}

	propertyType := propertyValue.Value.Type
	var dryType string
	switch propertyType {
	case "string":
		dryType = "Types::String"
	case "integer":
		dryType = "Types::Integer"
	case "array":
		dryType = GenerateDryTypeForArray(swagger, propertyValue)
	case "object":
		dryType = GenerateDryTypeForObject(swagger, propertyValue)
	}

	return dryType
}

// GenerateAttributeDefinition ...
// probably going to be some recursion here eventually
func GenerateAttributeDefinition(swagger *openapi3.T, key string, propertyValue *openapi3.SchemaRef) AttributeDefinition {
	dryType := GenerateDryType(swagger, propertyValue)
	return AttributeDefinition{
		Name:    key,
		DryType: dryType,
	}
}
