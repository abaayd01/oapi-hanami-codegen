package main

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func Test_toRackPath(t *testing.T) {
	type args struct {
		codegenPath string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "it converts a path with a single param correctly.",
			args: args{
				codegenPath: "/users/{user_id}",
			},
			want: "/users/:user_id",
		},
		{
			name: "doesn't do anything if there's no path params",
			args: args{
				codegenPath: "/users",
			},
			want: "/users",
		},
		{
			name: "still works if there's multiple path params",
			args: args{
				codegenPath: "/goals/{goal_id}/key_result/{key_result_id}",
			},
			want: "/goals/:goal_id/key_result/:key_result_id",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := toRackPath(tt.args.codegenPath)
			assert.Equal(t, tt.want, result)
		})
	}
}

func readFixture(filePath string) (string, error) {
	b, err := os.ReadFile("fixtures/" + filePath)
	if err != nil {
		return "", err
	}

	return string(b), nil
}

func TestNewGenerator_MissingTags(t *testing.T) {
	_, err := NewGenerator("fixtures/test_spec_missing_tags.yaml", "TestApp")
	assert.EqualError(t, err, ErrMissingTags.Error())
}

func TestGenerator_GenerateRoutesFile(t *testing.T) {
	g, err := NewGenerator("fixtures/test_spec.yaml", "TestApp")
	if err != nil {
		t.Fatalf("error creating generator: %s\n", err)
	}

	routesFileBuf, err := g.GenerateRoutesFile()
	if err != nil {
		t.Fatalf("error generating routes file: %s\n", err)
	}

	expectedFileBuf, err := readFixture("out/config/routes.rb")
	if err != nil {
		t.Fatalf("error reading fixture out/config/routes.rb: %s\n", err)
	}

	assert.Equal(t, expectedFileBuf, routesFileBuf.String())
}

func TestGenerator_GenerateActionDefinitions(t *testing.T) {
	g, err := NewGenerator("fixtures/test_spec.yaml", "TestApp")
	if err != nil {
		t.Fatalf("error creating generator: %s\n", err)
	}

	actionDefinitions, err := g.GenerateActionDefinitions()
	if err != nil {
		t.Fatalf("error generating action definitions: %s\n", err)
	}

	getBookByIdActionFixture, err := readFixture("out/actions/books/get_book_by_id.rb")
	if err != nil {
		t.Fatalf("error reading fixture out/actions/books/get_book_by_id.rb: %s\n", err)
	}

	getBooksActionFixture, err := readFixture("out/actions/books/get_books.rb")
	if err != nil {
		t.Fatalf("error reading fixture out/actions/books/get_books.rb: %s\n", err)
	}

	expectedActionDefinitions := map[string]ActionDefinition{
		"GetBookById": {
			ActionTemplateModel: ActionTemplateModel{
				AppName:    "TestApp",
				ActionName: "GetBookById",
				ModuleName: "books",
			},
			GeneratedCode: bytes.NewBufferString(getBookByIdActionFixture),
		},
		"GetBooks": {
			ActionTemplateModel: ActionTemplateModel{
				AppName:    "TestApp",
				ActionName: "GetBooks",
				ModuleName: "books",
			},
			GeneratedCode: bytes.NewBufferString(getBooksActionFixture),
		},
	}

	for _, actionDefinition := range actionDefinitions {
		key := actionDefinition.ActionName
		expectedActionDefinition, ok := expectedActionDefinitions[key]
		if !ok {
			t.Fatalf("expectedActionDefinition with key '%s' not found", key)
		}

		assert.Equal(t, expectedActionDefinition.GeneratedCode.String(), actionDefinition.GeneratedCode.String())
		assert.Equal(t, expectedActionDefinition, actionDefinition)
	}
}

func TestGenerator_GenerateServiceDefinitions(t *testing.T) {
	g, err := NewGenerator("fixtures/test_spec.yaml", "TestApp")
	if err != nil {
		t.Fatalf("error creating generator: %s\n", err)
	}

	serviceDefinitions, err := g.GenerateServiceDefinitions()
	if err != nil {
		t.Fatalf("error generating service definitions: %s\n", err)
	}

	getBookByIdServiceFixture, err := readFixture("out/actions/books/get_book_by_id_service.rb")
	if err != nil {
		t.Fatalf("error reading fixture out/actions/books/get_book_by_id_service.rb: %s\n", err)
	}

	getBooksServiceFixture, err := readFixture("out/actions/books/get_books_service.rb")
	if err != nil {
		t.Fatalf("error reading fixture out/actions/books/get_books_service.rb: %s\n", err)
	}

	expectedServiceDefinitions := map[string]ServiceDefinition{
		"GetBookByIdService": {
			ServiceTemplateModel: ServiceTemplateModel{
				AppName:     "TestApp",
				ServiceName: "GetBookByIdService",
				ModuleName:  "books",
			},
			GeneratedCode: bytes.NewBufferString(getBookByIdServiceFixture),
		},
		"GetBooksService": {
			ServiceTemplateModel: ServiceTemplateModel{
				AppName:     "TestApp",
				ServiceName: "GetBooksService",
				ModuleName:  "books",
			},
			GeneratedCode: bytes.NewBufferString(getBooksServiceFixture),
		},
	}

	for _, serviceDefinition := range serviceDefinitions {
		key := serviceDefinition.ServiceName
		expectedServiceDefinition, ok := expectedServiceDefinitions[key]
		if !ok {
			t.Fatalf("expectedServiceDefinition with key '%s' not found", key)
		}

		assert.Equal(t, expectedServiceDefinition.GeneratedCode.String(), serviceDefinition.GeneratedCode.String())
		assert.Equal(t, expectedServiceDefinition, serviceDefinition)
	}
}

// TODO: this test seems to be flaky since ordering of attributes in generated file isn't stable
// maybe test the underlying functions separately instead?
func TestGenerator_GenerateContractsFile(t *testing.T) {
	g, err := NewGenerator("fixtures/test_spec.yaml", "TestApp")
	if err != nil {
		t.Fatalf("error creating generator: %s\n", err)
	}

	contractsFileBuf, err := g.GenerateContractsFile()
	if err != nil {
		t.Fatalf("error generating contracts file: %s\n", err)
	}

	expectedFileBuf, err := readFixture("out/actions/contracts.rb")
	if err != nil {
		t.Fatalf("error reading fixture out/actions/contracts.rb: %s\n", err)
	}

	assert.Equal(t, expectedFileBuf, contractsFileBuf.String())
}

func TestGenerator_GenerateSchemas(t *testing.T) {
	g, err := NewGenerator("fixtures/test_spec.yaml", "TestApp")
	if err != nil {
		t.Fatalf("error creating generator: %s\n", err)
	}

	schemasFileBuf, err := g.GenerateSchemasFile()
	if err != nil {
		t.Fatalf("error generating schemas file: %s\n", err)
	}

	expectedFileBuf, err := readFixture("out/actions/schemas.rb")
	if err != nil {
		t.Fatalf("error reading fixture out/actions/schemas.rb: %s\n", err)
	}

	assert.Equal(t, expectedFileBuf, schemasFileBuf.String())
}
