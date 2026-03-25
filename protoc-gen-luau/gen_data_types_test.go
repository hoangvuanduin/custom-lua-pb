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
