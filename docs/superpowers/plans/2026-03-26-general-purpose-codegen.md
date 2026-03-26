# General-Purpose Proto-to-Luau Codegen Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Refactor protoc-gen-luau from hardcoded 4-module generation to extensible strategy pattern with proto option dispatch and a generic SchemaGenerator.

**Architecture:** Strategy pattern with Go interface. Proto file-level custom options (`luau_generator`) select which `LuauGenerator` implementation handles the file. `DataTypesGenerator` wraps existing logic unchanged. `SchemaGenerator` replaces both `gen_table_schema.go` and `gen_table_files.go` with generic per-message codegen.

**Tech Stack:** Go 1.26, google.golang.org/protobuf v1.36.11 (protogen + proto extensions)

**Spec:** `docs/specs/2026-03-26-general-purpose-codegen-design.md`

---

## File Structure

| File | Action | Responsibility |
|---|---|---|
| `luauoptions/luau_options.proto` | CREATE | Proto file-level option extensions |
| `luauoptions/luau_options.pb.go` | CREATE (generated) | Go extension descriptors for proto options |
| `generator.go` | CREATE | `LuauGenerator` interface + registry + option reader |
| `generate.go` | MODIFY | Replace if/else dispatch with registry lookup |
| `gen_data_types.go` | MODIFY | Wrap `generateDataTypes` in `DataTypesGenerator` struct; move `classifyMessages` + related types to own file |
| `classify.go` | CREATE | `classifyMessages()`, `compoundPair`, `classifiedMessages` — extracted from gen_data_types.go |
| `gen_schema.go` | CREATE | `SchemaGenerator` — generic per-message codegen |
| `gen_schema_emitters.go` | CREATE | Emitter functions for SchemaGenerator (type defs, constructors, codecs) |
| `topo_sort.go` | CREATE | Topological sort for message dependency ordering |
| `topo_sort_test.go` | CREATE | Unit tests for topological sort |
| `gen_schema_test.go` | CREATE | Tests for SchemaGenerator including subscription golden test |
| `generate_test.go` | MODIFY | Inject proto options into existing synthetic and golden tests |
| `gen_table_schema.go` | DELETE | Absorbed into SchemaGenerator |
| `gen_table_files.go` | DELETE | Absorbed into SchemaGenerator (keep `snakeToPascal`, `deriveModuleName`, `isFieldFromDifferentFile` — move to `luau_naming.go`) |
| `gen_table_schema_test.go` | DELETE | Tests moved to gen_schema_test.go |
| `luau_naming.go` | MODIFY | Add `snakeToPascal`, `deriveModuleName`, `isFieldFromDifferentFile` (from gen_table_files.go); add repeated/map field type support |

---

### Task 1: Proto Options Package

**Files:**
- Create: `protoc-gen-luau/luauoptions/luau_options.proto`
- Create: `protoc-gen-luau/luauoptions/luau_options.pb.go`

**Depends on:** None — parallelizable

- [ ] **Step 1: Create luauoptions directory**

```bash
mkdir -p protoc-gen-luau/luauoptions
```

- [ ] **Step 2: Write luau_options.proto**

```proto
// protoc-gen-luau/luauoptions/luau_options.proto
syntax = "proto3";
package luau;

option go_package = "protoc-gen-luau/luauoptions";

import "google/protobuf/descriptor.proto";

extend google.protobuf.FileOptions {
  // Which generator strategy to use: "data_types" | "schema"
  string luau_generator = 51000;

  // Override the require path for data_types import.
  // Default: "@lib/data_types"
  string luau_data_types_require = 51001;
}
```

- [ ] **Step 3: Generate Go code from proto**

Run:
```bash
cd protoc-gen-luau && protoc \
  --go_out=. --go_opt=paths=source_relative \
  -I luauoptions \
  -I $(go env GOMODCACHE)/google.golang.org/protobuf@v1.36.11/types/known \
  -I /opt/homebrew/include \
  luauoptions/luau_options.proto
```

Expected: `luauoptions/luau_options.pb.go` is generated with `E_LuauGenerator` and `E_LuauDataTypesRequire` extension variables.

If protoc include paths differ on your system, locate `google/protobuf/descriptor.proto` and point `-I` there. Common locations:
- macOS Homebrew: `/opt/homebrew/include`
- Linux: `/usr/include`
- Go module cache: `$(go env GOMODCACHE)/google.golang.org/protobuf@v1.36.11/types/known`

- [ ] **Step 4: Verify generated code compiles**

Run: `cd protoc-gen-luau && go build ./luauoptions/`

Expected: No errors.

- [ ] **Step 5: Verify extension variables exist**

Check that `luauoptions/luau_options.pb.go` contains:
- `var E_LuauGenerator` — extension info for field 51000
- `var E_LuauDataTypesRequire` — extension info for field 51001

Run: `grep -n "E_Luau" protoc-gen-luau/luauoptions/luau_options.pb.go`

Expected: Two matches.

- [ ] **Step 6: Commit**

```bash
cd protoc-gen-luau && git add luauoptions/
git commit -m "feat: add luau_options.proto with file-level generator options"
```

---

### Task 2: Generator Interface, Registry, and Dispatch Refactor

**Files:**
- Create: `protoc-gen-luau/generator.go`
- Create: `protoc-gen-luau/classify.go`
- Modify: `protoc-gen-luau/generate.go`
- Modify: `protoc-gen-luau/gen_data_types.go`

**Depends on:** Task 1

- [ ] **Step 1: Write test for option-based dispatch**

Add to `protoc-gen-luau/generate_test.go` a new test that verifies option-based routing:

```go
func TestGenerateWithOptions(t *testing.T) {
	// Minimal request with luau_generator option set on one file
	opts := &descriptorpb.FileOptions{}
	proto.SetExtension(opts, luauoptions.E_LuauGenerator, "data_types")

	req := &pluginpb.CodeGeneratorRequest{
		FileToGenerate: []string{"data_types.proto"},
		ProtoFile: []*descriptorpb.FileDescriptorProto{
			{
				Name:    proto.String("data_types.proto"),
				Syntax:  proto.String("proto3"),
				Options: opts,
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

	// Should produce data_types.luau
	found := false
	for _, f := range resp.File {
		if f.GetName() == "data_types.luau" {
			found = true
		}
	}
	if !found {
		t.Error("expected data_types.luau in output")
	}
}

func TestGenerateUnknownOption(t *testing.T) {
	opts := &descriptorpb.FileOptions{}
	proto.SetExtension(opts, luauoptions.E_LuauGenerator, "unknown_gen")

	req := &pluginpb.CodeGeneratorRequest{
		FileToGenerate: []string{"foo.proto"},
		ProtoFile: []*descriptorpb.FileDescriptorProto{
			{
				Name:    proto.String("foo.proto"),
				Syntax:  proto.String("proto3"),
				Options: opts,
			},
		},
	}

	plugin, err := protogen.Options{}.New(req)
	if err != nil {
		t.Fatalf("protogen.New: %v", err)
	}
	err = generate(plugin)
	if err == nil {
		t.Fatal("expected error for unknown generator type")
	}
	if !strings.Contains(err.Error(), "unknown_gen") {
		t.Errorf("error should mention unknown type, got: %v", err)
	}
}

func TestGenerateNoOption(t *testing.T) {
	// File with no luau_generator option — should be skipped
	req := &pluginpb.CodeGeneratorRequest{
		FileToGenerate: []string{"ignored.proto"},
		ProtoFile: []*descriptorpb.FileDescriptorProto{
			{
				Name:    proto.String("ignored.proto"),
				Syntax:  proto.String("proto3"),
				Options: &descriptorpb.FileOptions{},
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
	if len(resp.File) != 0 {
		t.Errorf("expected no output files, got %d", len(resp.File))
	}
}
```

Add import for `luauoptions`: `"protoc-gen-luau/luauoptions"` to the test file imports.

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd protoc-gen-luau && go test -run TestGenerateWithOptions -v`

Expected: FAIL — `luauoptions` package imported but `generate()` doesn't read options yet.

- [ ] **Step 3: Extract classifyMessages to classify.go**

Move from `gen_data_types.go` to new file `protoc-gen-luau/classify.go`:
- `type compoundPair struct`
- `type classifiedMessages struct`
- `func classifyMessages(file *protogen.File) *classifiedMessages`

The function body is unchanged. Just move it and remove from `gen_data_types.go`.

```go
// protoc-gen-luau/classify.go
package main

import (
	"sort"
	"strings"

	"google.golang.org/protobuf/compiler/protogen"
)

type compoundPair struct {
	subFields *protogen.Message
	wrapper   *protogen.Message
}

type classifiedMessages struct {
	simpleTypes    []*protogen.Message
	compoundPairs  []compoundPair
	customCompound *protogen.Message
	oneofMsg       *protogen.Message
}

func classifyMessages(file *protogen.File) *classifiedMessages {
	cm := &classifiedMessages{}
	subFieldsMap := make(map[string]*protogen.Message)
	wrapperMap := make(map[string]*protogen.Message)

	for _, msg := range file.Messages {
		name := msg.GoIdent.GoName

		hasRealOneof := false
		for _, oneof := range msg.Oneofs {
			if !oneof.Desc.IsSynthetic() {
				hasRealOneof = true
				break
			}
		}
		if hasRealOneof {
			cm.oneofMsg = msg
			continue
		}

		hasMap := false
		for _, f := range msg.Fields {
			if f.Desc.IsMap() {
				hasMap = true
				break
			}
		}
		if hasMap {
			cm.customCompound = msg
			continue
		}

		if strings.HasSuffix(name, "SubFields") {
			prefix := strings.TrimSuffix(name, "SubFields")
			subFieldsMap[prefix] = msg
			continue
		}

		if strings.HasSuffix(name, "Fields") {
			prefix := strings.TrimSuffix(name, "Fields")
			subFieldsMap[prefix] = msg
			continue
		}

		hasValueSubFields := false
		for _, f := range msg.Fields {
			if string(f.Desc.Name()) == "value_sub_fields" {
				hasValueSubFields = true
				break
			}
		}
		if hasValueSubFields {
			prefix := strings.TrimSuffix(name, "Type")
			wrapperMap[prefix] = msg
			continue
		}

		cm.simpleTypes = append(cm.simpleTypes, msg)
	}

	for prefix, sf := range subFieldsMap {
		if w, ok := wrapperMap[prefix]; ok {
			cm.compoundPairs = append(cm.compoundPairs, compoundPair{
				subFields: sf,
				wrapper:   w,
			})
		}
	}
	sort.Slice(cm.compoundPairs, func(i, j int) bool {
		return cm.compoundPairs[i].wrapper.GoIdent.GoName < cm.compoundPairs[j].wrapper.GoIdent.GoName
	})

	return cm
}
```

Remove the same code from `gen_data_types.go` (lines 1-95: the imports for `sort` and `strings`, the struct definitions, and `classifyMessages`). Keep only `generateDataTypes` and its imports.

- [ ] **Step 4: Create generator.go with interface and registry**

```go
// protoc-gen-luau/generator.go
package main

