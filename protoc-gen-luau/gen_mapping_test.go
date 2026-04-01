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

func TestMappingNestedInput(t *testing.T) {
	msg := &descriptorpb.DescriptorProto{
		Name: proto.String("TextboxOccupation"),
		Field: []*descriptorpb.FieldDescriptorProto{
			makeMappingField("master_occupation", 1, true, "master.asa_occupation"),
			makeMappingField("feeder_occupation", 2, true, "feeder.asa_occupation"),
			makeMappingField("occupation", 3, false, "target.sf_occupation"),
		},
	}

	files := buildMappingOutput(t, "test_mappings.proto", []*descriptorpb.DescriptorProto{msg})

	// Types file — input nested by namespace
	typesContent := files["test_mappings_types.luau"]
	if !strings.Contains(typesContent, "export type TextboxOccupationInput = {") {
		t.Error("missing TextboxOccupationInput")
	}
	if !strings.Contains(typesContent, "master: {") {
		t.Error("missing master namespace group in input type")
	}
	if !strings.Contains(typesContent, "asaOccupation: DataTypes.StringType?,") {
		t.Error("missing asaOccupation field inside namespace")
	}
	if !strings.Contains(typesContent, "feeder: {") {
		t.Error("missing feeder namespace group")
	}

	// Output type — flat with target field name
	if !strings.Contains(typesContent, "export type TextboxOccupationOutput = {") {
		t.Error("missing TextboxOccupationOutput")
	}
	if !strings.Contains(typesContent, "sfOccupation: DataTypes.StringType?,") {
		t.Error("missing sfOccupation in output type")
	}

	// Generated glue — nested extraction
	generated := files["test_mappings_generated.luau"]
	if !strings.Contains(generated, "master = if input and input.master then {") {
		t.Error("missing nested master extraction")
	}
	if !strings.Contains(generated, "asaOccupation = input.master and input.master.asaOccupation or nil,") {
		t.Error("missing asaOccupation extraction inside master")
	}

	// Transform call
	if !strings.Contains(generated, "local result = Transforms.textboxOccupation(transformInput)") {
		t.Error("missing transform call")
	}

	// Output unpacking
	if !strings.Contains(generated, "output.target.sfOccupation = result.sfOccupation") {
		t.Error("missing target assignment")
	}
}

func TestMappingCompoundNestedInput(t *testing.T) {
	msg := &descriptorpb.DescriptorProto{
		Name: proto.String("TextboxBankName"),
		Field: []*descriptorpb.FieldDescriptorProto{
			makeMappingField("master_bank_name", 1, true, "master.wire_instructions.asa_bankname"),
			makeMappingField("bank_name", 2, false, "target.sf_bank_name"),
		},
	}

	files := buildMappingOutput(t, "test_mappings.proto", []*descriptorpb.DescriptorProto{msg})

	// Types — nested compound path
	typesContent := files["test_mappings_types.luau"]
	if !strings.Contains(typesContent, "master: {") {
		t.Error("missing master in type")
	}
	if !strings.Contains(typesContent, "wireInstructions: {") {
		t.Error("missing wireInstructions in type")
	}
	if !strings.Contains(typesContent, "valueSubFields: {") {
		t.Error("missing valueSubFields in type")
	}
	if !strings.Contains(typesContent, "asaBankname: DataTypes.StringType?,") {
		t.Error("missing asaBankname leaf in type")
	}

	// Generated — nested extraction through valueSubFields
	generated := files["test_mappings_generated.luau"]
	if !strings.Contains(generated, "valueSubFields = if input.master.wireInstructions and input.master.wireInstructions.valueSubFields then {") {
		t.Error("missing valueSubFields extraction")
	}
}

func TestMappingMultiOutput(t *testing.T) {
	msg := &descriptorpb.DescriptorProto{
		Name: proto.String("NameSplitSigner"),
		Field: []*descriptorpb.FieldDescriptorProto{
			makeMappingField("full_name", 1, true, "master.signer_name"),
			makeMappingField("first_name", 2, false, "target.sf_first_name"),
			makeMappingField("middle_name", 3, false, "target.sf_middle_name"),
			makeMappingField("last_name", 4, false, "target.sf_last_name"),
		},
	}

	files := buildMappingOutput(t, "test_mappings.proto", []*descriptorpb.DescriptorProto{msg})

	typesContent := files["test_mappings_types.luau"]
	if !strings.Contains(typesContent, "sfFirstName: DataTypes.StringType?,") {
		t.Error("missing sfFirstName in output type")
	}
	if !strings.Contains(typesContent, "sfMiddleName: DataTypes.StringType?,") {
		t.Error("missing sfMiddleName in output type")
	}
	if !strings.Contains(typesContent, "sfLastName: DataTypes.StringType?,") {
		t.Error("missing sfLastName in output type")
	}

	generated := files["test_mappings_generated.luau"]
	if !strings.Contains(generated, "result.sfFirstName") {
		t.Error("missing result.sfFirstName access")
	}
	if !strings.Contains(generated, "result.sfMiddleName") {
		t.Error("missing result.sfMiddleName access")
	}
}
