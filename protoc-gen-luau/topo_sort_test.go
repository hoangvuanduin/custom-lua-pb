package main

import (
	"testing"

	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/pluginpb"
)

func buildTopoTestFile(t *testing.T, msgNames []string, deps map[string][]string) *protogen.File {
	t.Helper()
	fdp := &descriptorpb.FileDescriptorProto{
		Name:    proto.String("test.proto"),
		Syntax:  proto.String("proto3"),
		Package: proto.String("test"),
		Options: &descriptorpb.FileOptions{
			GoPackage: proto.String("example.com/test;test"),
		},
	}
	for _, name := range msgNames {
		md := &descriptorpb.DescriptorProto{Name: proto.String(name)}
		md.Field = append(md.Field, &descriptorpb.FieldDescriptorProto{
			Name:     proto.String("id"),
			Number:   proto.Int32(1),
			Type:     descriptorpb.FieldDescriptorProto_TYPE_STRING.Enum(),
			JsonName: proto.String("id"),
		})
		if fieldDeps, ok := deps[name]; ok {
			for j, dep := range fieldDeps {
				num := int32(10 + j)
				fieldName := "ref_" + dep
				md.Field = append(md.Field, &descriptorpb.FieldDescriptorProto{
					Name:     proto.String(fieldName),
					Number:   &num,
					Type:     descriptorpb.FieldDescriptorProto_TYPE_MESSAGE.Enum(),
					TypeName: proto.String(".test." + dep),
					JsonName: proto.String(fieldName),
				})
			}
		}
		fdp.MessageType = append(fdp.MessageType, md)
	}
	req := &pluginpb.CodeGeneratorRequest{
		FileToGenerate: []string{"test.proto"},
		ProtoFile:      []*descriptorpb.FileDescriptorProto{fdp},
	}
	plugin, err := protogen.Options{}.New(req)
	if err != nil {
		t.Fatalf("protogen.New: %v", err)
	}
	for _, f := range plugin.Files {
		if f.Desc.Path() == "test.proto" {
			return f
		}
	}
	t.Fatal("test.proto not found")
	return nil
}

func TestTopoSortNoDeps(t *testing.T) {
	f := buildTopoTestFile(t, []string{"A", "B", "C"}, nil)
	sorted, err := topoSortMessages(f.Messages, f.Desc)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(sorted) != 3 {
		t.Errorf("expected 3, got %d", len(sorted))
	}
}

func TestTopoSortLinearChain(t *testing.T) {
	f := buildTopoTestFile(t, []string{"C", "B", "A"}, map[string][]string{
		"C": {"B"}, "B": {"A"},
	})
	sorted, err := topoSortMessages(f.Messages, f.Desc)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	idx := map[string]int{}
	for i, m := range sorted {
		idx[m.GoIdent.GoName] = i
	}
	if idx["A"] > idx["B"] || idx["B"] > idx["C"] {
		t.Errorf("wrong order: A@%d B@%d C@%d", idx["A"], idx["B"], idx["C"])
	}
}

func TestTopoSortDiamond(t *testing.T) {
	f := buildTopoTestFile(t, []string{"D", "B", "C", "A"}, map[string][]string{
		"D": {"B", "C"}, "B": {"A"}, "C": {"A"},
	})
	sorted, err := topoSortMessages(f.Messages, f.Desc)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	idx := map[string]int{}
	for i, m := range sorted {
		idx[m.GoIdent.GoName] = i
	}
	if idx["A"] > idx["B"] || idx["A"] > idx["C"] {
		t.Errorf("A should come before B and C")
	}
	if idx["B"] > idx["D"] || idx["C"] > idx["D"] {
		t.Errorf("B and C should come before D")
	}
}