import (
	"fmt"

	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"

	"protoc-gen-luau/luauoptions"
)

// LuauGenerator is the interface for code generation strategies.
type LuauGenerator interface {
	Generate(plugin *protogen.Plugin, file *protogen.File)
}

// generators maps luau_generator option values to implementations.
var generators = map[string]LuauGenerator{
	"data_types": &DataTypesGenerator{},
	"schema":     &SchemaGenerator{},
}

// getGeneratorType reads the luau_generator file option from a proto file.
// Returns "" if the option is not set.
func getGeneratorType(f *protogen.File) string {
	opts := f.Proto.GetOptions()
	if opts == nil {
		return ""
	}
	if !proto.HasExtension(opts, luauoptions.E_LuauGenerator) {
		return ""
	}
	return proto.GetExtension(opts, luauoptions.E_LuauGenerator).(string)
}

// getDataTypesRequirePath reads the luau_data_types_require file option.
// Returns "@lib/data_types" if the option is not set.
func getDataTypesRequirePath(f *protogen.File) string {
	opts := f.Proto.GetOptions()
	if opts == nil {
		return "@lib/data_types"
	}
	if !proto.HasExtension(opts, luauoptions.E_LuauDataTypesRequire) {
		return "@lib/data_types"
	}
	v := proto.GetExtension(opts, luauoptions.E_LuauDataTypesRequire).(string)
	if v == "" {
		return "@lib/data_types"
	}
	return v
}

// DataTypesGenerator generates the data_types.luau module.
type DataTypesGenerator struct{}

func (d *DataTypesGenerator) Generate(plugin *protogen.Plugin, file *protogen.File) {
	generateDataTypes(plugin, file)
}

// SchemaGenerator generates generic Luau modules from any proto schema.
type SchemaGenerator struct{}

func (s *SchemaGenerator) Generate(plugin *protogen.Plugin, file *protogen.File) {
	generateSchema(plugin, file)
}

// generateSchema is the stub — implemented in gen_schema.go.
// For now, emit empty module so tests can pass.
func generateSchemaStub(plugin *protogen.Plugin, file *protogen.File) {
	protoPath := file.Desc.Path()
	baseName := strings.TrimSuffix(path.Base(protoPath), ".proto")
	g := plugin.NewGeneratedFile(baseName+".luau", "")
	g.P("--!strict")
	g.P("-- Generated by protoc-gen-luau from ", protoPath)
	g.P("return {}")
}

// lookupGenerator returns the generator for the given type, or an error.
func lookupGenerator(genType string) (LuauGenerator, error) {
	gen, ok := generators[genType]
	if !ok {
		return nil, fmt.Errorf("unknown luau_generator %q", genType)
	}
	return gen, nil
}
```

Note: remove the `generateSchemaStub` function once `gen_schema.go` exists. Initially `SchemaGenerator.Generate` can call `generateSchemaStub` while the real implementation is built.

- [ ] **Step 5: Refactor generate.go to use registry dispatch**

Replace the entire contents of `protoc-gen-luau/generate.go`:

```go
package main

import (
	"fmt"

	"google.golang.org/protobuf/compiler/protogen"
)

func generate(plugin *protogen.Plugin) error {
	for _, f := range plugin.Files {
		if !f.Generate {
			continue
		}

		genType := getGeneratorType(f)
		if genType == "" {
			continue // No luau_generator option — skip
		}

		gen, err := lookupGenerator(genType)
		if err != nil {
			return fmt.Errorf("%s: %w", f.Desc.Path(), err)
		}
		gen.Generate(plugin, f)
	}
	return nil
}
```

- [ ] **Step 6: Create initial gen_schema.go stub**

```go
// protoc-gen-luau/gen_schema.go
package main

import (
	"path"
	"strings"

	"google.golang.org/protobuf/compiler/protogen"
)

func generateSchema(plugin *protogen.Plugin, file *protogen.File) {
	protoPath := file.Desc.Path()
	baseName := strings.TrimSuffix(path.Base(protoPath), ".proto")
	moduleName := deriveModuleName(protoPath)
	requirePath := getDataTypesRequirePath(file)

	g := plugin.NewGeneratedFile(baseName+".luau", "")
	g.P("--!strict")
	g.P(`local DataTypes = require("`, requirePath, `")`)
	g.P()
	g.P("local ", moduleName, " = {}")
	g.P()

	// TODO: full implementation in subsequent tasks

	g.P("return ", moduleName)
}
```

- [ ] **Step 7: Move utility functions to luau_naming.go**

Move `snakeToPascal`, `deriveModuleName`, and `isFieldFromDifferentFile` from `gen_table_files.go` into `luau_naming.go`. These are shared utilities both generators need. Add the required imports (`path`, `strings`, `protoreflect`).

In `luau_naming.go`, add after the existing functions:

```go
// snakeToPascal converts "source_table" to "SourceTable".
func snakeToPascal(s string) string {
	parts := strings.Split(s, "_")
	for i := range parts {
		if len(parts[i]) > 0 {
			runes := []rune(parts[i])
			runes[0] = rune(strings.ToUpper(string(runes[0]))[0])
			parts[i] = string(runes)
		}
	}
	return strings.Join(parts, "")
}

// deriveModuleName extracts the base name from a proto path and converts to PascalCase.
func deriveModuleName(protoPath string) string {
	base := strings.TrimSuffix(path.Base(protoPath), ".proto")
	return snakeToPascal(base)
}

