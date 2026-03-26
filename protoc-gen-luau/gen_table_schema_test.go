package main

import (
	"os"
	"path"
	"strings"
	"testing"

	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/pluginpb"

	"protoc-gen-luau/luauoptions"
)

func getGeneratedFile(t *testing.T, name string) string {
	t.Helper()
	data, err := os.ReadFile("testdata/request.pb")
	if err != nil {
		t.Fatalf("reading request.pb: %v", err)
	}
	req := &pluginpb.CodeGeneratorRequest{}
	if err := proto.Unmarshal(data, req); err != nil {
		t.Fatalf("unmarshaling: %v", err)
	}
	for _, pf := range req.ProtoFile {
		if pf.Options == nil {
			pf.Options = &descriptorpb.FileOptions{}
		}
		baseName := path.Base(pf.GetName())
		baseName = strings.TrimSuffix(baseName, ".proto")
		switch baseName {
		case "data_types":
			proto.SetExtension(pf.Options, luauoptions.E_LuauGenerator, "data_types")
		case "table_schema", "source_table", "target_table":
			proto.SetExtension(pf.Options, luauoptions.E_LuauGenerator, "schema")
		}
	}
	plugin, err := protogen.Options{}.New(req)
	if err != nil {
		t.Fatalf("protogen.New: %v", err)
	}
	if err := generate(plugin); err != nil {
		t.Fatalf("generate: %v", err)
	}
	resp := plugin.Response()
	for _, f := range resp.File {
		if f.GetName() == name {
			return f.GetContent()
		}
	}
	t.Fatalf("file %s not found in response", name)
	return ""
}

func TestTableSchemaOutput(t *testing.T) {
	output := getGeneratedFile(t, "table_schema.luau")

	if !strings.Contains(output, `local DataTypes = require("@lib/data_types")`) {
		t.Error("missing DataTypes require")
	}
	if !strings.Contains(output, "local TableSchema = {}") {
		t.Error("missing TableSchema module")
	}
	if !strings.Contains(output, "export type SingleFieldType = {") {
		t.Error("missing SingleFieldType def")
	}
	// SingleFieldType has individual DataTypes fields (stringType, numberType, etc.)
	if !strings.Contains(output, "stringType: DataTypes.StringType?,") {
		t.Error("SingleFieldType should have stringType: DataTypes.StringType?")
	}
	if !strings.Contains(output, "numberType: DataTypes.NumberType?,") {
		t.Error("SingleFieldType should have numberType: DataTypes.NumberType?")
	}
	if !strings.Contains(output, "label: string,") {
		t.Error("SingleFieldType should have required label: string")
	}
	if !strings.Contains(output, "export type FieldGroup = {") {
		t.Error("missing FieldGroup def")
	}
	if !strings.Contains(output, "export type TableSchema = {") {
		t.Error("missing TableSchema def")
	}
	if !strings.Contains(output, "fieldsMap: {[string]: SingleFieldType},") {
		t.Error("TableSchema.fieldsMap should be map of SingleFieldType")
	}
	for _, fn := range []string{"makeSingleFieldType", "makeFieldGroup", "makeTableSchema"} {
		if !strings.Contains(output, "function TableSchema."+fn+"(") {
			t.Errorf("missing constructor %s", fn)
		}
	}
	for _, fn := range []string{"encodeJsonSingleFieldType", "decodeJsonSingleFieldType", "encodeJsonFieldGroup", "decodeJsonFieldGroup", "encodeJsonTableSchema", "decodeJsonTableSchema"} {
		if !strings.Contains(output, "function TableSchema."+fn+"(") {
			t.Errorf("missing codec %s", fn)
		}
	}
	// SingleFieldType encode should use optional helpers for DataTypes types
	if !strings.Contains(output, "encodeOptionalStringType(") {
		t.Error("SingleFieldType encode should use encodeOptionalStringType helper")
	}
	if !strings.HasSuffix(strings.TrimSpace(output), "return TableSchema") {
		t.Error("missing return TableSchema footer")
	}
}

