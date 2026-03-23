package ast

import (
	"go/ast"
	"go/token"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetModulePath(T *testing.T) {
	T.Parallel()

	T.Run("reads module path from go.mod", func(t *testing.T) {
		t.Parallel()

		dir := t.TempDir()
		err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module github.com/example/test\n\ngo 1.21\n"), 0o600)
		require.NoError(t, err)

		path, err := GetModulePath(dir)

		require.NoError(t, err)
		assert.Equal(t, "github.com/example/test", path)
	})

	T.Run("returns error when go.mod does not exist", func(t *testing.T) {
		t.Parallel()

		path, err := GetModulePath(t.TempDir())

		assert.Empty(t, path)
		assert.Error(t, err)
	})

	T.Run("returns error when no module directive found", func(t *testing.T) {
		t.Parallel()

		dir := t.TempDir()
		err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("go 1.21\n"), 0o600)
		require.NoError(t, err)

		path, err := GetModulePath(dir)

		assert.Empty(t, path)
		assert.Error(t, err)
	})
}

func TestBuildImportMap(T *testing.T) {
	T.Parallel()

	T.Run("builds map from imports", func(t *testing.T) {
		t.Parallel()

		file := &ast.File{
			Imports: []*ast.ImportSpec{
				{Path: &ast.BasicLit{Value: `"fmt"`}},
				{Path: &ast.BasicLit{Value: `"github.com/example/pkg"`}},
			},
		}

		result := BuildImportMap(file)

		assert.Equal(t, "fmt", result["fmt"])
		assert.Equal(t, "github.com/example/pkg", result["pkg"])
	})

	T.Run("handles aliased imports", func(t *testing.T) {
		t.Parallel()

		file := &ast.File{
			Imports: []*ast.ImportSpec{
				{
					Name: &ast.Ident{Name: "myfmt"},
					Path: &ast.BasicLit{Value: `"fmt"`},
				},
			},
		}

		result := BuildImportMap(file)

		assert.Equal(t, "fmt", result["myfmt"])
	})

	T.Run("excludes blank and dot imports", func(t *testing.T) {
		t.Parallel()

		file := &ast.File{
			Imports: []*ast.ImportSpec{
				{
					Name: &ast.Ident{Name: "_"},
					Path: &ast.BasicLit{Value: `"image/png"`},
				},
				{
					Name: &ast.Ident{Name: "."},
					Path: &ast.BasicLit{Value: `"testing"`},
				},
			},
		}

		result := BuildImportMap(file)

		assert.Empty(t, result)
	})

	T.Run("skips imports with nil path", func(t *testing.T) {
		t.Parallel()

		file := &ast.File{
			Imports: []*ast.ImportSpec{
				{Path: nil},
			},
		}

		result := BuildImportMap(file)

		assert.Empty(t, result)
	})
}

func TestFilterModuleImports(T *testing.T) {
	T.Parallel()

	T.Run("filters to module-internal imports", func(t *testing.T) {
		t.Parallel()

		imports := map[string]string{
			"fmt":     "fmt",
			"logging": "github.com/example/mod/observability/logging",
			"errors":  "github.com/example/mod/errors",
		}

		result := FilterModuleImports(imports, "github.com/example/mod")

		assert.Len(t, result, 2)
		assert.Equal(t, "observability/logging", result["logging"])
		assert.Equal(t, "errors", result["errors"])
	})

	T.Run("returns empty map when no module imports", func(t *testing.T) {
		t.Parallel()

		imports := map[string]string{
			"fmt": "fmt",
		}

		result := FilterModuleImports(imports, "github.com/example/mod")

		assert.Empty(t, result)
	})
}

func TestGetTagValue(T *testing.T) {
	T.Parallel()

	T.Run("extracts tag value", func(t *testing.T) {
		t.Parallel()

		assert.Equal(t, "name", GetTagValue(`json:"name"`, "json"))
	})

	T.Run("extracts tag value with omitempty", func(t *testing.T) {
		t.Parallel()

		assert.Equal(t, "name", GetTagValue(`json:"name,omitempty"`, "json"))
	})

	T.Run("extracts from multiple tags", func(t *testing.T) {
		t.Parallel()

		tag := `json:"name" env:"MY_VAR"`
		assert.Equal(t, "name", GetTagValue(tag, "json"))
		assert.Equal(t, "MY_VAR", GetTagValue(tag, "env"))
	})

	T.Run("returns empty for missing key", func(t *testing.T) {
		t.Parallel()

		assert.Equal(t, "", GetTagValue(`json:"name"`, "xml"))
	})

	T.Run("handles backtick-wrapped tags", func(t *testing.T) {
		t.Parallel()

		assert.Equal(t, "name", GetTagValue("`json:\"name\"`", "json"))
	})
}

func TestGetStructFields(T *testing.T) {
	T.Parallel()

	T.Run("returns field names and types", func(t *testing.T) {
		t.Parallel()

		st := &ast.StructType{
			Fields: &ast.FieldList{
				List: []*ast.Field{
					{
						Names: []*ast.Ident{{Name: "Name"}},
						Type:  &ast.Ident{Name: "string"},
					},
					{
						Names: []*ast.Ident{{Name: "Logger"}},
						Type: &ast.SelectorExpr{
							X:   &ast.Ident{Name: "logging"},
							Sel: &ast.Ident{Name: "Logger"},
						},
					},
				},
			},
		}

		fields := GetStructFields(st)

		assert.Equal(t, "string", fields["Name"])
		assert.Equal(t, "logging.Logger", fields["Logger"])
	})

	T.Run("excludes underscore fields", func(t *testing.T) {
		t.Parallel()

		st := &ast.StructType{
			Fields: &ast.FieldList{
				List: []*ast.Field{
					{
						Names: []*ast.Ident{{Name: "_"}},
						Type:  &ast.Ident{Name: "int"},
					},
				},
			},
		}

		fields := GetStructFields(st)

		assert.Empty(t, fields)
	})

	T.Run("handles multiple names per field", func(t *testing.T) {
		t.Parallel()

		st := &ast.StructType{
			Fields: &ast.FieldList{
				List: []*ast.Field{
					{
						Names: []*ast.Ident{
							{Name: "X", NamePos: token.NoPos},
							{Name: "Y", NamePos: token.NoPos},
						},
						Type: &ast.Ident{Name: "int"},
					},
				},
			},
		}

		fields := GetStructFields(st)

		assert.Equal(t, "int", fields["X"])
		assert.Equal(t, "int", fields["Y"])
	})
}