// isFieldFromDifferentFile returns true if the field's message type is defined in a
// different proto file than currentFile.
func isFieldFromDifferentFile(field *protogen.Field, currentFile *protogen.File) bool {
	if field.Desc.Kind() != protoreflect.MessageKind {
		return false
	}
	return field.Message.Desc.ParentFile().Path() != currentFile.Desc.Path()
}
```

Add `"path"` to the imports of `luau_naming.go`. Remove these three functions from `gen_table_files.go`.

- [ ] **Step 8: Remove the generateSchemaStub function from generator.go**

Now that `gen_schema.go` provides `generateSchema()`, remove the `generateSchemaStub` function from `generator.go` and ensure `SchemaGenerator.Generate` calls `generateSchema`. Also remove unused imports (`path`, `strings`) from generator.go.

- [ ] **Step 9: Run new tests**

Run: `cd protoc-gen-luau && go test -run "TestGenerateWithOptions|TestGenerateUnknownOption|TestGenerateNoOption" -v`

Expected: All 3 tests PASS.

- [ ] **Step 10: Update existing TestGenerate to use options**

The existing `TestGenerate` creates synthetic requests without options. Update it to inject options:

```go
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
			"data_types.proto",
			"table_schema.proto",
			"tables/source_table.proto",
			"tables/target_table.proto",
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

	wantNames := map[string]bool{
		"data_types.luau":   true,
		"table_schema.luau": true,
		"source_table.luau": true,
		"target_table.luau": true,
	}

	if len(resp.File) != len(wantNames) {
		t.Errorf("got %d files, want %d", len(resp.File), len(wantNames))
	}
	for _, f := range resp.File {
		if !wantNames[f.GetName()] {
			t.Errorf("unexpected file %q", f.GetName())
		}
	}
}
```

- [ ] **Step 11: Update TestGenerateGolden to inject options**

After unmarshaling request.pb, inject `luau_generator` options into each file descriptor before creating the plugin:

```go
func TestGenerateGolden(t *testing.T) {
	data, err := os.ReadFile("testdata/request.pb")
	if err != nil {
		t.Fatalf("reading request.pb: %v", err)
	}

	req := &pluginpb.CodeGeneratorRequest{}
	if err := proto.Unmarshal(data, req); err != nil {
		t.Fatalf("unmarshaling request: %v", err)
	}

	// Inject luau_generator options into each proto file
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
	// ... rest unchanged ...
```

Add `"path"` to the imports.

- [ ] **Step 12: Run all tests**

Run: `cd protoc-gen-luau && go test ./... -v`

Expected: `TestGenerate` PASS, `TestGenerateWithOptions` PASS, `TestGenerateUnknownOption` PASS, `TestGenerateNoOption` PASS. `TestGenerateGolden` will FAIL (SchemaGenerator is a stub — expected). Other tests that call `generate()` may fail similarly — that's expected until SchemaGenerator is implemented.

- [ ] **Step 13: Commit**

```bash
cd protoc-gen-luau
git add generator.go classify.go generate.go gen_data_types.go luau_naming.go generate_test.go
git commit -m "feat: add Generator interface, registry dispatch, extract classifyMessages"
```

---

### Task 3: Topological Sort

**Files:**
- Create: `protoc-gen-luau/topo_sort.go`
- Create: `protoc-gen-luau/topo_sort_test.go`

**Depends on:** None — parallelizable with Task 2

- [ ] **Step 1: Write failing tests for topological sort**

```go
// protoc-gen-luau/topo_sort_test.go
package main

import (
	"testing"

	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/pluginpb"
)

// makeTestMessages creates a protogen.File with messages that have specified dependencies.
// deps maps message name → list of message names it depends on (via fields).
func makeTestMessages(t *testing.T, msgNames []string, deps map[string][]string) []*protogen.Message {
	t.Helper()

	// Build FileDescriptorProto with messages
	fdp := &descriptorpb.FileDescriptorProto{
		Name:   proto.String("test.proto"),
		Syntax: proto.String("proto3"),
	}

	// Add all messages
	for i, name := range msgNames {
		md := &descriptorpb.DescriptorProto{Name: proto.String(name)}
		// Add a string field so the message isn't empty
		md.Field = append(md.Field, &descriptorpb.FieldDescriptorProto{
			Name:     proto.String("id"),
			Number:   proto.Int32(1),
			Type:     descriptorpb.FieldDescriptorProto_TYPE_STRING.Enum(),
			JsonName: proto.String("id"),
		})
		// Add message-type fields for dependencies
		if fieldDeps, ok := deps[name]; ok {
			for j, dep := range fieldDeps {
				fieldNum := int32(10 + i*10 + j)
				md.Field = append(md.Field, &descriptorpb.FieldDescriptorProto{
					Name:     proto.String(snakeToCamel(dep)),
					Number:   &fieldNum,
					Type:     descriptorpb.FieldDescriptorProto_TYPE_MESSAGE.Enum(),
					TypeName: proto.String("." + dep),
					JsonName: proto.String(snakeToCamel(dep)),
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
	var file *protogen.File
	for _, f := range plugin.Files {
		if f.GeneratedFilenamePrefix == "test" {
			file = f
			break
		}
	}
	if file == nil {
		t.Fatal("test.proto not found")
	}
	return file.Messages
}

func msgNames(msgs []*protogen.Message) []string {
	names := make([]string, len(msgs))
	for i, m := range msgs {
		names[i] = m.GoIdent.GoName
	}
	return names
}

func TestTopoSortNoDeps(t *testing.T) {
	msgs := makeTestMessages(t, []string{"A", "B", "C"}, nil)
	sorted, err := topoSortMessages(msgs, msgs[0].Desc.ParentFile())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(sorted) != 3 {
		t.Errorf("expected 3 messages, got %d", len(sorted))
	}
}

func TestTopoSortLinearChain(t *testing.T) {
	// C depends on B, B depends on A → output order: A, B, C
	msgs := makeTestMessages(t, []string{"C", "B", "A"}, map[string][]string{
		"C": {"B"},
		"B": {"A"},
	})
	sorted, err := topoSortMessages(msgs, msgs[0].Desc.ParentFile())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	names := msgNames(sorted)
	// A must come before B, B must come before C
	aIdx, bIdx, cIdx := -1, -1, -1
	for i, n := range names {
		switch n {
		case "A":
			aIdx = i
		case "B":
			bIdx = i
		case "C":
			cIdx = i
		}
	}
	if aIdx > bIdx || bIdx > cIdx {
		t.Errorf("wrong order: %v (expected A before B before C)", names)
	}
}

func TestTopoSortDiamond(t *testing.T) {
	// D depends on B and C; B and C both depend on A
	msgs := makeTestMessages(t, []string{"D", "B", "C", "A"}, map[string][]string{
		"D": {"B", "C"},
		"B": {"A"},
		"C": {"A"},
	})
	sorted, err := topoSortMessages(msgs, msgs[0].Desc.ParentFile())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	names := msgNames(sorted)
	// A must come before B and C; B and C must come before D
	idxMap := map[string]int{}
	for i, n := range names {
		idxMap[n] = i
	}
	if idxMap["A"] > idxMap["B"] || idxMap["A"] > idxMap["C"] {
		t.Errorf("A should come before B and C: %v", names)
	}
	if idxMap["B"] > idxMap["D"] || idxMap["C"] > idxMap["D"] {
		t.Errorf("B and C should come before D: %v", names)
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd protoc-gen-luau && go test -run TestTopoSort -v`

Expected: FAIL — `topoSortMessages` undefined.

- [ ] **Step 3: Implement topological sort**

```go
// protoc-gen-luau/topo_sort.go
package main

import (
	"fmt"

	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// topoSortMessages sorts messages in dependency order using Kahn's algorithm.
// A message depends on another if it has a field whose type is a message
// defined in the same file. Returns error on circular dependency.
func topoSortMessages(msgs []*protogen.Message, fileDesc protoreflect.FileDescriptor) ([]*protogen.Message, error) {
	filePath := fileDesc.Path()

	// Build name → message index
	nameIdx := make(map[string]int, len(msgs))
	for i, m := range msgs {
		nameIdx[m.GoIdent.GoName] = i
	}

	// Build adjacency: inDegree and dependents
	inDegree := make([]int, len(msgs))
	// dependents[i] = list of message indices that depend on msgs[i]
	dependents := make([][]int, len(msgs))

	for i, m := range msgs {
		for _, field := range m.Fields {
			if field.Desc.Kind() != protoreflect.MessageKind {
				continue
			}
			// Check if the field's message type is in the same file
			if field.Message.Desc.ParentFile().Path() != filePath {
				continue
			}
			depName := field.Message.GoIdent.GoName
			if j, ok := nameIdx[depName]; ok && j != i {
				dependents[j] = append(dependents[j], i)
				inDegree[i]++
			}
		}
		// Also check nested message references in repeated fields
		// (repeated fields still have Kind() == MessageKind for message elements)
	}

	// Kahn's algorithm
	var queue []int
	for i, deg := range inDegree {
		if deg == 0 {
			queue = append(queue, i)
		}
	}

	var sorted []*protogen.Message
	for len(queue) > 0 {
		curr := queue[0]
		queue = queue[1:]
		sorted = append(sorted, msgs[curr])

		for _, dep := range dependents[curr] {
			inDegree[dep]--
			if inDegree[dep] == 0 {
				queue = append(queue, dep)
			}
		}
	}

	if len(sorted) != len(msgs) {
		return nil, fmt.Errorf("circular dependency detected among messages in %s", filePath)
	}

	return sorted, nil
}
```

- [ ] **Step 4: Run tests**

Run: `cd protoc-gen-luau && go test -run TestTopoSort -v`

Expected: All 3 tests PASS.

- [ ] **Step 5: Commit**

```bash
cd protoc-gen-luau && git add topo_sort.go topo_sort_test.go
git commit -m "feat: add topological sort for message dependency ordering"
```

---

### Task 4: SchemaGenerator — Generic Type Definitions and Constructors

**Files:**
- Create: `protoc-gen-luau/gen_schema_emitters.go`
- Modify: `protoc-gen-luau/gen_schema.go`
- Modify: `protoc-gen-luau/luau_naming.go`
- Create: `protoc-gen-luau/gen_schema_test.go`

**Depends on:** Task 2, Task 3

This task implements the SchemaGenerator's type definition and constructor emission for all field kinds. Codecs come in Task 5.

- [ ] **Step 1: Extend luau_naming.go with repeated and map field support**

Add to `luau_naming.go`:

```go
// luauTypeRefGeneric returns the Luau type string for a field in a generic schema context.
// Handles scalar, message (local vs imported), repeated, and map fields.
// Uses proto3 optional semantics: message fields without `optional` are still treated
// as potentially nil in proto3 (messages are pointer-like).
func luauTypeRefGeneric(field *protogen.Field, currentFile *protogen.File) string {
	if field.Desc.IsMap() {
		// Map<K, V> — K is always string in our use case
		valField := field.Message.Fields[1] // map value is the second field of the entry message
		valType := luauTypeRefGeneric(valField, currentFile)
		return "{[string]: " + valType + "}"
	}
	if field.Desc.IsList() {
		elemType := luauElementType(field, currentFile)
		return "{" + elemType + "}"
	}
	if field.Desc.Kind() == protoreflect.MessageKind {
		if field.Message.Desc.ParentFile().Path() == currentFile.Desc.Path() {
			return field.Message.GoIdent.GoName
		}
		// For now, assume all imports are from data_types
		return "DataTypes." + field.Message.GoIdent.GoName
	}
	return scalarLuauType(field)
}

// luauElementType returns the Luau type for the element of a repeated field.
func luauElementType(field *protogen.Field, currentFile *protogen.File) string {
	switch field.Desc.Kind() {
	case protoreflect.StringKind:
		return "string"
	case protoreflect.DoubleKind, protoreflect.Int32Kind, protoreflect.FloatKind, protoreflect.Int64Kind:
		return "number"
	case protoreflect.BoolKind:
		return "boolean"
	case protoreflect.MessageKind:
		if field.Message.Desc.ParentFile().Path() == currentFile.Desc.Path() {
			return field.Message.GoIdent.GoName
		}
		return "DataTypes." + field.Message.GoIdent.GoName
	default:
		return "any"
	}
}

// scalarLuauType returns the Luau type for a scalar proto field.
func scalarLuauType(field *protogen.Field) string {
	switch field.Desc.Kind() {
	case protoreflect.StringKind:
		return "string"
	case protoreflect.DoubleKind, protoreflect.Int32Kind, protoreflect.FloatKind, protoreflect.Int64Kind:
		return "number"
	case protoreflect.BoolKind:
		return "boolean"
	default:
		return "any"
	}
}

// isFieldOptional returns true if a field should be optional (? suffix) in Luau.
// Rules:
// - Fields with proto3 `optional` keyword: optional
// - Message-kind fields (without `optional`): optional (proto3 messages are nullable)
// - Repeated fields: NOT optional (always present, default {})
// - Map fields: NOT optional (always present, default {})
// - Scalar fields without `optional`: required (not optional)
func isFieldOptional(field *protogen.Field) bool {
	if field.Desc.IsList() || field.Desc.IsMap() {
		return false
	}
	if field.Desc.HasOptionalKeyword() {
		return true
	}
	if field.Desc.Kind() == protoreflect.MessageKind {
		return true // proto3 message fields are always nullable
	}
	return false
}

// luauDefaultValueGeneric returns the Luau default for a field in a constructor.
// Returns "" for fields with no default (nil/required).
func luauDefaultValueGeneric(field *protogen.Field) string {
	if field.Desc.IsList() {
		return "{}"
	}
	if field.Desc.IsMap() {
		return "{}"
	}
	if field.Desc.Kind() == protoreflect.MessageKind {
		return "" // nil
	}
	if field.Desc.HasOptionalKeyword() {
		return "" // nil
	}
	// Required scalar — provide zero-value default
	switch field.Desc.Kind() {
	case protoreflect.StringKind:
		return `""`
	case protoreflect.DoubleKind, protoreflect.Int32Kind, protoreflect.FloatKind, protoreflect.Int64Kind:
		return "0"
	case protoreflect.BoolKind:
		return "" // special: use if ~= nil pattern
	default:
		return `""`
	}
}
```

- [ ] **Step 2: Write test for SchemaGenerator type definitions**

```go
// protoc-gen-luau/gen_schema_test.go
package main

import (
	"strings"
	"testing"

	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/pluginpb"

	"protoc-gen-luau/luauoptions"
)

// buildSchemaRequest creates a CodeGeneratorRequest for testing SchemaGenerator.
// Returns the proto file basename and the request.
func buildSchemaRequest(t *testing.T, schemaFile *descriptorpb.FileDescriptorProto) *pluginpb.CodeGeneratorRequest {
	t.Helper()

	// data_types.proto with basic messages for cross-file references
	dtOpts := &descriptorpb.FileOptions{}
	proto.SetExtension(dtOpts, luauoptions.E_LuauGenerator, "data_types")

	dtFile := &descriptorpb.FileDescriptorProto{
		Name:    proto.String("data_types.proto"),
		Syntax:  proto.String("proto3"),
		Options: dtOpts,
		MessageType: []*descriptorpb.DescriptorProto{
			{
				Name: proto.String("StringType"),
				Field: []*descriptorpb.FieldDescriptorProto{
					{Name: proto.String("type_id"), Number: proto.Int32(1), Type: descriptorpb.FieldDescriptorProto_TYPE_STRING.Enum(), JsonName: proto.String("typeId")},
					{Name: proto.String("value"), Number: proto.Int32(2), Type: descriptorpb.FieldDescriptorProto_TYPE_STRING.Enum(), JsonName: proto.String("value")},
				},
			},
			{
				Name: proto.String("NumberType"),
				Field: []*descriptorpb.FieldDescriptorProto{
					{Name: proto.String("type_id"), Number: proto.Int32(1), Type: descriptorpb.FieldDescriptorProto_TYPE_STRING.Enum(), JsonName: proto.String("typeId")},
					{Name: proto.String("value"), Number: proto.Int32(2), Type: descriptorpb.FieldDescriptorProto_TYPE_DOUBLE.Enum(), JsonName: proto.String("value")},
				},
			},
		},
	}

	return &pluginpb.CodeGeneratorRequest{
		FileToGenerate: []string{schemaFile.GetName()},
		ProtoFile:      []*descriptorpb.FileDescriptorProto{dtFile, schemaFile},
	}
}

func getSchemaOutput(t *testing.T, req *pluginpb.CodeGeneratorRequest, fileName string) string {
	t.Helper()
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
		if f.GetName() == fileName {
			return f.GetContent()
		}
	}
	t.Fatalf("file %s not found in response", fileName)
	return ""
}

func TestSchemaTypeDefsSimple(t *testing.T) {
	schemaOpts := &descriptorpb.FileOptions{}
	proto.SetExtension(schemaOpts, luauoptions.E_LuauGenerator, "schema")

	schemaFile := &descriptorpb.FileDescriptorProto{
		Name:       proto.String("simple_schema.proto"),
		Syntax:     proto.String("proto3"),
		Dependency: []string{"data_types.proto"},
		Options:    schemaOpts,
		MessageType: []*descriptorpb.DescriptorProto{
			{
				Name: proto.String("SimpleRecord"),
				Field: []*descriptorpb.FieldDescriptorProto{
					{Name: proto.String("type_id"), Number: proto.Int32(1),
						Type: descriptorpb.FieldDescriptorProto_TYPE_STRING.Enum(), JsonName: proto.String("typeId")},
					{Name: proto.String("label"), Number: proto.Int32(2),
						Type: descriptorpb.FieldDescriptorProto_TYPE_STRING.Enum(), JsonName: proto.String("label")},
					{Name: proto.String("count"), Number: proto.Int32(3),
						Type: descriptorpb.FieldDescriptorProto_TYPE_INT32.Enum(), JsonName: proto.String("count")},
					{Name: proto.String("active"), Number: proto.Int32(4),
						Type: descriptorpb.FieldDescriptorProto_TYPE_BOOL.Enum(), JsonName: proto.String("active")},
					{Name: proto.String("nickname"), Number: proto.Int32(5),
						Type: descriptorpb.FieldDescriptorProto_TYPE_STRING.Enum(), JsonName: proto.String("nickname"),
						Proto3Optional: proto.Bool(true)},
				},
			},
		},
	}

	req := buildSchemaRequest(t, schemaFile)
	output := getSchemaOutput(t, req, "simple_schema.luau")

	// Header
	if !strings.Contains(output, "--!strict") {
		t.Error("missing --!strict")
	}
	if !strings.Contains(output, `local DataTypes = require("@lib/data_types")`) {
		t.Error("missing DataTypes require")
	}
	if !strings.Contains(output, "local SimpleSchema = {}") {
		t.Error("missing module declaration")
	}

	// Type definition
	if !strings.Contains(output, "export type SimpleRecord = {") {
		t.Error("missing SimpleRecord type def")
	}
	// Required fields (no ?)
	if !strings.Contains(output, "\ttypeId: string,") {
		t.Error("typeId should be required string")
	}
	if !strings.Contains(output, "\tlabel: string,") {
		t.Error("label should be required string")
	}
	if !strings.Contains(output, "\tcount: number,") {
		t.Error("count should be required number")
	}
	if !strings.Contains(output, "\tactive: boolean,") {
		t.Error("active should be required boolean")
	}
	// Optional field (has ?)
	if !strings.Contains(output, "\tnickname: string?,") {
		t.Error("nickname should be optional string")
	}

	// Constructor
	if !strings.Contains(output, "function SimpleSchema.makeSimpleRecord(") {
		t.Error("missing constructor")
	}

	// Footer
	if !strings.HasSuffix(strings.TrimSpace(output), "return SimpleSchema") {
		t.Error("missing return footer")
	}
}
```

- [ ] **Step 3: Run test to verify it fails**

Run: `cd protoc-gen-luau && go test -run TestSchemaTypeDefsSimple -v`

Expected: FAIL — SchemaGenerator only emits stub.

- [ ] **Step 4: Implement gen_schema_emitters.go — type definitions and constructors**

```go
// protoc-gen-luau/gen_schema_emitters.go
package main

import (
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// emitSchemaTypeDef emits an `export type` for a message in a generic schema.
func emitSchemaTypeDef(g *protogen.GeneratedFile, msg *protogen.Message, currentFile *protogen.File) {
	name := msg.GoIdent.GoName
	g.P("export type ", name, " = {")
	for _, field := range msg.Fields {
		luauName := snakeToCamel(string(field.Desc.Name()))
		luauType := luauTypeRefGeneric(field, currentFile)
		if isFieldOptional(field) {
			g.P("\t", luauName, ": ", luauType, "?,")
		} else {
			g.P("\t", luauName, ": ", luauType, ",")
		}
	}
	g.P("}")
	g.P()
}

// emitSchemaConstructor emits a make function for a message in a generic schema.
func emitSchemaConstructor(g *protogen.GeneratedFile, msg *protogen.Message, moduleName string, currentFile *protogen.File) {
	name := msg.GoIdent.GoName
	g.P("function ", moduleName, ".make", name, "(config: {")
	for _, field := range msg.Fields {
		luauName := snakeToCamel(string(field.Desc.Name()))
		luauType := luauTypeRefGeneric(field, currentFile)
		// All constructor params are optional with ?
		g.P("\t", luauName, ": ", luauType, "?,")
	}
	g.P("}?): ", name)
	g.P("\tlocal c = config or {}")
	g.P("\treturn {")
	for _, field := range msg.Fields {
		luauName := snakeToCamel(string(field.Desc.Name()))
		dflt := luauDefaultValueGeneric(field)
		if field.Desc.Kind() == protoreflect.BoolKind && !field.Desc.HasOptionalKeyword() {
			g.P("\t\t", luauName, " = if c.", luauName, " ~= nil then c.", luauName, " else false,")
		} else if dflt != "" {
			g.P("\t\t", luauName, " = c.", luauName, " or ", dflt, ",")
		} else {
			g.P("\t\t", luauName, " = c.", luauName, ",")
		}
	}
	g.P("\t}")
	g.P("end")
	g.P()
}
```

- [ ] **Step 5: Update gen_schema.go to use emitters**

Replace the stub in `gen_schema.go`:

```go
package main

import (
	"path"
	"strings"

	"google.golang.org/protobuf/compiler/protogen"
)

func generateSchema(plugin *protogen.Plugin, file *protogen.File) {
	protoPath := file.Desc.Path()
	baseName := strings.TrimSuffix(path.Base(protoPath), ".proto")
	moduleName := deriveModuleName(protoPath)
	requirePath := getDataTypesRequirePath(file)

	g := plugin.NewGeneratedFile(baseName+".luau", "")

	// Header
	g.P("--!strict")
	g.P(`local DataTypes = require("`, requirePath, `")`)
	g.P()
	g.P("local ", moduleName, " = {}")
	g.P()

	// Classify and sort messages
	cm := classifyMessages(file)

	// Collect all non-compound messages for generic codegen
	var allMessages []*protogen.Message
	allMessages = append(allMessages, cm.simpleTypes...)
	if cm.oneofMsg != nil {
		allMessages = append(allMessages, cm.oneofMsg)
	}

	// Topological sort
	sorted, err := topoSortMessages(allMessages, file.Desc)
	if err != nil {
		// Fall back to original order on error
		sorted = allMessages
	}

	// Compound pair type definitions (before sorted messages, since sorted messages may reference them)
	for _, pair := range cm.compoundPairs {
		sfName := pair.subFields.GoIdent.GoName
		emitCompoundSubFieldsDef(g, pair.subFields, file)
		g.P()
		emitCompoundWrapperDef(g, pair.wrapper, sfName)
		g.P()
	}

	// Type definitions for sorted messages
	for _, msg := range sorted {
		emitSchemaTypeDef(g, msg, file)
	}

	// Compound pair constructors
	for _, pair := range cm.compoundPairs {
		sfName := pair.subFields.GoIdent.GoName
		emitCompoundSubFieldsConstructor(g, pair.subFields, moduleName, file)
		emitCompoundWrapperConstructor(g, pair.wrapper, sfName, moduleName)
	}

	// Constructors for sorted messages
	for _, msg := range sorted {
		emitSchemaConstructor(g, msg, moduleName, file)
	}

	// Per-module optional helpers for DataTypes-imported types
	emitSchemaOptionalHelpers(g, sorted, file)

	// Compound pair codecs
	for _, pair := range cm.compoundPairs {
		sfName := pair.subFields.GoIdent.GoName
		emitCompoundSubFieldsEncodeWith(g, pair.subFields, moduleName, tableFileHelperName)
		emitCompoundSubFieldsDecodeWith(g, pair.subFields, moduleName, tableFileHelperName)
		emitCompoundWrapperEncode(g, pair.wrapper, sfName, moduleName)
		emitCompoundWrapperDecode(g, pair.wrapper, sfName, moduleName)
	}

	// Codecs for sorted messages
	for _, msg := range sorted {
		emitSchemaEncode(g, msg, moduleName, file)
		emitSchemaDecode(g, msg, moduleName, file)
	}

	// Footer
	g.P("return ", moduleName)
}

// emitSchemaOptionalHelpers scans all messages for DataTypes-imported types
// and emits decode/encode optional helpers for each unique type.
func emitSchemaOptionalHelpers(g *protogen.GeneratedFile, msgs []*protogen.Message, currentFile *protogen.File) {
	typeSet := make(map[string]bool)
	for _, msg := range msgs {
		for _, field := range msg.Fields {
			if isFieldFromDifferentFile(field, currentFile) && isFieldOptional(field) {
				typeSet[field.Message.GoIdent.GoName] = true
			}
		}
	}

	typeNames := sortedKeys(typeSet)
	for _, typeName := range typeNames {
		g.P("local function decodeOptional", typeName, "(json: any): DataTypes.", typeName, "?")
		g.P("\tif json == nil then return nil end")
		g.P("\treturn DataTypes.decodeJson", typeName, "(json)")
		g.P("end")
		g.P("local function encodeOptional", typeName, "(v: DataTypes.", typeName, "?): {[string]: any}?")
		g.P("\tif v == nil then return nil end")
		g.P("\treturn DataTypes.encodeJson", typeName, "(v)")
		g.P("end")
	}
	if len(typeNames) > 0 {
		g.P()
	}
}

// sortedKeys returns the keys of a map[string]bool sorted alphabetically.
func sortedKeys(m map[string]bool) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
```

Add `"sort"` to imports.

- [ ] **Step 6: Add stub codec emitters** (to be implemented in Task 5)

Add to `gen_schema_emitters.go`:

```go
// emitSchemaEncode emits the JSON encoder for a message. Stub — implemented in Task 5.
func emitSchemaEncode(g *protogen.GeneratedFile, msg *protogen.Message, moduleName string, currentFile *protogen.File) {
	name := msg.GoIdent.GoName
	g.P("function ", moduleName, ".encodeJson", name, "(v: ", name, "): {[string]: any}")
	g.P("\treturn {} -- TODO: implement in Task 5")
	g.P("end")
	g.P()
}

// emitSchemaDecode emits the JSON decoder for a message. Stub — implemented in Task 5.
func emitSchemaDecode(g *protogen.GeneratedFile, msg *protogen.Message, moduleName string, currentFile *protogen.File) {
	name := msg.GoIdent.GoName
	g.P("function ", moduleName, ".decodeJson", name, "(json: {[string]: any}): ", name)
	g.P("\treturn {} :: any -- TODO: implement in Task 5")
	g.P("end")
	g.P()
}
```

- [ ] **Step 7: Run tests**

Run: `cd protoc-gen-luau && go test -run TestSchemaTypeDefsSimple -v`

Expected: PASS — type defs, constructor, header, footer all generated correctly.

- [ ] **Step 8: Add test for DataTypes reference and local cross-reference fields**

Add to `gen_schema_test.go`:

```go
func TestSchemaTypeDefsWithRefs(t *testing.T) {
	schemaOpts := &descriptorpb.FileOptions{}
	proto.SetExtension(schemaOpts, luauoptions.E_LuauGenerator, "schema")

	schemaFile := &descriptorpb.FileDescriptorProto{
		Name:       proto.String("ref_schema.proto"),
		Syntax:     proto.String("proto3"),
		Dependency: []string{"data_types.proto"},
		Options:    schemaOpts,
		MessageType: []*descriptorpb.DescriptorProto{
			{
				Name: proto.String("InnerRecord"),
				Field: []*descriptorpb.FieldDescriptorProto{
					{Name: proto.String("name"), Number: proto.Int32(1),
						Type: descriptorpb.FieldDescriptorProto_TYPE_STRING.Enum(), JsonName: proto.String("name")},
				},
			},
			{
				Name: proto.String("OuterRecord"),
				Field: []*descriptorpb.FieldDescriptorProto{
					{Name: proto.String("inner"), Number: proto.Int32(1),
						Type: descriptorpb.FieldDescriptorProto_TYPE_MESSAGE.Enum(),
						TypeName: proto.String(".InnerRecord"), JsonName: proto.String("inner"),
						Proto3Optional: proto.Bool(true)},
					{Name: proto.String("some_type"), Number: proto.Int32(2),
						Type: descriptorpb.FieldDescriptorProto_TYPE_MESSAGE.Enum(),
						TypeName: proto.String(".StringType"), JsonName: proto.String("someType"),
						Proto3Optional: proto.Bool(true)},
					{Name: proto.String("tags"), Number: proto.Int32(3),
						Type: descriptorpb.FieldDescriptorProto_TYPE_STRING.Enum(),
						Label: descriptorpb.FieldDescriptorProto_LABEL_REPEATED.Enum(),
						JsonName: proto.String("tags")},
				},
			},
		},
	}

	req := buildSchemaRequest(t, schemaFile)
	output := getSchemaOutput(t, req, "ref_schema.luau")

	// InnerRecord should come before OuterRecord (dependency order)
	innerIdx := strings.Index(output, "export type InnerRecord")
	outerIdx := strings.Index(output, "export type OuterRecord")
	if innerIdx < 0 || outerIdx < 0 {
		t.Fatal("missing type definitions")
	}
	if innerIdx > outerIdx {
		t.Error("InnerRecord should be defined before OuterRecord (dependency order)")
	}

	// Local message ref — optional
	if !strings.Contains(output, "\tinner: InnerRecord?,") {
		t.Error("inner should be local InnerRecord?")
	}

	// DataTypes ref — optional, prefixed
	if !strings.Contains(output, "\tsomeType: DataTypes.StringType?,") {
		t.Error("someType should be DataTypes.StringType?")
	}

	// Repeated scalar
	if !strings.Contains(output, "\ttags: {string},") {
		t.Error("tags should be {string} (repeated, no ?)")
	}
}
```

- [ ] **Step 9: Run tests**

Run: `cd protoc-gen-luau && go test -run "TestSchemaTypeDefs" -v`

Expected: Both tests PASS.

- [ ] **Step 10: Commit**

```bash
cd protoc-gen-luau && git add gen_schema.go gen_schema_emitters.go gen_schema_test.go luau_naming.go
git commit -m "feat: SchemaGenerator type definitions and constructors for all field kinds"
```

---

### Task 5: SchemaGenerator — Generic Codec Emission

**Files:**
- Modify: `protoc-gen-luau/gen_schema_emitters.go`
- Modify: `protoc-gen-luau/gen_schema_test.go`

**Depends on:** Task 4

- [ ] **Step 1: Write codec test**

Add to `gen_schema_test.go`:

```go
func TestSchemaCodecsSimple(t *testing.T) {
	schemaOpts := &descriptorpb.FileOptions{}
	proto.SetExtension(schemaOpts, luauoptions.E_LuauGenerator, "schema")

	schemaFile := &descriptorpb.FileDescriptorProto{
		Name:       proto.String("codec_test.proto"),
		Syntax:     proto.String("proto3"),
		Dependency: []string{"data_types.proto"},
		Options:    schemaOpts,
		MessageType: []*descriptorpb.DescriptorProto{
			{
				Name: proto.String("CodecRecord"),
				Field: []*descriptorpb.FieldDescriptorProto{
					// Required string
					{Name: proto.String("type_id"), Number: proto.Int32(1),
						Type: descriptorpb.FieldDescriptorProto_TYPE_STRING.Enum(), JsonName: proto.String("typeId")},
					// Optional string
					{Name: proto.String("nickname"), Number: proto.Int32(2),
						Type: descriptorpb.FieldDescriptorProto_TYPE_STRING.Enum(), JsonName: proto.String("nickname"),
						Proto3Optional: proto.Bool(true)},
					// Optional DataTypes ref
					{Name: proto.String("value"), Number: proto.Int32(3),
						Type: descriptorpb.FieldDescriptorProto_TYPE_MESSAGE.Enum(),
						TypeName: proto.String(".StringType"), JsonName: proto.String("value"),
						Proto3Optional: proto.Bool(true)},
					// Repeated string
					{Name: proto.String("tags"), Number: proto.Int32(4),
						Type: descriptorpb.FieldDescriptorProto_TYPE_STRING.Enum(),
						Label: descriptorpb.FieldDescriptorProto_LABEL_REPEATED.Enum(),
						JsonName: proto.String("tags")},
				},
			},
		},
	}

	req := buildSchemaRequest(t, schemaFile)
	output := getSchemaOutput(t, req, "codec_test.luau")

	// Encode function exists
	if !strings.Contains(output, "function CodecTest.encodeJsonCodecRecord(") {
		t.Error("missing encodeJsonCodecRecord")
	}
	// Required string in encode: direct
	if !strings.Contains(output, "typeId = v.typeId") {
		t.Error("encode should include typeId = v.typeId")
	}
	// Optional string in encode: nil check
	if !strings.Contains(output, "v.nickname") {
		t.Error("encode should handle nickname")
	}
	// Optional DataTypes in encode: helper
	if !strings.Contains(output, "encodeOptionalStringType") {
		t.Error("encode should use encodeOptionalStringType helper")
	}
	// Repeated in encode: direct
	if !strings.Contains(output, "tags = v.tags") {
		t.Error("encode should include tags = v.tags")
	}

	// Decode function exists
	if !strings.Contains(output, "function CodecTest.decodeJsonCodecRecord(") {
		t.Error("missing decodeJsonCodecRecord")
	}
	// Required string in decode: with default
	if !strings.Contains(output, `typeId = json.typeId or ""`) {
		t.Error("decode should default typeId to empty string")
	}
	// Optional string in decode: no default
	if !strings.Contains(output, "nickname = json.nickname,") {
		t.Error("decode should pass nickname through (optional, nil OK)")
	}
	// Optional DataTypes in decode: helper
	if !strings.Contains(output, "decodeOptionalStringType") {
		t.Error("decode should use decodeOptionalStringType helper")
	}
	// Repeated in decode: with default
	if !strings.Contains(output, "tags = json.tags or {}") {
		t.Error("decode should default tags to {}")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd protoc-gen-luau && go test -run TestSchemaCodecsSimple -v`

Expected: FAIL — codec stubs return empty.

- [ ] **Step 3: Implement full codec emitters**

Replace the stub `emitSchemaEncode` and `emitSchemaDecode` in `gen_schema_emitters.go`:

```go
// emitSchemaEncode emits the JSON encoder for a message.
func emitSchemaEncode(g *protogen.GeneratedFile, msg *protogen.Message, moduleName string, currentFile *protogen.File) {
	name := msg.GoIdent.GoName

	// Separate fields by encode strategy
	var directFields, optionalScalarFields, optionalMsgFields, repeatedMsgFields []*protogen.Field
	for _, field := range msg.Fields {
		if field.Desc.IsList() && field.Desc.Kind() == protoreflect.MessageKind {
			repeatedMsgFields = append(repeatedMsgFields, field)
		} else if field.Desc.IsList() || field.Desc.IsMap() {
			directFields = append(directFields, field) // repeated scalars and maps encode directly
		} else if isFieldOptional(field) && field.Desc.Kind() == protoreflect.MessageKind {
			optionalMsgFields = append(optionalMsgFields, field)
		} else if field.Desc.HasOptionalKeyword() {
			optionalScalarFields = append(optionalScalarFields, field)
		} else {
			directFields = append(directFields, field)
		}
	}

	needsResult := len(optionalScalarFields) > 0 || len(optionalMsgFields) > 0 || len(repeatedMsgFields) > 0

	g.P("function ", moduleName, ".encodeJson", name, "(v: ", name, "): {[string]: any}")
	if !needsResult {
		// All fields are direct — simple return
		g.P("\treturn {")
		for _, field := range msg.Fields {
			luauName := snakeToCamel(string(field.Desc.Name()))
			g.P("\t\t", luauName, " = v.", luauName, ",")
		}
		g.P("\t}")
	} else {
		g.P("\tlocal result: {[string]: any} = {")
		for _, field := range directFields {
			luauName := snakeToCamel(string(field.Desc.Name()))
			g.P("\t\t", luauName, " = v.", luauName, ",")
		}
		g.P("\t}")

		for _, field := range optionalScalarFields {
			luauName := snakeToCamel(string(field.Desc.Name()))
			g.P("\tif v.", luauName, " ~= nil then result.", luauName, " = v.", luauName, " end")
		}

		for _, field := range optionalMsgFields {
			luauName := snakeToCamel(string(field.Desc.Name()))
			if isFieldFromDifferentFile(field, currentFile) {
				// DataTypes ref — use optional helper
				typeName := field.Message.GoIdent.GoName
				g.P("\tresult.", luauName, " = encodeOptional", typeName, "(v.", luauName, ")")
			} else {
				// Local ref — if-then encode
				typeName := field.Message.GoIdent.GoName
				g.P("\tif v.", luauName, " ~= nil then result.", luauName, " = ", moduleName, ".encodeJson", typeName, "(v.", luauName, ") end")
			}
		}

		for _, field := range repeatedMsgFields {
			luauName := snakeToCamel(string(field.Desc.Name()))
			typeName := field.Message.GoIdent.GoName
			isImported := isFieldFromDifferentFile(field, currentFile)
			g.P("\tlocal ", luauName, "Encoded: {any} = {}")
			g.P("\tfor _, item in v.", luauName, " do")
			if isImported {
				g.P("\t\ttable.insert(", luauName, "Encoded, DataTypes.encodeJson", typeName, "(item))")
			} else {
				g.P("\t\ttable.insert(", luauName, "Encoded, ", moduleName, ".encodeJson", typeName, "(item))")
			}
			g.P("\tend")
			g.P("\tresult.", luauName, " = ", luauName, "Encoded")
		}

		g.P("\treturn result")
	}
	g.P("end")
	g.P()
}

// emitSchemaDecode emits the JSON decoder for a message.
func emitSchemaDecode(g *protogen.GeneratedFile, msg *protogen.Message, moduleName string, currentFile *protogen.File) {
	name := msg.GoIdent.GoName

	// Check for repeated message fields that need pre-processing
	var repeatedMsgFields []*protogen.Field
	for _, field := range msg.Fields {
		if field.Desc.IsList() && field.Desc.Kind() == protoreflect.MessageKind {
			repeatedMsgFields = append(repeatedMsgFields, field)
		}
	}

	g.P("function ", moduleName, ".decodeJson", name, "(json: {[string]: any}): ", name)

	// Pre-decode repeated message fields
	for _, field := range repeatedMsgFields {
		luauName := snakeToCamel(string(field.Desc.Name()))
		typeName := field.Message.GoIdent.GoName
		isImported := isFieldFromDifferentFile(field, currentFile)
		g.P("\tlocal ", luauName, "Decoded: {", luauTypeRefGeneric(field, currentFile)[1:len(luauTypeRefGeneric(field, currentFile))-1], "} = {}")
		g.P("\tif json.", luauName, " ~= nil then")
		g.P("\t\tfor _, item in json.", luauName, " :: {{[string]: any}} do")
		if isImported {
			g.P("\t\t\ttable.insert(", luauName, "Decoded, DataTypes.decodeJson", typeName, "(item))")
		} else {
			g.P("\t\t\ttable.insert(", luauName, "Decoded, ", moduleName, ".decodeJson", typeName, "(item))")
		}
		g.P("\t\tend")
		g.P("\tend")
	}

	g.P("\treturn {")
	for _, field := range msg.Fields {
		luauName := snakeToCamel(string(field.Desc.Name()))

		if field.Desc.IsList() && field.Desc.Kind() == protoreflect.MessageKind {
			// Repeated message — use pre-decoded local
			g.P("\t\t", luauName, " = ", luauName, "Decoded,")
			continue
		}

		if field.Desc.IsList() {
			// Repeated scalar — direct with default
			g.P("\t\t", luauName, " = json.", luauName, " or {},")
			continue
		}

		if field.Desc.IsMap() {
			g.P("\t\t", luauName, " = json.", luauName, " or {},")
			continue
		}

		if isFieldOptional(field) && field.Desc.Kind() == protoreflect.MessageKind {
			if isFieldFromDifferentFile(field, currentFile) {
				// Optional DataTypes ref — use helper
				typeName := field.Message.GoIdent.GoName
				g.P("\t\t", luauName, " = decodeOptional", typeName, "(json.", luauName, "),")
			} else {
				// Optional local ref — if-then-else
				typeName := field.Message.GoIdent.GoName
				g.P("\t\t", luauName, " = if json.", luauName, " then ", moduleName, ".decodeJson", typeName, "(json.", luauName, ") else nil,")
			}
			continue
		}

		if field.Desc.HasOptionalKeyword() {
			// Optional scalar — pass through (nil is valid)
			g.P("\t\t", luauName, " = json.", luauName, ",")
			continue
		}

		// Required scalar — with default
		if field.Desc.Kind() == protoreflect.BoolKind {
			g.P("\t\t", luauName, " = if json.", luauName, " ~= nil then json.", luauName, " else false,")
		} else {
			dflt := luauDefaultValueGeneric(field)
			if dflt != "" {
				g.P("\t\t", luauName, " = json.", luauName, " or ", dflt, ",")
			} else {
				g.P("\t\t", luauName, " = json.", luauName, ",")
			}
		}
	}
	g.P("\t}")
	g.P("end")
	g.P()
}
```

- [ ] **Step 4: Run codec test**

Run: `cd protoc-gen-luau && go test -run TestSchemaCodecsSimple -v`

Expected: PASS.

- [ ] **Step 5: Add test for repeated message and local ref codecs**

Add to `gen_schema_test.go`:

```go
func TestSchemaCodecsComplex(t *testing.T) {
	schemaOpts := &descriptorpb.FileOptions{}
	proto.SetExtension(schemaOpts, luauoptions.E_LuauGenerator, "schema")

	schemaFile := &descriptorpb.FileDescriptorProto{
		Name:       proto.String("complex_codec.proto"),
		Syntax:     proto.String("proto3"),
		Dependency: []string{"data_types.proto"},
		Options:    schemaOpts,
		MessageType: []*descriptorpb.DescriptorProto{
			{
				Name: proto.String("ChildRecord"),
				Field: []*descriptorpb.FieldDescriptorProto{
					{Name: proto.String("name"), Number: proto.Int32(1),
						Type: descriptorpb.FieldDescriptorProto_TYPE_STRING.Enum(), JsonName: proto.String("name")},
				},
			},
			{
				Name: proto.String("ParentRecord"),
				Field: []*descriptorpb.FieldDescriptorProto{
					// Optional local ref
					{Name: proto.String("child"), Number: proto.Int32(1),
						Type: descriptorpb.FieldDescriptorProto_TYPE_MESSAGE.Enum(),
						TypeName: proto.String(".ChildRecord"), JsonName: proto.String("child"),
						Proto3Optional: proto.Bool(true)},
					// Repeated local ref
					{Name: proto.String("children"), Number: proto.Int32(2),
						Type: descriptorpb.FieldDescriptorProto_TYPE_MESSAGE.Enum(),
						TypeName: proto.String(".ChildRecord"), JsonName: proto.String("children"),
						Label: descriptorpb.FieldDescriptorProto_LABEL_REPEATED.Enum()},
					// Repeated DataTypes ref
					{Name: proto.String("types"), Number: proto.Int32(3),
						Type: descriptorpb.FieldDescriptorProto_TYPE_MESSAGE.Enum(),
						TypeName: proto.String(".StringType"), JsonName: proto.String("types"),
						Label: descriptorpb.FieldDescriptorProto_LABEL_REPEATED.Enum()},
				},
			},
		},
	}

	req := buildSchemaRequest(t, schemaFile)
	output := getSchemaOutput(t, req, "complex_codec.luau")

	// Encode: optional local ref uses if-then
	if !strings.Contains(output, "ComplexCodec.encodeJsonChildRecord(v.child)") {
		t.Error("encode should call ComplexCodec.encodeJsonChildRecord for optional local ref")
	}
	// Encode: repeated local ref uses iteration
	if !strings.Contains(output, "ComplexCodec.encodeJsonChildRecord(item)") {
		t.Error("encode should iterate repeated local ref with ComplexCodec.encodeJsonChildRecord")
	}
	// Encode: repeated DataTypes ref uses iteration
	if !strings.Contains(output, "DataTypes.encodeJsonStringType(item)") {
		t.Error("encode should iterate repeated DataTypes ref with DataTypes.encodeJsonStringType")
	}

	// Decode: optional local ref uses if-then-else
	if !strings.Contains(output, "ComplexCodec.decodeJsonChildRecord(json.child)") {
		t.Error("decode should call ComplexCodec.decodeJsonChildRecord for optional local ref")
	}
	// Decode: repeated local ref uses iteration
	if !strings.Contains(output, "ComplexCodec.decodeJsonChildRecord(item)") {
		t.Error("decode should iterate repeated local ref")
	}
	// Decode: repeated DataTypes ref uses iteration
	if !strings.Contains(output, "DataTypes.decodeJsonStringType(item)") {
		t.Error("decode should iterate repeated DataTypes ref")
	}
}
```

- [ ] **Step 6: Run all schema tests**

Run: `cd protoc-gen-luau && go test -run "TestSchema" -v`

Expected: All tests PASS.

- [ ] **Step 7: Commit**

```bash
cd protoc-gen-luau && git add gen_schema_emitters.go gen_schema_test.go
git commit -m "feat: SchemaGenerator encode/decode codecs for all field kinds"
```

---

### Task 6: Backward Compatibility — Existing Golden Tests

**Files:**
- Modify: `protoc-gen-luau/generate_test.go`
- Modify: `protoc-gen-luau/gen_schema_test.go`

**Depends on:** Task 5

This task verifies that the SchemaGenerator produces output equivalent to the old table generators for existing proto files. The golden test with `testdata/request.pb` must pass with injected options.

- [ ] **Step 1: Run the golden test**

Run: `cd protoc-gen-luau && go test -run TestGenerateGolden -v`

Expected: Compare output. The SchemaGenerator may produce slightly different output than the old hardcoded generators. Document differences.

- [ ] **Step 2: Analyze and fix any differences**

Common differences to expect:
- **Ordering**: The old `gen_table_files.go` emitted compound types first, then FieldsMap, then Schema. The new SchemaGenerator uses topological sort, which may order differently. Fix by adjusting the emit order in `generateSchema()`.
- **Optional helpers**: The old code collected helpers from FieldsMap fields only. The new code collects from all messages. May produce additional helpers. Fix by scoping helper collection.
- **Hardcoded `table_schema.go` types**: `SingleFieldType`, `FieldGroup`, `TableSchema` were hardcoded. The SchemaGenerator generates them from proto message definitions. Output should match if proto messages are defined correctly in `testdata/request.pb`.

For each difference found, adjust the SchemaGenerator to match. The goal is byte-identical golden output.

- [ ] **Step 3: Update golden files if output is correct but different format**

If the SchemaGenerator output is functionally correct but has minor formatting differences from the old golden files, update the golden files:

Run: `cd protoc-gen-luau && UPDATE_GOLDEN=1 go test -run TestGenerateGolden -v`

Then verify: `cd protoc-gen-luau && go test -run TestGenerateGolden -v`

Expected: PASS.

- [ ] **Step 4: Run full test suite**

Run: `cd protoc-gen-luau && go test ./... -v`

Expected: All tests PASS (except possibly `gen_table_schema_test.go` tests that reference old function names — these will be deleted in Task 8).

- [ ] **Step 5: Commit**

```bash
cd protoc-gen-luau && git add generate_test.go gen_schema.go gen_schema_emitters.go testdata/golden/
git commit -m "feat: backward-compatible SchemaGenerator output, golden tests pass"
```

---

### Task 7: Subscription Schema Golden Test

**Files:**
- Modify: `protoc-gen-luau/gen_schema_test.go`

**Depends on:** Task 6

This task adds the full subscription/fund domain test from the spec, exercising all field kinds.

- [ ] **Step 1: Write subscription schema test**

Add to `gen_schema_test.go` a comprehensive test that creates a synthetic CodeGeneratorRequest with the subscription domain messages. The test must include:

- `IndividualInvestorInfo` — required message ref (`name: IndividualNameType`), optional scalars, optional DataTypes refs
- `EntityInvestorInfo` — required string, optional DataTypes refs
- `SubscriptionData` — optional local refs (`IndividualInvestorInfo`), repeated DataTypes refs (`Signatory`), repeated local refs
- `SubscriptionWorkflowData` — required DataTypes ref, optional scalars, repeated strings
- `FundInfo` — nested message (`KeyLabel`), repeated nested messages
- `FundSubInterface` — repeated `SubscriptionData` (3-level nesting test), optional `FundInfo`

Key assertions:
- Dependency order: `IndividualInvestorInfo` and `EntityInvestorInfo` before `SubscriptionData`, `SubscriptionData` and `FundInfo` before `FundSubInterface`
- Required fields have no `?` in type def
- Optional fields have `?` in type def
- Repeated fields are `{Type}` with no `?`
- Encode uses correct dispatch (optional helper for DataTypes, module-qualified for local, iteration for repeated)
- Decode uses correct dispatch (helper for DataTypes, if-then-else for local, iteration loop for repeated)

```go
func TestSubscriptionSchemaGolden(t *testing.T) {
	// Build a comprehensive subscription-domain schema test.
	// This exercises all field kinds from the spec.

	schemaOpts := &descriptorpb.FileOptions{}
	proto.SetExtension(schemaOpts, luauoptions.E_LuauGenerator, "schema")

	// data_types.proto with the types referenced by subscription schema
	dtOpts := &descriptorpb.FileOptions{}
	proto.SetExtension(dtOpts, luauoptions.E_LuauGenerator, "data_types")

	dtFile := &descriptorpb.FileDescriptorProto{
		Name:    proto.String("data_types.proto"),
		Syntax:  proto.String("proto3"),
		Options: dtOpts,
		MessageType: []*descriptorpb.DescriptorProto{
			{Name: proto.String("StringType"), Field: []*descriptorpb.FieldDescriptorProto{
				{Name: proto.String("type_id"), Number: proto.Int32(1), Type: descriptorpb.FieldDescriptorProto_TYPE_STRING.Enum(), JsonName: proto.String("typeId")},
				{Name: proto.String("value"), Number: proto.Int32(2), Type: descriptorpb.FieldDescriptorProto_TYPE_STRING.Enum(), JsonName: proto.String("value")},
			}},
			{Name: proto.String("RadioGroupType"), Field: []*descriptorpb.FieldDescriptorProto{
				{Name: proto.String("type_id"), Number: proto.Int32(1), Type: descriptorpb.FieldDescriptorProto_TYPE_STRING.Enum(), JsonName: proto.String("typeId")},
				{Name: proto.String("value"), Number: proto.Int32(2), Type: descriptorpb.FieldDescriptorProto_TYPE_STRING.Enum(), JsonName: proto.String("value")},
			}},
			{Name: proto.String("MoneyType"), Field: []*descriptorpb.FieldDescriptorProto{
				{Name: proto.String("type_id"), Number: proto.Int32(1), Type: descriptorpb.FieldDescriptorProto_TYPE_STRING.Enum(), JsonName: proto.String("typeId")},
				{Name: proto.String("value"), Number: proto.Int32(2), Type: descriptorpb.FieldDescriptorProto_TYPE_DOUBLE.Enum(), JsonName: proto.String("value")},
			}},
			{Name: proto.String("IndividualNameType"), Field: []*descriptorpb.FieldDescriptorProto{
				{Name: proto.String("type_id"), Number: proto.Int32(1), Type: descriptorpb.FieldDescriptorProto_TYPE_STRING.Enum(), JsonName: proto.String("typeId")},
			}},
			{Name: proto.String("SignatoryType"), Field: []*descriptorpb.FieldDescriptorProto{
				{Name: proto.String("type_id"), Number: proto.Int32(1), Type: descriptorpb.FieldDescriptorProto_TYPE_STRING.Enum(), JsonName: proto.String("typeId")},
			}},
			{Name: proto.String("ContactInfoType"), Field: []*descriptorpb.FieldDescriptorProto{
				{Name: proto.String("type_id"), Number: proto.Int32(1), Type: descriptorpb.FieldDescriptorProto_TYPE_STRING.Enum(), JsonName: proto.String("typeId")},
			}},
			{Name: proto.String("DateTimeType"), Field: []*descriptorpb.FieldDescriptorProto{
				{Name: proto.String("type_id"), Number: proto.Int32(1), Type: descriptorpb.FieldDescriptorProto_TYPE_STRING.Enum(), JsonName: proto.String("typeId")},
			}},
		},
	}

	subFile := &descriptorpb.FileDescriptorProto{
		Name:       proto.String("subscription_data.proto"),
		Syntax:     proto.String("proto3"),
		Dependency: []string{"data_types.proto"},
		Options:    schemaOpts,
		MessageType: []*descriptorpb.DescriptorProto{
			// IndividualInvestorInfo — required msg ref, optional scalars, optional DataTypes refs
			{
				Name: proto.String("IndividualInvestorInfo"),
				Field: []*descriptorpb.FieldDescriptorProto{
					{Name: proto.String("type_id"), Number: proto.Int32(1), Type: descriptorpb.FieldDescriptorProto_TYPE_STRING.Enum(), JsonName: proto.String("typeId")},
					{Name: proto.String("name"), Number: proto.Int32(2), Type: descriptorpb.FieldDescriptorProto_TYPE_MESSAGE.Enum(), TypeName: proto.String(".IndividualNameType"), JsonName: proto.String("name")},
					{Name: proto.String("initials"), Number: proto.Int32(3), Type: descriptorpb.FieldDescriptorProto_TYPE_STRING.Enum(), JsonName: proto.String("initials"), Proto3Optional: proto.Bool(true)},
					{Name: proto.String("nationality"), Number: proto.Int32(4), Type: descriptorpb.FieldDescriptorProto_TYPE_STRING.Enum(), JsonName: proto.String("nationality"), Proto3Optional: proto.Bool(true)},
				},
			},
			// SubscriptionWorkflowData — required DataTypes ref, optional scalars, repeated strings
			{
				Name: proto.String("SubscriptionWorkflowData"),
				Field: []*descriptorpb.FieldDescriptorProto{
					{Name: proto.String("type_id"), Number: proto.Int32(1), Type: descriptorpb.FieldDescriptorProto_TYPE_STRING.Enum(), JsonName: proto.String("typeId")},
					{Name: proto.String("date_submitted"), Number: proto.Int32(2), Type: descriptorpb.FieldDescriptorProto_TYPE_MESSAGE.Enum(), TypeName: proto.String(".DateTimeType"), JsonName: proto.String("dateSubmitted")},
					{Name: proto.String("client_matter_id"), Number: proto.Int32(3), Type: descriptorpb.FieldDescriptorProto_TYPE_STRING.Enum(), JsonName: proto.String("clientMatterId"), Proto3Optional: proto.Bool(true)},
					{Name: proto.String("tags"), Number: proto.Int32(4), Type: descriptorpb.FieldDescriptorProto_TYPE_STRING.Enum(), Label: descriptorpb.FieldDescriptorProto_LABEL_REPEATED.Enum(), JsonName: proto.String("tags")},
				},
			},
			// SubscriptionData — local refs, repeated DataTypes refs, repeated local refs
			{
				Name: proto.String("SubscriptionData"),
				Field: []*descriptorpb.FieldDescriptorProto{
					{Name: proto.String("type_id"), Number: proto.Int32(1), Type: descriptorpb.FieldDescriptorProto_TYPE_STRING.Enum(), JsonName: proto.String("typeId")},
					{Name: proto.String("investor_type"), Number: proto.Int32(2), Type: descriptorpb.FieldDescriptorProto_TYPE_MESSAGE.Enum(), TypeName: proto.String(".RadioGroupType"), JsonName: proto.String("investorType"), Proto3Optional: proto.Bool(true)},
					{Name: proto.String("individual_investor"), Number: proto.Int32(3), Type: descriptorpb.FieldDescriptorProto_TYPE_MESSAGE.Enum(), TypeName: proto.String(".IndividualInvestorInfo"), JsonName: proto.String("individualInvestor"), Proto3Optional: proto.Bool(true)},
					{Name: proto.String("commitment_amount"), Number: proto.Int32(4), Type: descriptorpb.FieldDescriptorProto_TYPE_MESSAGE.Enum(), TypeName: proto.String(".MoneyType"), JsonName: proto.String("commitmentAmount"), Proto3Optional: proto.Bool(true)},
					{Name: proto.String("primary_contacts"), Number: proto.Int32(5), Type: descriptorpb.FieldDescriptorProto_TYPE_MESSAGE.Enum(), TypeName: proto.String(".ContactInfoType"), Label: descriptorpb.FieldDescriptorProto_LABEL_REPEATED.Enum(), JsonName: proto.String("primaryContacts")},
					{Name: proto.String("lp_signers"), Number: proto.Int32(6), Type: descriptorpb.FieldDescriptorProto_TYPE_MESSAGE.Enum(), TypeName: proto.String(".SignatoryType"), Label: descriptorpb.FieldDescriptorProto_LABEL_REPEATED.Enum(), JsonName: proto.String("lpSigners")},
					{Name: proto.String("workflow_data"), Number: proto.Int32(7), Type: descriptorpb.FieldDescriptorProto_TYPE_MESSAGE.Enum(), TypeName: proto.String(".SubscriptionWorkflowData"), JsonName: proto.String("workflowData"), Proto3Optional: proto.Bool(true)},
				},
			},
			// FundSubInterface — repeated SubscriptionData (3-level nesting)
			{
				Name: proto.String("FundSubInterface"),
				Field: []*descriptorpb.FieldDescriptorProto{
					{Name: proto.String("type_id"), Number: proto.Int32(1), Type: descriptorpb.FieldDescriptorProto_TYPE_STRING.Enum(), JsonName: proto.String("typeId")},
					{Name: proto.String("subscriptions"), Number: proto.Int32(2), Type: descriptorpb.FieldDescriptorProto_TYPE_MESSAGE.Enum(), TypeName: proto.String(".SubscriptionData"), Label: descriptorpb.FieldDescriptorProto_LABEL_REPEATED.Enum(), JsonName: proto.String("subscriptions")},
				},
			},
		},
	}

	req := &pluginpb.CodeGeneratorRequest{
		FileToGenerate: []string{"subscription_data.proto"},
		ProtoFile:      []*descriptorpb.FileDescriptorProto{dtFile, subFile},
	}

	output := getSchemaOutput(t, req, "subscription_data.luau")

	// Dependency ordering
	iiIdx := strings.Index(output, "export type IndividualInvestorInfo")
	swdIdx := strings.Index(output, "export type SubscriptionWorkflowData")
	sdIdx := strings.Index(output, "export type SubscriptionData")
	fsiIdx := strings.Index(output, "export type FundSubInterface")

	if iiIdx < 0 || swdIdx < 0 || sdIdx < 0 || fsiIdx < 0 {
		t.Fatal("missing type definitions")
	}
	if iiIdx > sdIdx {
		t.Error("IndividualInvestorInfo must come before SubscriptionData")
	}
	if swdIdx > sdIdx {
		t.Error("SubscriptionWorkflowData must come before SubscriptionData")
	}
	if sdIdx > fsiIdx {
		t.Error("SubscriptionData must come before FundSubInterface")
	}

	// IndividualInvestorInfo — required message field (no ?)
	if !strings.Contains(output, "\tname: DataTypes.IndividualNameType,") {
		t.Error("IndividualInvestorInfo.name should be required DataTypes.IndividualNameType (no ?)")
	}
	// Optional scalar (has ?)
	if !strings.Contains(output, "\tinitials: string?,") {
		t.Error("IndividualInvestorInfo.initials should be optional string?")
	}

	// SubscriptionData — optional local ref
	if !strings.Contains(output, "\tindividualInvestor: IndividualInvestorInfo?,") {
		t.Error("SubscriptionData.individualInvestor should be optional IndividualInvestorInfo?")
	}
	// Repeated DataTypes ref
	if !strings.Contains(output, "\tprimaryContacts: {DataTypes.ContactInfoType},") {
		t.Error("SubscriptionData.primaryContacts should be {DataTypes.ContactInfoType}")
	}
	// Repeated local ref at top level
	if !strings.Contains(output, "\tsubscriptions: {SubscriptionData},") {
		t.Error("FundSubInterface.subscriptions should be {SubscriptionData}")
	}

	// Constructors exist
	for _, name := range []string{"makeIndividualInvestorInfo", "makeSubscriptionWorkflowData", "makeSubscriptionData", "makeFundSubInterface"} {
		if !strings.Contains(output, "function SubscriptionData."+name+"(") {
			t.Errorf("missing constructor %s", name)
		}
	}

	// Codecs exist
	for _, name := range []string{"encodeJsonIndividualInvestorInfo", "decodeJsonIndividualInvestorInfo", "encodeJsonSubscriptionData", "decodeJsonSubscriptionData", "encodeJsonFundSubInterface", "decodeJsonFundSubInterface"} {
		if !strings.Contains(output, "function SubscriptionData."+name+"(") {
			t.Errorf("missing codec %s", name)
		}
	}

	// Encode: repeated DataTypes ref iteration
	if !strings.Contains(output, "DataTypes.encodeJsonContactInfoType(item)") {
		t.Error("encode should iterate primaryContacts with DataTypes.encodeJsonContactInfoType")
	}
	// Encode: repeated local ref iteration
	if !strings.Contains(output, "SubscriptionData.encodeJsonSubscriptionData(item)") {
		t.Error("encode should iterate subscriptions with SubscriptionData.encodeJsonSubscriptionData")
	}
	// Encode: optional local ref
	if !strings.Contains(output, "SubscriptionData.encodeJsonIndividualInvestorInfo(v.individualInvestor)") {
		t.Error("encode should call SubscriptionData.encodeJsonIndividualInvestorInfo for optional local ref")
	}

	// Optional helpers for DataTypes types
	if !strings.Contains(output, "decodeOptionalRadioGroupType") {
		t.Error("missing decodeOptionalRadioGroupType helper")
	}
	if !strings.Contains(output, "encodeOptionalMoneyType") {
		t.Error("missing encodeOptionalMoneyType helper")
	}

	// Footer
	if !strings.HasSuffix(strings.TrimSpace(output), "return SubscriptionData") {
		t.Error("missing return footer")
	}
}
```

- [ ] **Step 2: Run subscription test**

Run: `cd protoc-gen-luau && go test -run TestSubscriptionSchemaGolden -v`

Expected: PASS.

- [ ] **Step 3: Run full test suite**

Run: `cd protoc-gen-luau && go test ./... -v`

Expected: All tests PASS.

- [ ] **Step 4: Commit**

```bash
cd protoc-gen-luau && git add gen_schema_test.go
git commit -m "test: add subscription schema golden test exercising all field kinds"
```

---

### Task 8: Cleanup — Delete Old Files

**Files:**
- Delete: `protoc-gen-luau/gen_table_schema.go`
- Delete: `protoc-gen-luau/gen_table_files.go`
- Delete: `protoc-gen-luau/gen_table_schema_test.go`

**Depends on:** Task 6, Task 7

- [ ] **Step 1: Delete old generator files**

Remove `gen_table_schema.go` and `gen_table_files.go`. Their functions are now either:
- Absorbed into `gen_schema.go` / `gen_schema_emitters.go` (SchemaGenerator)
- Moved to `luau_naming.go` (`snakeToPascal`, `deriveModuleName`, `isFieldFromDifferentFile`)
- Moved to `classify.go` (`classifyMessages`, types)

Also remove `gen_table_schema_test.go` — its tests (`TestTableSchemaOutput`, `TestSourceTableOutput`, `TestTargetTableOutput`) test old function names. Equivalent coverage is provided by the golden test and the new schema tests.

```bash
cd protoc-gen-luau
rm gen_table_schema.go gen_table_files.go gen_table_schema_test.go
```

- [ ] **Step 2: Verify compilation**

Run: `cd protoc-gen-luau && go build ./...`

Expected: No errors. If there are undefined function references, they need to be traced:
- Functions from `gen_table_files.go` that were table-specific (`emitFieldsMapDef`, `emitFieldsMapConstructor`, `emitTableFileSchemaTypeDef`, etc.) should NOT be referenced anymore — the SchemaGenerator uses its own generic emitters.
- Functions from `gen_table_files.go` that were shared (`emitPerModuleOptionalHelpers`, `fieldsMapFieldHelperName`, `emitFieldsMapEncode`, `emitFieldsMapDecode`, etc.) should also not be referenced — replaced by generic equivalents.
- If any old function is still referenced, check where and replace with the generic equivalent.

- [ ] **Step 3: Run full test suite**

Run: `cd protoc-gen-luau && go test ./... -v`

Expected: All tests PASS.

- [ ] **Step 4: Verify no dead code**

Run: `cd protoc-gen-luau && go vet ./...`

Expected: No errors. Check for unused functions/variables.

- [ ] **Step 5: Commit**

```bash
cd protoc-gen-luau
git add -A
git commit -m "refactor: delete old table generators, SchemaGenerator handles all schemas"
```
