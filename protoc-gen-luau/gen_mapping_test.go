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

func buildMappingOutput(t *testing.T, protoName string, messages []*descriptorpb.DescriptorProto) map[string]string {
	t.Helper()

	mappingOpts := &descriptorpb.FileOptions{GoPackage: proto.String("gen/mappings;mappings")}
	proto.SetExtension(mappingOpts, luauoptions.E_LuauGenerator, "mapping")

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

	// Types file should exist with Input AND Output
	typesContent, ok := files["test_mappings_types.luau"]
	if !ok {
		t.Fatal("test_mappings_types.luau not found")
	}
	if !strings.Contains(typesContent, "export type TextboxOccupationInput = {") {
		t.Error("missing TextboxOccupationInput type")
	}
	// Input fields use actual path as struct field name
	if !strings.Contains(typesContent, "masterAsaOccupation: DataTypes.StringType?,") {
		t.Error("missing masterAsaOccupation field in input type")
	}
	if !strings.Contains(typesContent, "feederAsaOccupation: DataTypes.StringType?,") {
		t.Error("missing feederAsaOccupation field in input type")
	}
	if !strings.Contains(typesContent, "export type TextboxOccupationOutput = {") {
		t.Error("missing TextboxOccupationOutput type")
	}
	// Output field uses target field name without "target." prefix
	if !strings.Contains(typesContent, "sfOccupation: DataTypes.StringType?,") {
		t.Error("missing sfOccupation field in output type")
	}

	// Generated file
	generated, ok := files["test_mappings_generated.luau"]
	if !ok {
		t.Fatal("test_mappings_generated.luau not found")
	}

	// Glue builds input struct
	if !strings.Contains(generated, "local transformInput = {") {
		t.Error("missing transformInput struct construction")
	}
	if !strings.Contains(generated, "masterAsaOccupation = input.master and input.master.asaOccupation or nil,") {
		t.Error("missing masterAsaOccupation extraction in struct")
	}

	// Glue calls transform with struct
	if !strings.Contains(generated, "local result = Transforms.textboxOccupation(transformInput)") {
		t.Error("missing transform call with struct param")
	}

	// Glue unpacks output
	if !strings.Contains(generated, "result.sfOccupation") {
		t.Error("missing result.sfOccupation access")
	}
	if !strings.Contains(generated, "output.target.sfOccupation = result.sfOccupation") {
		t.Error("missing target assignment")
	}
}

func TestMappingMultiOutput(t *testing.T) {
	msg := &descriptorpb.DescriptorProto{
		Name: proto.String("NameSplitSigner"),
		Field: []*descriptorpb.FieldDescriptorProto{
			makeMappingField("master_signer", 1, true, "master.signer_name"),
			makeMappingField("first_name", 2, false, "target.sf_signer_first_name"),
			makeMappingField("middle_name", 3, false, "target.sf_signer_middle_name"),
			makeMappingField("last_name", 4, false, "target.sf_signer_last_name"),
		},
	}

	files := buildMappingOutput(t, "test_mappings.proto", []*descriptorpb.DescriptorProto{msg})

	typesContent := files["test_mappings_types.luau"]
	if !strings.Contains(typesContent, "export type NameSplitSignerOutput = {") {
		t.Error("missing NameSplitSignerOutput")
	}
	if !strings.Contains(typesContent, "sfSignerFirstName: DataTypes.StringType?,") {
		t.Error("missing sfSignerFirstName in output type")
	}

	generated := files["test_mappings_generated.luau"]
	if !strings.Contains(generated, "result.sfSignerFirstName") {
		t.Error("missing result.sfSignerFirstName access")
	}
}

func TestMappingCompoundPath(t *testing.T) {
	msg := &descriptorpb.DescriptorProto{
		Name: proto.String("TextboxBankName"),
		Field: []*descriptorpb.FieldDescriptorProto{
			makeMappingField("master_bank_name", 1, true, "master.wire_instructions.asa_bankname"),
			makeMappingField("bank_name", 2, false, "target.sf_bank_name"),
		},
	}

	files := buildMappingOutput(t, "test_mappings.proto", []*descriptorpb.DescriptorProto{msg})

	typesContent := files["test_mappings_types.luau"]
	// Compound source path: all segments joined
	if !strings.Contains(typesContent, "masterWireInstructionsAsaBankname: DataTypes.StringType?,") {
		t.Error("missing compound path struct field in input type")
	}

	generated := files["test_mappings_generated.luau"]
	expectedExtraction := "masterWireInstructionsAsaBankname = input.master and input.master.wireInstructions and input.master.wireInstructions.valueSubFields and input.master.wireInstructions.valueSubFields.asaBankname or nil,"
	if !strings.Contains(generated, expectedExtraction) {
		t.Errorf("missing compound path extraction.\nwant: %s", expectedExtraction)
	}
}