func TestSourceTableOutput(t *testing.T) {
	output := getGeneratedFile(t, "source_table.luau")

	// Header
	if !strings.Contains(output, `local DataTypes = require("@lib/data_types")`) {
		t.Error("missing DataTypes require")
	}
	if !strings.Contains(output, "local SourceTable = {}") {
		t.Error("missing SourceTable module")
	}

	// Local compound type definitions
	if !strings.Contains(output, "export type LpSignatoryFields = {") {
		t.Error("missing LpSignatoryFields type def")
	}
	if !strings.Contains(output, "export type LpSignatoryType = {") {
		t.Error("missing LpSignatoryType type def")
	}
	if !strings.Contains(output, "export type W9Fields = {") {
		t.Error("missing W9Fields type def")
	}
	if !strings.Contains(output, "export type W9Type = {") {
		t.Error("missing W9Type type def")
	}

	// LpSignatoryFields should reference DataTypes.StringType?
	if !strings.Contains(output, "asaCommitmentAmount: DataTypes.StringType?,") {
		t.Error("LpSignatoryFields.asaCommitmentAmount should be DataTypes.StringType?")
	}

	// FieldsMap type def with mixed local and DataTypes references
	if !strings.Contains(output, "export type SourceTableFieldsMap = {") {
		t.Error("missing SourceTableFieldsMap type def")
	}
	if !strings.Contains(output, "lpSignatory: LpSignatoryType?,") {
		t.Error("FieldsMap lpSignatory should be local LpSignatoryType?")
	}
	if !strings.Contains(output, "asaFullnameInvestornameAmlquestionnaire: DataTypes.StringType?,") {
		t.Error("FieldsMap imported field should be DataTypes.StringType?")
	}

	// Schema type def
	if !strings.Contains(output, "export type SourceTableSchema = {") {
		t.Error("missing SourceTableSchema type def")
	}
	if !strings.Contains(output, "fieldsMap: SourceTableFieldsMap?,") {
		t.Error("Schema fieldsMap should be SourceTableFieldsMap?")
	}

	// Constructors for compound types
	if !strings.Contains(output, "function SourceTable.makeLpSignatoryFields(") {
		t.Error("missing makeLpSignatoryFields constructor")
	}
	if !strings.Contains(output, "function SourceTable.makeLpSignatoryType(") {
		t.Error("missing makeLpSignatoryType constructor")
	}
	if !strings.Contains(output, "function SourceTable.makeW9Fields(") {
		t.Error("missing makeW9Fields constructor")
	}
	if !strings.Contains(output, "function SourceTable.makeW9Type(") {
		t.Error("missing makeW9Type constructor")
	}

	// FieldsMap and Schema constructors
	if !strings.Contains(output, "function SourceTable.makeSourceTableFieldsMap(") {
		t.Error("missing makeSourceTableFieldsMap constructor")
	}
	if !strings.Contains(output, "function SourceTable.makeSourceTableSchema(") {
		t.Error("missing makeSourceTableSchema constructor")
	}

	// Per-module optional helpers for DataTypes types used in FieldsMap
	if !strings.Contains(output, "local function decodeOptionalStringType(json: any): DataTypes.StringType?") {
		t.Error("missing decodeOptionalStringType helper")
	}
	if !strings.Contains(output, "local function encodeOptionalStringType(v: DataTypes.StringType?): {[string]: any}?") {
		t.Error("missing encodeOptionalStringType helper")
	}
	if !strings.Contains(output, "local function decodeOptionalMultipleCheckboxType(json: any): DataTypes.MultipleCheckboxType?") {
		t.Error("missing decodeOptionalMultipleCheckboxType helper")
	}
	if !strings.Contains(output, "local function encodeOptionalMultipleCheckboxType(v: DataTypes.MultipleCheckboxType?): {[string]: any}?") {
		t.Error("missing encodeOptionalMultipleCheckboxType helper")
	}

	// Compound codecs
	if !strings.Contains(output, "function SourceTable.encodeJsonLpSignatoryFields(") {
		t.Error("missing encodeJsonLpSignatoryFields")
	}
	if !strings.Contains(output, "function SourceTable.decodeJsonLpSignatoryFields(") {
		t.Error("missing decodeJsonLpSignatoryFields")
	}
	if !strings.Contains(output, "function SourceTable.encodeJsonLpSignatoryType(") {
		t.Error("missing encodeJsonLpSignatoryType")
	}

	// FieldsMap encode — local compound uses ModuleName.encodeJson pattern
	if !strings.Contains(output, "function SourceTable.encodeJsonSourceTableFieldsMap(") {
		t.Error("missing encodeJsonSourceTableFieldsMap")
	}
	if !strings.Contains(output, "SourceTable.encodeJsonLpSignatoryType(so.lpSignatory)") {
		t.Error("FieldsMap encode should use SourceTable.encodeJsonLpSignatoryType for local compound")
	}

	// FieldsMap decode — local compound uses if-then-else pattern
	if !strings.Contains(output, "function SourceTable.decodeJsonSourceTableFieldsMap(") {
		t.Error("missing decodeJsonSourceTableFieldsMap")
	}
	if !strings.Contains(output, "SourceTable.decodeJsonLpSignatoryType(json.lpSignatory)") {
		t.Error("FieldsMap decode should use SourceTable.decodeJsonLpSignatoryType for local compound")
	}

	// Schema codecs
	if !strings.Contains(output, "function SourceTable.encodeJsonSourceTableSchema(") {
		t.Error("missing encodeJsonSourceTableSchema")
	}
	if !strings.Contains(output, "function SourceTable.decodeJsonSourceTableSchema(") {
		t.Error("missing decodeJsonSourceTableSchema")
	}

	// Footer
	if !strings.HasSuffix(strings.TrimSpace(output), "return SourceTable") {
		t.Error("missing return SourceTable footer")
	}
}

