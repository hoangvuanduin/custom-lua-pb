package main

import (
	"os"
	"strings"
	"testing"

	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/pluginpb"
)

func generateDataTypesOutput(t *testing.T) string {
	t.Helper()
	data, err := os.ReadFile("testdata/request.pb")
	if err != nil {
		t.Fatalf("reading request.pb: %v", err)
	}
	req := &pluginpb.CodeGeneratorRequest{}
	if err := proto.Unmarshal(data, req); err != nil {
		t.Fatalf("unmarshaling: %v", err)
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
	for _, f := range resp.File {
		if f.GetName() == "data_types.luau" {
			return f.GetContent()
		}
	}
	t.Fatal("data_types.luau not found in response")
	return ""
}

func TestDataTypesSimpleTypes(t *testing.T) {
	output := generateDataTypesOutput(t)

	// Type definitions
	for _, name := range []string{"StringType", "NumberType", "BooleanType", "EnumType", "MultipleCheckboxType", "RadioGroupType"} {
		if !strings.Contains(output, "export type "+name+" = {") {
			t.Errorf("missing type definition for %s", name)
		}
	}

	// Constructors
	for _, name := range []string{"makeStringType", "makeNumberType", "makeBooleanType", "makeEnumType", "makeMultipleCheckboxType", "makeRadioGroupType"} {
		if !strings.Contains(output, "function DataTypes."+name+"(") {
			t.Errorf("missing constructor %s", name)
		}
	}

	// Encode/decode
	for _, name := range []string{"encodeJsonStringType", "decodeJsonStringType", "encodeJsonNumberType", "decodeJsonNumberType", "encodeJsonBooleanType", "decodeJsonBooleanType", "encodeJsonEnumType", "decodeJsonEnumType", "encodeJsonMultipleCheckboxType", "decodeJsonMultipleCheckboxType", "encodeJsonRadioGroupType", "decodeJsonRadioGroupType"} {
		if !strings.Contains(output, "function DataTypes."+name+"(") {
			t.Errorf("missing codec function %s", name)
		}
	}

	// Spot-check: StringType has optional regex field
	if !strings.Contains(output, "regex: string?,") {
		t.Error("StringType.regex should be optional")
	}

	// Spot-check: NumberType has optional minValue/maxValue
	if !strings.Contains(output, "minValue: number?,") {
		t.Error("NumberType.minValue should be optional")
	}

	// Spot-check: BooleanType constructor uses special boolean default
	if !strings.Contains(output, "if c.value ~= nil then c.value else false") {
		t.Error("BooleanType constructor should use special boolean handling")
	}
}

func TestDataTypesHeader(t *testing.T) {
	output := generateDataTypesOutput(t)
	if !strings.HasPrefix(output, "--!strict\n") {
		t.Error("missing --!strict header")
	}
	if !strings.Contains(output, "local DataTypes = {}") {
		t.Error("missing local DataTypes = {}")
	}
	if !strings.HasSuffix(strings.TrimSpace(output), "return DataTypes") {
		t.Error("missing return DataTypes footer")
	}
}
