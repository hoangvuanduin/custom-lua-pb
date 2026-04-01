package main

import (
	"strings"
	"testing"

	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/pluginpb"

	"protoc-gen-luau/luauoptions"
	"protoc-gen-luau/mappingoptions"
)

// buildMappingOutput creates a CodeGeneratorRequest with a mapping proto file,
// runs the generator, and returns the content of the generated files.
func buildMappingOutput(t *testing.T, protoName string, messages []*descriptorpb.DescriptorProto) map[string]string {
	t.Helper()

	mappingOpts := &descriptorpb.FileOptions{GoPackage: proto.String("gen/mappings;mappings")}
	proto.SetExtension(mappingOpts, luauoptions.E_LuauGenerator, "mapping")

	// data_types.proto must exist for field type resolution
	dtOpts := &descriptorpb.FileOptions{GoPackage: proto.String("gen/datatypes;datatypes")}
	proto.SetExtension(dtOpts, luauoptions.E_LuauGenerator, "data_types")

	req := &pluginpb.CodeGeneratorRequest{
		FileToGenerate: []string{"data_types.proto", protoName},
		ProtoFile: []*descriptorpb.FileDescriptorProto{
			{
				Name:    proto.String("data_types.proto"),
				Package: proto.String("data_types"),
				Syntax:  proto.String("proto3"),
				Options: dtOpts,
				MessageType: []*descriptorpb.DescriptorProto{
					makeScalarMessage("StringType",
						stringField("type_id", 1),
						stringField("value", 2),
					),
				},
			},
			{
				Name:        proto.String(protoName),
				Package:     proto.String("test_mappings"),
				Syntax:      proto.String("proto3"),
				Options:     mappingOpts,
				Dependency:  []string{"data_types.proto"},
				MessageType: messages,
			},
		},
	}

	plugin, err := protogen.Options{}.New(req)
	if err != nil {
		t.Fatalf("protogen.New: %v", err)
	}
	if err := generate(plugin); err != nil {
		t.Fatalf("generate: %v", err)
	}
	resp := plugin.Response()
	if resp.GetError() != "" {
		t.Fatalf("plugin error: %s", resp.GetError())
	}

	result := make(map[string]string)
	for _, f := range resp.File {
		result[f.GetName()] = f.GetContent()
	}
	return result
}

// makeMappingField creates a field descriptor with a source or target annotation.
func makeMappingField(name string, number int32, isSource bool, path string) *descriptorpb.FieldDescriptorProto {
	fd := &descriptorpb.FieldDescriptorProto{
		Name:     proto.String(name),
		Number:   proto.Int32(number),
		Type:     descriptorpb.FieldDescriptorProto_TYPE_MESSAGE.Enum(),
		Label:    descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL.Enum(),
		TypeName: proto.String(".data_types.StringType"),
		Options:  &descriptorpb.FieldOptions{},
	}
	if isSource {
		proto.SetExtension(fd.Options, mappingoptions.E_Source, path)
	} else {
		proto.SetExtension(fd.Options, mappingoptions.E_Target, path)
	}
	return fd
}

// TestMappingSingleOutput verifies a basic single-output rule generates correct
// trace metadata, extraction, transform call, and target assignment.
func TestMappingSingleOutput(t *testing.T) {
	msg := &descriptorpb.DescriptorProto{
		Name: proto.String("TextboxOccupation"),
		Field: []*descriptorpb.FieldDescriptorProto{
			makeMappingField("master_occupation", 1, true, "master.asa_occupation"),
			makeMappingField("feeder_occupation", 2, true, "feeder.asa_occupation"),
			makeMappingField("occupation", 3, false, "target.sf_occupation"),
		},
	}

	files := buildMappingOutput(t, "test_mappings.proto", []*descriptorpb.DescriptorProto{msg})

	generated, ok := files["test_mappings_generated.luau"]
	if !ok {
		t.Fatal("test_mappings_generated.luau not found in output")
	}

	// Header imports
	if !strings.Contains(generated, `local Transforms = require("@lib/test_transforms")`) {
		t.Error("missing or incorrect Transforms require path")
	}
	if !strings.Contains(generated, `local SourceTables = require("@lib/source_tables")`) {
		t.Error("missing SourceTables require")
	}
	if !strings.Contains(generated, `local TargetTables = require("@lib/target_tables")`) {
		t.Error("missing TargetTables require")
	}

	// No trace metadata
	if strings.Contains(generated, "mappingTrace") {
		t.Error("should not contain mappingTrace — tracing comes from proto directly")
	}

	// Glue function
	if !strings.Contains(generated, "function Generated.runTextboxOccupation(") {
		t.Error("missing runTextboxOccupation function")
	}
	// Source extraction (2-segment)
	if !strings.Contains(generated, "local masterOccupation = input.master and input.master.asaOccupation or nil") {
		t.Error("missing masterOccupation extraction")
	}
	if !strings.Contains(generated, "local feederOccupation = input.feeder and input.feeder.asaOccupation or nil") {
		t.Error("missing feederOccupation extraction")
	}
	// Transform call
	if !strings.Contains(generated, "Transforms.textboxOccupation(masterOccupation, feederOccupation)") {
		t.Error("missing transform call with correct params")
	}
	// Target assignment
	if !strings.Contains(generated, "output.target.sfOccupation = occupation") {
		t.Error("missing target assignment")
	}

	// Single output comment
	if !strings.Contains(generated, "-- Single output glue") {
		t.Error("missing single output comment")
	}

	// runAll
	if !strings.Contains(generated, "function Generated.runAll(") {
		t.Error("missing runAll function")
	}
	if !strings.Contains(generated, "Generated.runTextboxOccupation(input, output)") {
		t.Error("missing runTextboxOccupation in runAll")
	}

	// No types file for single-output
	if _, ok := files["test_mappings_types.luau"]; ok {
		t.Error("types file should not be generated for single-output rules only")
	}
}

