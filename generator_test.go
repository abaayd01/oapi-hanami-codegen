package main

import (
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
			want: "/users/:user_idzx",
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
	b, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}

	return string(b), nil
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
