package main

import (
	"bytes"
	"fmt"
	"io/ioutil"

	"github.com/goccy/go-yaml"
	"github.com/iancoleman/strcase"
	"github.com/moznion/gowrtr/generator"
)

// Document Object.
type Document struct {
	Version    string `yaml:"openapi"`
	Components *Components
}

// Components Object
type Components struct {
	Schemas map[string]*Schema
}

// Schema Object
type Schema struct {
	Title            string
	MultipleOf       int `yaml:"multipleOf"`
	Maximum          int
	ExclusiveMaximum bool `yaml:"exclusiveMaximum"`
	Minimum          int
	ExclusiveMinimum bool `yaml:"exclusiveMinimum"`
	MaxLength        int  `yaml:"maxLength"`
	MinLength        int  `yaml:"minLength"`
	Pattern          string
	MaxItems         int `yaml:"maxItems"`
	MinItems         int `yaml:"minItems"`
	MaxProperties    int `yaml:"maxProperties"`
	MinProperties    int `yaml:"minProperties"`
	Required         []string
	Enum             []string

	Type                 string
	AllOf                []*Schema `yaml:"allOf"`
	OneOf                []*Schema `yaml:"oneOf"`
	AnyOf                []*Schema `yaml:"anyOf"`
	Not                  *Schema
	Items                *Schema
	Properties           map[string]*Schema
	AdditionalProperties *Schema `yaml:"additionalProperties"`
	Description          string
	Format               string
	Default              string

	Ref string `yaml:"$ref"`

	Extension map[string]interface{}
}

func load(filePath string) (*Document, error) {
	spec, _ := ioutil.ReadFile(filePath)
	buf := bytes.NewBuffer(spec)

	decoder := yaml.NewDecoder(buf, yaml.RecursiveDir(true), yaml.ReferenceDirs("spec"))

	var document Document
	_ = decoder.Decode(&document)

	return &document, nil
}

func genStruct(schemas map[string]*Schema, root *generator.Root) (*generator.Root, error) {
	for k1, v1 := range schemas {
		oaiStruct := generator.NewStruct(strcase.ToCamel(k1))
		if v1.Type == "object" {
			for k2, v2 := range v1.Properties {
				switch v2.Type {
				case "string":
					oaiStruct = oaiStruct.AddField(strcase.ToCamel(k2), "string", fmt.Sprintf("json:\"%s\"", k2))
				case "integer":
					oaiStruct = oaiStruct.AddField(strcase.ToCamel(k2), "int", fmt.Sprintf("json:\"%s\"", k2))
				case "boolean":
					oaiStruct = oaiStruct.AddField(strcase.ToCamel(k2), "bool", fmt.Sprintf("json:\"%s\"", k2))
				case "object":
					oaiStruct = oaiStruct.AddField(strcase.ToCamel(k2), strcase.ToCamel(k2), fmt.Sprintf("json:\"%s\"", k2))
					m := map[string]*Schema{k2: v2}
					root, _ = genStruct(m, root)
				case "array":
					switch v2.Items.Type {
					case "string":
						oaiStruct = oaiStruct.AddField(strcase.ToCamel(k2), "[]string", fmt.Sprintf("json:\"%s\"", k2))
					case "integer":
						oaiStruct = oaiStruct.AddField(strcase.ToCamel(k2), "[]int", fmt.Sprintf("json:\"%s\"", k2))
					case "object":
						oaiStruct = oaiStruct.AddField(strcase.ToCamel(k2), "[]"+strcase.ToCamel(k2), fmt.Sprintf("json:\"%s\"", k2))
						m := map[string]*Schema{k2: v2.Items}
						root, _ = genStruct(m, root)
					default:
						//
					}
				default:
					//
				}
			}
		}
		root = root.AddStatements(oaiStruct).Gofmt("-s").Goimports()
	}

	return root, nil
}

func main() {
	doc, _ := load("./spec/api.yaml")

	root := generator.NewRoot(
		generator.NewComment(" THIS CODE WAS AUTO GENERATED; DO NOT EDIT."),
		generator.NewPackage("main"),
		generator.NewNewline(),
	)

	root, _ = genStruct(doc.Components.Schemas, root)

	generated, err := root.Generate(0)
	if err != nil {
		panic(err)
	}
	fmt.Println(generated)

}
