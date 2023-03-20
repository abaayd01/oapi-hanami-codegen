package main

import (
	"github.com/stretchr/testify/assert"
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

func TestNewGenerator_MissingTags(t *testing.T) {
	_, err := NewGenerator("fixtures/test_spec_missing_tags.yaml", "TestApp")
	assert.ErrorContains(t, err, ErrMissingTags.Error())
}

func TestGenerator_GenerateRoutesFileTemplateModel(t *testing.T) {
	g, err := NewGenerator("fixtures/test_spec.yaml", "TestApp")
	if err != nil {
		t.Fatalf("error creating generator: %s\n", err)
	}

	model, err := g.GenerateRoutesFileTemplateModel()
	if err != nil {
		t.Fatalf("error generating routes template model: %s\n", err)
	}

	expected := RoutesFileTemplateModel{
		AppName: "TestApp",
		Routes: []RouteTemplateModel{
			{
				Method:        "GET",
				ModuleName:    "books",
				OperationName: "GetBooks",
				Path:          "/books",
			},
			{
				Method:        "GET",
				ModuleName:    "books",
				OperationName: "GetBookById",
				Path:          "/books/:bookId",
			},
		},
	}

	assert.Equal(t, expected, model)
}

func TestGenerator_GenerateActionTemplateModels(t *testing.T) {
	g, err := NewGenerator("fixtures/test_spec.yaml", "TestApp")
	if err != nil {
		t.Fatalf("error creating generator: %s\n", err)
	}

	actionTemplateModels, err := g.GenerateActionTemplateModels()
	if err != nil {
		t.Fatalf("error generating action template models: %s\n", err)
	}

	expectedActionTemplateModels := []ActionTemplateModel{
		{
			AppName:    "TestApp",
			ActionName: "GetBookById",
			ModuleName: "books",
		},
		{
			AppName:    "TestApp",
			ActionName: "GetBooks",
			ModuleName: "books",
		},
	}

	assert.ElementsMatch(t, expectedActionTemplateModels, actionTemplateModels)
}

func TestGenerator_GenerateServiceTemplateModels(t *testing.T) {
	g, err := NewGenerator("fixtures/test_spec.yaml", "TestApp")
	if err != nil {
		t.Fatalf("error creating generator: %s\n", err)
	}

	serviceTemplateModels, err := g.GenerateServiceTemplateModels()
	if err != nil {
		t.Fatalf("error generating service tempalte models: %s\n", err)
	}

	expectedServiceTemplateModels := []ServiceTemplateModel{
		{
			AppName:     "TestApp",
			ServiceName: "GetBookByIdService",
			ModuleName:  "books",
		},
		{
			AppName:     "TestApp",
			ServiceName: "GetBooksService",
			ModuleName:  "books",
		},
	}

	assert.ElementsMatch(t, expectedServiceTemplateModels, serviceTemplateModels)
}

func TestGenerator_GenerateContractsFileTemplateModel(t *testing.T) {
	g, err := NewGenerator("fixtures/test_spec.yaml", "TestApp")
	if err != nil {
		t.Fatalf("error creating generator: %s\n", err)
	}

	model, err := g.GenerateContractsFileTemplateModel()
	if err != nil {
		t.Fatalf("error generating contracts file: %s\n", err)
	}

	expected := ContractsFileTemplateModel{
		AppName: "TestApp",
		Contracts: []ContractTemplateModel{
			{
				ContractName: "GetBooksRequestContract",
				BaseClass:    "Hanami::Action::Params",
				Attributes:   nil,
			},
			{
				ContractName: "GetBooksResponseContract",
				BaseClass:    "Dry::Validation::Contract",
				Attributes: []AttributeDefinition{
					{
						AttributeName: "books",
						AttributeType: ":hash",
						Verb:          "array",
						HasChildren:   true,
						NestedAttributes: []AttributeDefinition{
							{
								AttributeName:    "author",
								AttributeType:    ":string",
								Verb:             "value",
								HasChildren:      false,
								NestedAttributes: nil,
								Required:         false,
							},
							{
								AttributeName:    "title",
								AttributeType:    ":string",
								Verb:             "value",
								HasChildren:      false,
								NestedAttributes: nil,
								Required:         false,
							},
						},
						Required: false,
					},
				},
			},
			{
				ContractName: "GetBookByIdRequestContract",
				BaseClass:    "Hanami::Action::Params",
				Attributes:   nil,
			},
			{
				ContractName: "GetBookByIdResponseContract",
				BaseClass:    "Dry::Validation::Contract",
				Attributes: []AttributeDefinition{
					{
						AttributeName:    "author",
						AttributeType:    ":string",
						Verb:             "value",
						HasChildren:      false,
						NestedAttributes: nil,
						Required:         true,
					},
					{
						AttributeName: "avatar",
						AttributeType: ":hash",
						Verb:          "value",
						HasChildren:   true,
						NestedAttributes: []AttributeDefinition{
							{
								AttributeName:    "id",
								AttributeType:    ":integer",
								Verb:             "value",
								HasChildren:      false,
								NestedAttributes: nil,
								Required:         false,
							},
							{
								AttributeName:    "profileImageUrl",
								AttributeType:    ":string",
								Verb:             "value",
								HasChildren:      false,
								NestedAttributes: nil,
								Required:         false,
							},
						},
						Required: false,
					},
					{
						AttributeName: "reviews",
						AttributeType: ":hash",
						Verb:          "array",
						HasChildren:   true,
						NestedAttributes: []AttributeDefinition{
							{
								AttributeName:    "rating",
								AttributeType:    ":integer",
								Verb:             "value",
								HasChildren:      false,
								NestedAttributes: nil,
								Required:         false,
							},
							{
								AttributeName:    "text",
								AttributeType:    ":string",
								Verb:             "value",
								HasChildren:      false,
								NestedAttributes: nil,
								Required:         true,
							},
							{
								AttributeName: "user",
								AttributeType: ":hash",
								Verb:          "value",
								HasChildren:   true,
								NestedAttributes: []AttributeDefinition{
									{
										AttributeName:    "id",
										AttributeType:    ":string",
										Verb:             "value",
										HasChildren:      false,
										NestedAttributes: nil,
										Required:         false,
									},
									{
										AttributeName:    "name",
										AttributeType:    ":string",
										Verb:             "value",
										HasChildren:      false,
										NestedAttributes: nil,
										Required:         false,
									},
								},
								Required: true,
							},
						},
						Required: true,
					},
					{
						AttributeName:    "title",
						AttributeType:    ":string",
						Verb:             "value",
						HasChildren:      false,
						NestedAttributes: nil,
						Required:         true,
					},
				},
			},
		},
	}

	assert.Equal(t, expected, model)
}

func TestGenerator_GenerateSchemasFileTemplateModel(t *testing.T) {
	g, err := NewGenerator("fixtures/test_spec.yaml", "TestApp")
	if err != nil {
		t.Fatalf("error creating generator: %s\n", err)
	}

	schemasFileTemplateModel, err := g.GenerateSchemasFileTemplateModel()
	if err != nil {
		t.Fatalf("error generating schemas file template model: %s\n", err)
	}

	expected := SchemasFileTemplateModel{
		AppName: "TestApp",
		Schemas: nil,
	}

	assert.Equal(t, expected, schemasFileTemplateModel)
}
