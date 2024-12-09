package helpers

import (
	"encoding/json"
	"go/ast"
	"go/build"
	"go/parser"
	"go/token"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/Parallels/prl-devops-service/api_documentation/package/cache"
)

func CleanString(str string) string {
	str = strings.TrimSpace(strings.ReplaceAll(str, "\n", ""))
	str = strings.ReplaceAll(str, "\t", "")
	str = strings.ReplaceAll(str, "\r", "")
	str = strings.ReplaceAll(str, "\"", "")

	return str
}

func ConvertModelToJson(fileName, modelName string) string {
	isArray := strings.Contains(modelName, "[]")
	if modelName == "map[string]string" {
		return `{
  "additionalProp1": "string",
  "additionalProp2": "string",
  "additionalProp3": "string"
}`
	}

	fileSet := token.NewFileSet()
	node, err := parser.ParseFile(fileSet, fileName, nil, parser.AllErrors)
	if err != nil {
		log.Fatal(err)
	}

	structFields := make(map[string]interface{})
	found := findModelStructInASTRecursive(node, modelName, structFields, fileName)

	// If the struct wasn't found in the current file, look in imports
	if !found {
		found = findModelStructInImportsRecursive(node, modelName, structFields, fileName)
	}

	if !found {
		return cache.CacheObjectNotFound
	}

	var jsonBytes []byte
	var marshalErr error
	if isArray {
		arrayStructFields := []map[string]interface{}{structFields}
		jsonBytes, marshalErr = json.MarshalIndent(arrayStructFields, "", "  ")
	} else {
		jsonBytes, marshalErr = json.MarshalIndent(structFields, "", "  ")
	}

	if marshalErr != nil {
		return cache.CacheObjectNotFound
	}

	return string(jsonBytes)
}

func NormalizeString(str string) string {
	str = strings.TrimSpace(str)
	str = strings.ReplaceAll(str, "\n", "")
	str = strings.ReplaceAll(str, "\t", "")
	str = strings.ReplaceAll(str, "\r", "")
	str = strings.ReplaceAll(str, "\"", "")
	str = strings.ReplaceAll(str, " ", "_")
	str = strings.ReplaceAll(str, "-", "_")
	str = strings.ReplaceAll(str, ".", "_")
	str = strings.ReplaceAll(str, "/", "_")
	str = strings.ReplaceAll(str, ":", "_")
	str = strings.ReplaceAll(str, "(", "_")
	str = strings.ReplaceAll(str, ")", "_")
	str = strings.ReplaceAll(str, "[", "_")
	str = strings.ReplaceAll(str, "]", "_")
	str = strings.ReplaceAll(str, "{", "_")
	str = strings.ReplaceAll(str, "}", "_")
	str = strings.ReplaceAll(str, "<", "_")
	str = strings.ReplaceAll(str, ">", "_")
	str = strings.ReplaceAll(str, "=", "_")
	str = strings.ReplaceAll(str, ",", "_")
	str = strings.ReplaceAll(str, ";", "_")
	str = strings.ReplaceAll(str, "!", "_")
	str = strings.ReplaceAll(str, "@", "_")
	str = strings.ReplaceAll(str, "#", "_")
	str = strings.ReplaceAll(str, "$", "_")
	str = strings.ReplaceAll(str, "%", "_")
	str = strings.ReplaceAll(str, "^", "_")
	str = strings.ReplaceAll(str, "&", "_")
	str = strings.ReplaceAll(str, "*", "_")

	return strings.ToLower(str)
}

func RemoveEmptyStrings(arr []string) []string {
	var result []string
	for _, str := range arr {
		if str != "" {
			result = append(result, str)
		}
	}
	return result
}

