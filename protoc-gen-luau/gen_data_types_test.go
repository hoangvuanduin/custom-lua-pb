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

func TestDataTypesCompoundTypes(t *testing.T) {
	output := generateDataTypesOutput(t)

	if !strings.Contains(output, "export type FieldValue = {") {
		t.Error("missing FieldValue type")
	}

	for _, helper := range []string{"encodeOptionalStringField", "decodeOptionalStringField", "encodeOptionalNumberField", "decodeOptionalNumberField", "encodeOptionalMultipleCheckboxField", "decodeOptionalMultipleCheckboxField"} {
		if !strings.Contains(output, "local function "+helper+"(") {
			t.Errorf("missing optional helper %s", helper)
		}
	}

	compounds := []string{"PhoneFax", "DateTime", "Money", "Address", "IndividualName", "Country", "BaseContact", "SubmissionContact", "Signatory", "BankInfo", "BankAccountInfo", "WireInstructions", "BrokerageFirm", "BrokerageAccount", "ServiceContactPoint"}
	for _, name := range compounds {
		if !strings.Contains(output, "export type "+name+"SubFields = {") {
			t.Errorf("missing SubFields def for %s", name)
		}
		if !strings.Contains(output, "export type "+name+"Type = {") {
			t.Errorf("missing Type def for %s", name)
		}
		if !strings.Contains(output, "function DataTypes.make"+name+"SubFields(") {
			t.Errorf("missing SubFields constructor for %s", name)
		}
		if !strings.Contains(output, "function DataTypes.make"+name+"Type(") {
			t.Errorf("missing Type constructor for %s", name)
		}
		if !strings.Contains(output, "function DataTypes.encodeJson"+name+"SubFields(") {
			t.Errorf("missing SubFields encode for %s", name)
		}
		if !strings.Contains(output, "function DataTypes.decodeJson"+name+"SubFields(") {
			t.Errorf("missing SubFields decode for %s", name)
		}
		if !strings.Contains(output, "function DataTypes.encodeJson"+name+"Type(") {
			t.Errorf("missing Type encode for %s", name)
		}
		if !strings.Contains(output, "function DataTypes.decodeJson"+name+"Type(") {
			t.Errorf("missing Type decode for %s", name)
		}
	}
}

func TestDataTypesDispatch(t *testing.T) {
	output := generateDataTypesOutput(t)

	if !strings.Contains(output, "export type CustomCompoundType = {") {
		t.Error("missing CustomCompoundType definition")
	}
	if !strings.Contains(output, "function DataTypes.makeCustomCompoundType(") {
		t.Error("missing CustomCompoundType constructor")
	}
	if !strings.Contains(output, "function DataTypes.encodeJsonCustomCompoundType(") {
		t.Error("missing CustomCompoundType encode")
	}
	if !strings.Contains(output, "function DataTypes.decodeJsonCustomCompoundType(") {
		t.Error("missing CustomCompoundType decode")
	}

	if !strings.Contains(output, "local STRING_TYPE_IDS: {[string]: boolean} = {") {
		t.Error("missing STRING_TYPE_IDS table")
	}
	if !strings.Contains(output, "local NUMBER_TYPE_IDS: {[string]: boolean} = {") {
		t.Error("missing NUMBER_TYPE_IDS table")
	}
	if !strings.Contains(output, "local ENUM_TYPE_IDS: {[string]: boolean} = {") {
		t.Error("missing ENUM_TYPE_IDS table")
	}
	if !strings.Contains(output, "local COMPOUND_TYPE_CODECS: {[string]: {decode: string, encode: string}} = {") {
		t.Error("missing COMPOUND_TYPE_CODECS table")
	}

	if !strings.Contains(output, "function DataTypes.decodeFieldValueByTypeId(") {
		t.Error("missing decodeFieldValueByTypeId")
	}
	if !strings.Contains(output, "function DataTypes.encodeFieldValueByTypeId(") {
		t.Error("missing encodeFieldValueByTypeId")
	}

	if !strings.Contains(output, "Ssn = true") {
		t.Error("STRING_TYPE_IDS missing Ssn entry")
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