// TestMappingMultiOutput verifies multi-output rules generate a types file and
// use result struct access pattern.
func TestMappingMultiOutput(t *testing.T) {
	msg := &descriptorpb.DescriptorProto{
		Name: proto.String("NameSplitSigner"),
		Field: []*descriptorpb.FieldDescriptorProto{
			makeMappingField("master_signer", 1, true, "master.signer_name"),
			makeMappingField("signer_first_name", 2, false, "target.sf_signer_first_name"),
			makeMappingField("signer_middle_name", 3, false, "target.sf_signer_middle_name"),
			makeMappingField("signer_last_name", 4, false, "target.sf_signer_last_name"),
		},
	}

	files := buildMappingOutput(t, "test_mappings.proto", []*descriptorpb.DescriptorProto{msg})

	// Types file should exist
	typesContent, ok := files["test_mappings_types.luau"]
	if !ok {
		t.Fatal("test_mappings_types.luau not found — should be generated for multi-output rules")
	}

	// Type definition
	if !strings.Contains(typesContent, "export type NameSplitSignerOutput = {") {
		t.Error("missing NameSplitSignerOutput type definition")
	}
	if !strings.Contains(typesContent, "\tsignerFirstName: DataTypes.StringType?,") {
		t.Error("missing signerFirstName field in output type")
	}
	if !strings.Contains(typesContent, "\tsignerMiddleName: DataTypes.StringType?,") {
		t.Error("missing signerMiddleName field in output type")
	}
	if !strings.Contains(typesContent, "\tsignerLastName: DataTypes.StringType?,") {
		t.Error("missing signerLastName field in output type")
	}

	// Generated file
	generated, ok := files["test_mappings_generated.luau"]
	if !ok {
		t.Fatal("test_mappings_generated.luau not found")
	}

	// Multi output comment
	if !strings.Contains(generated, "-- Multi output glue") {
		t.Error("missing multi output comment")
	}

	// Result struct access pattern
	if !strings.Contains(generated, "local result = Transforms.nameSplitSigner(masterSigner)") {
		t.Error("missing transform call with result assignment")
	}
	if !strings.Contains(generated, "if result.signerFirstName ~= nil then output.target.sfSignerFirstName = result.signerFirstName end") {
		t.Error("missing result.signerFirstName access")
	}
	if !strings.Contains(generated, "if result.signerMiddleName ~= nil then output.target.sfSignerMiddleName = result.signerMiddleName end") {
		t.Error("missing result.signerMiddleName access")
	}
	if !strings.Contains(generated, "if result.signerLastName ~= nil then output.target.sfSignerLastName = result.signerLastName end") {
		t.Error("missing result.signerLastName access")
	}
}

// TestMappingCompoundPath verifies 3-segment source paths insert valueSubFields.
func TestMappingCompoundPath(t *testing.T) {
	msg := &descriptorpb.DescriptorProto{
		Name: proto.String("TextboxBankName"),
		Field: []*descriptorpb.FieldDescriptorProto{
			makeMappingField("master_bank_name", 1, true, "master.wire_instructions.asa_bankname"),
			makeMappingField("feeder_bank_name", 2, true, "feeder.wire_instructions.asa_bankname"),
			makeMappingField("bank_name", 3, false, "target.sf_bank_name"),
		},
	}

	files := buildMappingOutput(t, "test_mappings.proto", []*descriptorpb.DescriptorProto{msg})

	generated, ok := files["test_mappings_generated.luau"]
	if !ok {
		t.Fatal("test_mappings_generated.luau not found")
	}

	// 3-segment extraction with valueSubFields
	expectedExtraction := "input.master and input.master.wireInstructions and input.master.wireInstructions.valueSubFields and input.master.wireInstructions.valueSubFields.asaBankname or nil"
	if !strings.Contains(generated, expectedExtraction) {
		t.Errorf("missing compound path extraction.\nwant: %s", expectedExtraction)
	}

	expectedFeeder := "input.feeder and input.feeder.wireInstructions and input.feeder.wireInstructions.valueSubFields and input.feeder.wireInstructions.valueSubFields.asaBankname or nil"
	if !strings.Contains(generated, expectedFeeder) {
		t.Errorf("missing feeder compound path extraction.\nwant: %s", expectedFeeder)
	}

}
