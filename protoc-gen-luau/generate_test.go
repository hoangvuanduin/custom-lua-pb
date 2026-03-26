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

func TestGenerate(t *testing.T) {
	dtOpts := &descriptorpb.FileOptions{GoPackage: proto.String("protoc-gen-luau/gen/datatypes;datatypes")}
	proto.SetExtension(dtOpts, luauoptions.E_LuauGenerator, "data_types")

	tsOpts := &descriptorpb.FileOptions{GoPackage: proto.String("protoc-gen-luau/gen/tableschema;tableschema")}
	proto.SetExtension(tsOpts, luauoptions.E_LuauGenerator, "schema")

	stOpts := &descriptorpb.FileOptions{GoPackage: proto.String("protoc-gen-luau/gen/tables/source;source")}
	proto.SetExtension(stOpts, luauoptions.E_LuauGenerator, "schema")

	ttOpts := &descriptorpb.FileOptions{GoPackage: proto.String("protoc-gen-luau/gen/tables/target;target")}
	proto.SetExtension(ttOpts, luauoptions.E_LuauGenerator, "schema")

	req := &pluginpb.CodeGeneratorRequest{
		FileToGenerate: []string{
			"data_types.proto", "table_schema.proto",
			"tables/source_table.proto", "tables/target_table.proto",
		},
		ProtoFile: []*descriptorpb.FileDescriptorProto{
			{Name: proto.String("data_types.proto"), Syntax: proto.String("proto3"), Options: dtOpts},
			{Name: proto.String("table_schema.proto"), Syntax: proto.String("proto3"),
				Dependency: []string{"data_types.proto"}, Options: tsOpts},
			{Name: proto.String("tables/source_table.proto"), Syntax: proto.String("proto3"),
				Dependency: []string{"data_types.proto"}, Options: stOpts},
			{Name: proto.String("tables/target_table.proto"), Syntax: proto.String("proto3"),
				Dependency: []string{"data_types.proto"}, Options: ttOpts},
		},
	}

	plugin, err := protogen.Options{}.New(req)
	if err != nil { t.Fatalf("protogen.New: %v", err) }
	if err := generate(plugin); err != nil { t.Fatalf("generate: %v", err) }
	resp := plugin.Response()
	if resp.GetError() != "" { t.Fatalf("plugin error: %s", resp.GetError()) }

	wantNames := map[string]bool{
		"data_types.luau": true, "table_schema.luau": true,
		"source_table.luau": true, "target_table.luau": true,
	}
	if len(resp.File) != len(wantNames) {
		t.Errorf("got %d files, want %d", len(resp.File), len(wantNames))
	}
	for _, f := range resp.File {
		if !wantNames[f.GetName()] { t.Errorf("unexpected file %q", f.GetName()) }
	}
}

func TestGenerateGolden(t *testing.T) {
	data, err := os.ReadFile("testdata/request.pb")
	if err != nil {
		t.Fatalf("reading request.pb: %v", err)
	}

	req := &pluginpb.CodeGeneratorRequest{}
	if err := proto.Unmarshal(data, req); err != nil {
		t.Fatalf("unmarshaling request: %v", err)
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
	if resp.GetError() != "" {
		t.Fatalf("plugin error: %s", resp.GetError())
	}

	if got, want := len(resp.File), 4; got != want {
		t.Errorf("got %d output files, want %d", got, want)
	}

	updateGolden := os.Getenv("UPDATE_GOLDEN") == "1"

	for _, f := range resp.File {
		goldenPath := "testdata/golden/" + f.GetName()
		if updateGolden {
			if err := os.MkdirAll("testdata/golden", 0755); err != nil {
				t.Fatalf("mkdir golden: %v", err)
			}
			if err := os.WriteFile(goldenPath, []byte(f.GetContent()), 0644); err != nil {
				t.Fatalf("writing golden %s: %v", goldenPath, err)
			}
			t.Logf("updated golden: %s", goldenPath)
			continue
		}
		want, err := os.ReadFile(goldenPath)
		if err != nil {
			t.Errorf("reading golden %s: %v", goldenPath, err)
			continue
		}
		if f.GetContent() != string(want) {
			t.Errorf("file %s differs from golden:\ngot:\n%s\nwant:\n%s",
				f.GetName(), f.GetContent(), string(want))
		}
	}
}

func TestGenerateWithOptions(t *testing.T) {
	opts := &descriptorpb.FileOptions{GoPackage: proto.String("protoc-gen-luau/gen/datatypes;datatypes")}
	proto.SetExtension(opts, luauoptions.E_LuauGenerator, "data_types")

	req := &pluginpb.CodeGeneratorRequest{
		FileToGenerate: []string{"data_types.proto"},
		ProtoFile: []*descriptorpb.FileDescriptorProto{
			{Name: proto.String("data_types.proto"), Syntax: proto.String("proto3"), Options: opts},
		},
	}

	plugin, err := protogen.Options{}.New(req)
	if err != nil { t.Fatalf("protogen.New: %v", err) }
	if err := generate(plugin); err != nil { t.Fatalf("generate: %v", err) }
	resp := plugin.Response()
	if resp.GetError() != "" { t.Fatalf("plugin error: %s", resp.GetError()) }

	found := false
	for _, f := range resp.File {
		if f.GetName() == "data_types.luau" { found = true }
	}
	if !found { t.Error("expected data_types.luau in output") }
}

func TestGenerateUnknownOption(t *testing.T) {
	opts := &descriptorpb.FileOptions{GoPackage: proto.String("protoc-gen-luau/gen/foo;foo")}
	proto.SetExtension(opts, luauoptions.E_LuauGenerator, "unknown_gen")

	req := &pluginpb.CodeGeneratorRequest{
		FileToGenerate: []string{"foo.proto"},
		ProtoFile: []*descriptorpb.FileDescriptorProto{
			{Name: proto.String("foo.proto"), Syntax: proto.String("proto3"), Options: opts},
		},
	}

	plugin, err := protogen.Options{}.New(req)
	if err != nil { t.Fatalf("protogen.New: %v", err) }
	err = generate(plugin)
	if err == nil { t.Fatal("expected error for unknown generator type") }
	if !strings.Contains(err.Error(), "unknown_gen") {
		t.Errorf("error should mention unknown type, got: %v", err)
	}
}

func TestGenerateNoOption(t *testing.T) {
	req := &pluginpb.CodeGeneratorRequest{
		FileToGenerate: []string{"ignored.proto"},
		ProtoFile: []*descriptorpb.FileDescriptorProto{
			{Name: proto.String("ignored.proto"), Syntax: proto.String("proto3"),
				Options: &descriptorpb.FileOptions{GoPackage: proto.String("protoc-gen-luau/gen/ignored;ignored")}},
		},
	}

	plugin, err := protogen.Options{}.New(req)
	if err != nil { t.Fatalf("protogen.New: %v", err) }
	if err := generate(plugin); err != nil { t.Fatalf("generate: %v", err) }
	resp := plugin.Response()
	if len(resp.File) != 0 { t.Errorf("expected no output files, got %d", len(resp.File)) }
}
