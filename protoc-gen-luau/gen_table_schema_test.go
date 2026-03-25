package main

import (
	"os"
	"strings"
	"testing"

	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/pluginpb"
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
	if !strings.Contains(output, "value: DataTypes.FieldValue?,") {
		t.Error("SingleFieldType.value should be DataTypes.FieldValue?")
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
	for _, fn := range []string{"encodeJsonSingleFieldType", "decodeJsonSingleFieldType", "encodeJsonTableSchema", "decodeJsonTableSchema"} {
		if !strings.Contains(output, "function TableSchema."+fn+"(") {
			t.Errorf("missing codec %s", fn)
		}
	}
	if !strings.Contains(output, "DataTypes.encodeFieldValueByTypeId") {
		t.Error("SingleFieldType encode should use DataTypes.encodeFieldValueByTypeId")
	}
	if !strings.HasSuffix(strings.TrimSpace(output), "return TableSchema") {
		t.Error("missing return TableSchema footer")
	}
}
