package factsengine

// TODO: move to separate repo.

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"path"
	"strings"

	"github.com/santhosh-tekuri/jsonschema/v5"
)

const (
	schemasFolder = "json_schema"
)

//go:embed json_schema/*
var schemas embed.FS

var compiledSchemas map[string]*jsonschema.Schema // nolint

func init() {
	compiledSchemas = make(map[string]*jsonschema.Schema)
	entries, err := schemas.ReadDir(schemasFolder)
	if err != nil {
		panic(err)
	}

	for _, entry := range entries {
		name := entry.Name()
		fp := path.Join(schemasFolder, name)
		data, err := fs.ReadFile(schemas, fp)
		if err != nil {
			panic(err)
		}
		compiledSchema, err := jsonschema.CompileString(fp, string(data))
		if err != nil {
			panic(err)
		}
		key := strings.TrimSuffix(name, ".schema.json")
		compiledSchemas[key] = compiledSchema
	}
}

func Validate(schemaName string, jsonString []byte) error {
	var v interface{}

	schema, found := compiledSchemas[schemaName]
	if !found {
		return fmt.Errorf("schema %s not found", schemaName)
	}

	if err := json.Unmarshal(jsonString, &v); err != nil {
		return err
	}

	return schema.Validate(v)
}

func ValidateAndUnmarshall(schemaName string, jsonString []byte, dest interface{}) error {
	if err := Validate(schemaName, jsonString); err != nil {
		return err
	}

	return json.Unmarshal(jsonString, &dest)
}