func TestTargetTableOutput(t *testing.T) {
	output := getGeneratedFile(t, "target_table.luau")

	// Header
	if !strings.Contains(output, `local DataTypes = require("@lib/data_types")`) {
		t.Error("missing DataTypes require")
	}
	if !strings.Contains(output, "local TargetTable = {}") {
		t.Error("missing TargetTable module")
	}

	// No local compound types
	if strings.Contains(output, "SubFields = {") {
		t.Error("target_table should have no local compound SubFields types")
	}

	// FieldsMap type def
	if !strings.Contains(output, "export type TargetTableFieldsMap = {") {
		t.Error("missing TargetTableFieldsMap type def")
	}
	if !strings.Contains(output, "sfAgreementNullCommitmentC: DataTypes.MoneyType?,") {
		t.Error("FieldsMap should have MoneyType field")
	}
	if !strings.Contains(output, "sfAccountSubscriptionInvestorName: DataTypes.StringType?,") {
		t.Error("FieldsMap should have StringType field")
	}
	if !strings.Contains(output, "sfAccountSubscriptionInvestorWlcPubliclyListedOnAStockExchangeC: DataTypes.RadioGroupType?,") {
		t.Error("FieldsMap should have RadioGroupType field")
	}

	// Schema type def
	if !strings.Contains(output, "export type TargetTableSchema = {") {
		t.Error("missing TargetTableSchema type def")
	}

	// Constructors
	if !strings.Contains(output, "function TargetTable.makeTargetTableFieldsMap(") {
		t.Error("missing makeTargetTableFieldsMap constructor")
	}
	if !strings.Contains(output, "function TargetTable.makeTargetTableSchema(") {
		t.Error("missing makeTargetTableSchema constructor")
	}

	// Per-module optional helpers — 4 types: StringType, MultipleCheckboxType, RadioGroupType, MoneyType
	if !strings.Contains(output, "local function decodeOptionalStringType(") {
		t.Error("missing decodeOptionalStringType helper")
	}
	if !strings.Contains(output, "local function decodeOptionalMultipleCheckboxType(") {
		t.Error("missing decodeOptionalMultipleCheckboxType helper")
	}
	if !strings.Contains(output, "local function decodeOptionalRadioGroupType(") {
		t.Error("missing decodeOptionalRadioGroupType helper")
	}
	if !strings.Contains(output, "local function decodeOptionalMoneyType(") {
		t.Error("missing decodeOptionalMoneyType helper")
	}

	// FieldsMap encode/decode
	if !strings.Contains(output, "function TargetTable.encodeJsonTargetTableFieldsMap(") {
		t.Error("missing encodeJsonTargetTableFieldsMap")
	}
	if !strings.Contains(output, "function TargetTable.decodeJsonTargetTableFieldsMap(") {
		t.Error("missing decodeJsonTargetTableFieldsMap")
	}

	// All fields use optional helpers (no local compounds in target_table)
	if !strings.Contains(output, "encodeOptionalMoneyType(ta.sfAgreementNullCommitmentC)") {
		t.Error("FieldsMap encode should use encodeOptionalMoneyType for MoneyType field")
	}
	if !strings.Contains(output, "decodeOptionalStringType(json.sfAccountSubscriptionInvestorName)") {
		t.Error("FieldsMap decode should use decodeOptionalStringType for StringType field")
	}

	// Schema codecs
	if !strings.Contains(output, "function TargetTable.encodeJsonTargetTableSchema(") {
		t.Error("missing encodeJsonTargetTableSchema")
	}
	if !strings.Contains(output, "function TargetTable.decodeJsonTargetTableSchema(") {
		t.Error("missing decodeJsonTargetTableSchema")
	}

	// Footer
	if !strings.HasSuffix(strings.TrimSpace(output), "return TargetTable") {
		t.Error("missing return TargetTable footer")
	}
}