func findModelStructInASTRecursive(node *ast.File, modelName string, structFields map[string]interface{}, fileName string) bool {
	modelName = strings.TrimPrefix(modelName, "[]")
	if strings.Contains(modelName, ".") {
		modelName = strings.Split(modelName, ".")[1]
	}
	found := false

	ast.Inspect(node, func(n ast.Node) bool {
		genDecl, ok := n.(*ast.GenDecl)
		if !ok || genDecl.Tok != token.TYPE {
			return true
		}

		for _, spec := range genDecl.Specs {
			typeSpec, ok := spec.(*ast.TypeSpec)
			if !ok || typeSpec.Name.Name != modelName {
				continue
			}

			structType, ok := typeSpec.Type.(*ast.StructType)
			if !ok {
				continue
			}

			for _, field := range structType.Fields.List {
				if len(field.Names) > 0 {
					fieldName := getJSONTag(field)
					if fieldName == "-" {
						continue // Skip fields with json:"-"
					}
					fieldType := getTypeString(field.Type)

					if IsBasicType(fieldType) {
						structFields[fieldName] = fieldType
					} else if strings.HasPrefix(fieldType, "[]") {
						// Handle slice of structs
						elementType := fieldType[2:] // Remove "[]"
						nestedFields := make(map[string]interface{})
						if findModelStructInASTRecursive(node, elementType, nestedFields, fileName) ||
							findModelStructInImportsRecursive(node, elementType, nestedFields, fileName) {
							structFields[fieldName] = []map[string]interface{}{nestedFields}
						} else {
							structFields[fieldName] = fieldType // Fallback to type name if unresolved
						}
					} else {
						// Handle nested struct
						nestedFields := make(map[string]interface{})
						if findModelStructInASTRecursive(node, fieldType, nestedFields, fileName) ||
							findModelStructInImportsRecursive(node, fieldType, nestedFields, fileName) {
							structFields[fieldName] = nestedFields
						} else {
							structFields[fieldName] = fieldType // Fallback to type name if unresolved
						}
					}
				}
			}
			found = true
			return false
		}
		return true
	})

	return found
}

func findModelStructInImportsRecursive(node *ast.File, modelName string, structFields map[string]interface{}, fileName string) bool {
	for _, imp := range node.Imports {
		importPath := strings.Trim(imp.Path.Value, `"`) // Remove quotes
		pkg, err := build.Import(importPath, ".", build.FindOnly)
		if err != nil {
			log.Printf("Failed to resolve import path %s: %v", importPath, err)
			continue
		}

		dir := pkg.Dir
		err = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if filepath.Ext(path) != ".go" {
				return nil
			}

			fileSet := token.NewFileSet()
			fileNode, err := parser.ParseFile(fileSet, path, nil, parser.AllErrors)
			if err != nil {
				return nil // Skip files that can't be parsed
			}

			if findModelStructInASTRecursive(fileNode, modelName, structFields, path) {
				return filepath.SkipDir // Stop searching once found
			}
			return nil
		})
		if err != nil {
			log.Printf("Error walking through files in package %s: %v", importPath, err)
		}

		if len(structFields) > 0 {
			return true
		}
	}

	return false
}

func getJSONTag(field *ast.Field) string {
	if field.Tag != nil {
		tag := field.Tag.Value
		tag = strings.Trim(tag, "`")
		structTag := reflect.StructTag(tag)
		jsonTag := structTag.Get("json")

		if idx := strings.Index(jsonTag, ","); idx != -1 {
			jsonTag = jsonTag[:idx]
		}

		if jsonTag != "" {
			return jsonTag
		}
	}

	if len(field.Names) > 0 {
		return field.Names[0].Name
	}

	return ""
}

func IsBasicType(fieldType string) bool {
	basicTypes := map[string]bool{
		"string": true, "int": true, "int8": true, "int16": true, "int32": true, "int64": true,
		"uint": true, "uint8": true, "uint16": true, "uint32": true, "uint64": true,
		"float32": true, "float64": true, "bool": true, "complex64": true, "complex128": true,
		"byte": true, "rune": true, "uintptr": true,
	}

	return basicTypes[fieldType]
}

func getTypeString(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.ArrayType:
		return "[]" + getTypeString(t.Elt)
	case *ast.StarExpr:
		return "*" + getTypeString(t.X)
	case *ast.SelectorExpr:
		return t.X.(*ast.Ident).Name + "." + t.Sel.Name
	case *ast.MapType:
		return "map[" + getTypeString(t.Key) + "]" + getTypeString(t.Value)
	default:
		return "unknown"
	}
}
