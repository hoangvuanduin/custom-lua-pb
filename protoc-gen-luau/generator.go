package main

import (
	"fmt"

	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"

	"protoc-gen-luau/luauoptions"
)

// LuauGenerator is the interface for code generation strategies.
type LuauGenerator interface {
	Generate(plugin *protogen.Plugin, file *protogen.File)
}

// generators maps luau_generator option values to implementations.
var generators = map[string]LuauGenerator{
	"data_types": &DataTypesGenerator{},
	"schema":     &SchemaGenerator{},
}

// getGeneratorType reads the luau_generator file option from a proto file.
func getGeneratorType(f *protogen.File) string {
	opts := f.Proto.GetOptions()
	if opts == nil {
		return ""
	}
	if !proto.HasExtension(opts, luauoptions.E_LuauGenerator) {
		return ""
	}
	return proto.GetExtension(opts, luauoptions.E_LuauGenerator).(string)
}

// getDataTypesRequirePath reads the luau_data_types_require file option.
func getDataTypesRequirePath(f *protogen.File) string {
	opts := f.Proto.GetOptions()
	if opts == nil {
		return "@lib/data_types"
	}
	if !proto.HasExtension(opts, luauoptions.E_LuauDataTypesRequire) {
		return "@lib/data_types"
	}
	v := proto.GetExtension(opts, luauoptions.E_LuauDataTypesRequire).(string)
	if v == "" {
		return "@lib/data_types"
	}
	return v
}

// DataTypesGenerator generates the data_types.luau module.
type DataTypesGenerator struct{}

func (d *DataTypesGenerator) Generate(plugin *protogen.Plugin, file *protogen.File) {
	generateDataTypes(plugin, file)
}

// SchemaGenerator generates generic Luau modules from any proto schema.
type SchemaGenerator struct{}

func (s *SchemaGenerator) Generate(plugin *protogen.Plugin, file *protogen.File) {
	generateSchema(plugin, file)
}

// lookupGenerator returns the generator for the given type, or an error.
func lookupGenerator(genType string) (LuauGenerator, error) {
	gen, ok := generators[genType]
	if !ok {
		return nil, fmt.Errorf("unknown luau_generator %q", genType)
	}
	return gen, nil
}
