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

// buildSchemaOutput creates a CodeGeneratorRequest with a data_types.proto and a schema.proto,
// runs the generator, and returns the content of the schema output file.
func buildSchemaOutput(t *testing.T, dataTypesMessages []*descriptorpb.DescriptorProto, schemaMessages []*descriptorpb.DescriptorProto) string {
	t.Helper()

	dtOpts := &descriptorpb.FileOptions{GoPackage: proto.String("gen/datatypes;datatypes")}
	proto.SetExtension(dtOpts, luauoptions.E_LuauGenerator, "data_types")

	schemaOpts := &descriptorpb.FileOptions{GoPackage: proto.String("gen/myschema;myschema")}
	proto.SetExtension(schemaOpts, luauoptions.E_LuauGenerator, "schema")
	proto.SetExtension(schemaOpts, luauoptions.E_LuauDataTypesRequire, "@lib/data_types")

	req := &pluginpb.CodeGeneratorRequest{
		FileToGenerate: []string{"data_types.proto", "my_schema.proto"},
		ProtoFile: []*descriptorpb.FileDescriptorProto{
			{
				Name:        proto.String("data_types.proto"),
				Package:     proto.String("data_types"),
				Syntax:      proto.String("proto3"),
				Options:     dtOpts,
				MessageType: dataTypesMessages,
			},
			{
				Name:        proto.String("my_schema.proto"),
				Package:     proto.String("my_schema"),
				Syntax:      proto.String("proto3"),
				Options:     schemaOpts,
				Dependency:  []string{"data_types.proto"},
				MessageType: schemaMessages,
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
	for _, f := range resp.File {
		if f.GetName() == "my_schema.luau" {
			return f.GetContent()
		}
	}
	t.Fatal("my_schema.luau not found in response")
	return ""
}

// Helper to create a simple message descriptor with scalar fields.
func makeScalarMessage(name string, fields ...*descriptorpb.FieldDescriptorProto) *descriptorpb.DescriptorProto {
	return &descriptorpb.DescriptorProto{
		Name:  proto.String(name),
		Field: fields,
	}
}

func stringField(name string, number int32) *descriptorpb.FieldDescriptorProto {
	return &descriptorpb.FieldDescriptorProto{
		Name:   proto.String(name),
		Number: proto.Int32(number),
		Type:   descriptorpb.FieldDescriptorProto_TYPE_STRING.Enum(),
		Label:  descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL.Enum(),
	}
}

func numberField(name string, number int32) *descriptorpb.FieldDescriptorProto {
	return &descriptorpb.FieldDescriptorProto{
		Name:   proto.String(name),
		Number: proto.Int32(number),
		Type:   descriptorpb.FieldDescriptorProto_TYPE_INT32.Enum(),
		Label:  descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL.Enum(),
	}
}

func boolField(name string, number int32) *descriptorpb.FieldDescriptorProto {
	return &descriptorpb.FieldDescriptorProto{
		Name:   proto.String(name),
		Number: proto.Int32(number),
		Type:   descriptorpb.FieldDescriptorProto_TYPE_BOOL.Enum(),
		Label:  descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL.Enum(),
	}
}

func optionalStringField(name string, number int32) *descriptorpb.FieldDescriptorProto {
	return &descriptorpb.FieldDescriptorProto{
		Name:           proto.String(name),
		Number:         proto.Int32(number),
		Type:           descriptorpb.FieldDescriptorProto_TYPE_STRING.Enum(),
		Label:          descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL.Enum(),
		Proto3Optional: proto.Bool(true),
		OneofIndex:     proto.Int32(0),
	}
}

func messageField(name string, number int32, typeName string) *descriptorpb.FieldDescriptorProto {
	return &descriptorpb.FieldDescriptorProto{
		Name:     proto.String(name),
		Number:   proto.Int32(number),
		Type:     descriptorpb.FieldDescriptorProto_TYPE_MESSAGE.Enum(),
		Label:    descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL.Enum(),
		TypeName: proto.String(typeName),
	}
}

func repeatedStringField(name string, number int32) *descriptorpb.FieldDescriptorProto {
	return &descriptorpb.FieldDescriptorProto{
		Name:   proto.String(name),
		Number: proto.Int32(number),
		Type:   descriptorpb.FieldDescriptorProto_TYPE_STRING.Enum(),
		Label:  descriptorpb.FieldDescriptorProto_LABEL_REPEATED.Enum(),
	}
}

func repeatedMessageField(name string, number int32, typeName string) *descriptorpb.FieldDescriptorProto {
	return &descriptorpb.FieldDescriptorProto{
		Name:     proto.String(name),
		Number:   proto.Int32(number),
		Type:     descriptorpb.FieldDescriptorProto_TYPE_MESSAGE.Enum(),
		Label:    descriptorpb.FieldDescriptorProto_LABEL_REPEATED.Enum(),
		TypeName: proto.String(typeName),
	}
}

// TestSchemaTypeDefsSimple verifies type definitions for a message with
// required string, required number, required bool, and optional string fields.
func TestSchemaTypeDefsSimple(t *testing.T) {
	// optional string requires a synthetic oneof
	schemaMsg := &descriptorpb.DescriptorProto{
		Name: proto.String("MyRecord"),
		Field: []*descriptorpb.FieldDescriptorProto{
			stringField("name", 1),
			numberField("count", 2),
			boolField("active", 3),
			optionalStringField("nickname", 4),
		},
		OneofDecl: []*descriptorpb.OneofDescriptorProto{
			{Name: proto.String("_nickname")},
		},
	}

	output := buildSchemaOutput(t, nil, []*descriptorpb.DescriptorProto{schemaMsg})

	// Type definition
	if !strings.Contains(output, "export type MyRecord = {") {
		t.Error("missing MyRecord type definition")
	}
	// Required string — no ?
	if !strings.Contains(output, "\tname: string,") {
		t.Error("name should be required string (no ?)")
	}
	// Required number — no ?
	if !strings.Contains(output, "\tcount: number,") {
		t.Error("count should be required number (no ?)")
	}
	// Required bool — no ?
	if !strings.Contains(output, "\tactive: boolean,") {
		t.Error("active should be required boolean (no ?)")
	}
	// Optional string — has ?
	if !strings.Contains(output, "\tnickname: string?,") {
		t.Error("nickname should be optional string (with ?)")
	}

	// Constructor — all params optional
	if !strings.Contains(output, "function MySchema.makeMyRecord(") {
		t.Error("missing makeMyRecord constructor")
	}
	// Required string default: or ""
	if !strings.Contains(output, `name = c.name or "",`) {
		t.Error("constructor name should default to empty string")
	}
	// Required number default: or 0
	if !strings.Contains(output, "count = c.count or 0,") {
		t.Error("constructor count should default to 0")
	}
	// Required bool: special if ~= nil handling
	if !strings.Contains(output, "active = if c.active ~= nil then c.active else false,") {
		t.Error("constructor active should use special bool handling")
	}
	// Optional string: no default (nil OK)
	if !strings.Contains(output, "\t\tnickname = c.nickname,") {
		t.Error("constructor nickname should pass through (nil OK)")
	}
}

// TestSchemaTypeDefsWithRefs verifies type definitions for messages referencing
// both local messages and DataTypes messages, ensuring correct dependency order
// and type prefixing.
func TestSchemaTypeDefsWithRefs(t *testing.T) {
	dtMessages := []*descriptorpb.DescriptorProto{
		makeScalarMessage("StringType",
			stringField("type_id", 1),
			stringField("value", 2),
		),
	}

	// Inner must come before Outer in output (topo sort)
	inner := makeScalarMessage("Inner",
		stringField("label", 1),
	)
	outer := &descriptorpb.DescriptorProto{
		Name: proto.String("Outer"),
		Field: []*descriptorpb.FieldDescriptorProto{
			stringField("title", 1),
			messageField("detail", 2, ".my_schema.Inner"),
			messageField("str_type", 3, ".data_types.StringType"),
		},
	}

	output := buildSchemaOutput(t, dtMessages, []*descriptorpb.DescriptorProto{outer, inner})

	// Inner type def should exist
	if !strings.Contains(output, "export type Inner = {") {
		t.Error("missing Inner type def")
	}
	// Outer type def should exist
	if !strings.Contains(output, "export type Outer = {") {
		t.Error("missing Outer type def")
	}

	// Inner type should appear before Outer type (dependency ordering)
	innerIdx := strings.Index(output, "export type Inner = {")
	outerIdx := strings.Index(output, "export type Outer = {")
	if innerIdx >= outerIdx {
		t.Error("Inner should appear before Outer (topo sort)")
	}

	// Local message ref — no prefix, optional
	if !strings.Contains(output, "\tdetail: Inner?,") {
		t.Error("Outer.detail should be Inner? (local message, optional)")
	}
	// DataTypes message ref — prefixed, optional
	if !strings.Contains(output, "\tstrType: DataTypes.StringType?,") {
		t.Error("Outer.strType should be DataTypes.StringType? (imported, optional)")
	}

	// Constructor should have local ref
	if !strings.Contains(output, "function MySchema.makeOuter(") {
		t.Error("missing makeOuter constructor")
	}
	if !strings.Contains(output, "\tdetail: Inner?,") {
		t.Error("makeOuter config should have detail: Inner?")
	}
}

// TestSchemaCodecsSimple verifies encode/decode for required string, optional string,
// optional DataTypes ref, and repeated string fields.
func TestSchemaCodecsSimple(t *testing.T) {
	dtMessages := []*descriptorpb.DescriptorProto{
		makeScalarMessage("StringType",
			stringField("type_id", 1),
			stringField("value", 2),
		),
	}

	schemaMsg := &descriptorpb.DescriptorProto{
		Name: proto.String("SimpleMsg"),
		Field: []*descriptorpb.FieldDescriptorProto{
			stringField("title", 1),
			optionalStringField("subtitle", 2),
			messageField("str_ref", 3, ".data_types.StringType"),
			repeatedStringField("tags", 4),
		},
		OneofDecl: []*descriptorpb.OneofDescriptorProto{
			{Name: proto.String("_subtitle")},
		},
	}

	output := buildSchemaOutput(t, dtMessages, []*descriptorpb.DescriptorProto{schemaMsg})

	// Encode function exists
	if !strings.Contains(output, "function MySchema.encodeJsonSimpleMsg(") {
		t.Error("missing encodeJsonSimpleMsg")
	}
	// Decode function exists
	if !strings.Contains(output, "function MySchema.decodeJsonSimpleMsg(") {
		t.Error("missing decodeJsonSimpleMsg")
	}

	// Encode: required string — direct in result literal
	if !strings.Contains(output, "title = si.title,") {
		t.Error("encode should have direct title field")
	}
	// Encode: repeated string — direct
	if !strings.Contains(output, "tags = si.tags,") {
		t.Error("encode should have direct tags field")
	}
	// Encode: optional string — if check
	if !strings.Contains(output, "if si.subtitle ~= nil then") {
		t.Error("encode should nil-check optional subtitle")
	}
	// Encode: optional DataTypes ref — uses helper
	if !strings.Contains(output, "encodeOptionalStringType(si.strRef)") {
		t.Error("encode should use encodeOptionalStringType helper for strRef")
	}

	// Decode: required string — or ""
	if !strings.Contains(output, `title = json.title or "",`) {
		t.Error("decode should default title to empty string")
	}
	// Decode: optional string — pass through
	if !strings.Contains(output, "\t\tsubtitle = json.subtitle,") {
		t.Error("decode should pass through optional subtitle")
	}
	// Decode: optional DataTypes ref — uses helper
	if !strings.Contains(output, "decodeOptionalStringType(json.strRef)") {
		t.Error("decode should use decodeOptionalStringType helper for strRef")
	}
	// Decode: repeated string — or {}
	if !strings.Contains(output, "tags = json.tags or {},") {
		t.Error("decode should default tags to {}")
	}

	// Optional helper should be emitted
	if !strings.Contains(output, "local function decodeOptionalStringType(") {
		t.Error("missing decodeOptionalStringType helper")
	}
	if !strings.Contains(output, "local function encodeOptionalStringType(") {
		t.Error("missing encodeOptionalStringType helper")
	}
}

// TestSchemaCodecsComplex verifies encode/decode for optional local ref,
// repeated local ref, and repeated DataTypes ref fields.
func TestSchemaCodecsComplex(t *testing.T) {
	dtMessages := []*descriptorpb.DescriptorProto{
		makeScalarMessage("NumberType",
			stringField("type_id", 1),
			numberField("value", 2),
		),
	}

	child := makeScalarMessage("Child",
		stringField("label", 1),
		numberField("score", 2),
	)

	parent := &descriptorpb.DescriptorProto{
		Name: proto.String("Parent"),
		Field: []*descriptorpb.FieldDescriptorProto{
			stringField("name", 1),
			messageField("only_child", 2, ".my_schema.Child"),
			repeatedMessageField("children", 3, ".my_schema.Child"),
			repeatedMessageField("nums", 4, ".data_types.NumberType"),
		},
	}

	output := buildSchemaOutput(t, dtMessages, []*descriptorpb.DescriptorProto{parent, child})

	// Encode function
	if !strings.Contains(output, "function MySchema.encodeJsonParent(") {
		t.Error("missing encodeJsonParent")
	}
	// Decode function
	if !strings.Contains(output, "function MySchema.decodeJsonParent(") {
		t.Error("missing decodeJsonParent")
	}
	// Child codecs should also exist (needed by parent)
	if !strings.Contains(output, "function MySchema.encodeJsonChild(") {
		t.Error("missing encodeJsonChild")
	}
	if !strings.Contains(output, "function MySchema.decodeJsonChild(") {
		t.Error("missing decodeJsonChild")
	}

	// Encode: optional local message — if-then with Module.encodeJson
	if !strings.Contains(output, "MySchema.encodeJsonChild(p.onlyChild)") {
		t.Error("encode should use MySchema.encodeJsonChild for optional local ref")
	}
	if !strings.Contains(output, "if p.onlyChild ~= nil then") {
		t.Error("encode should nil-check optional local ref onlyChild")
	}

	// Encode: repeated local message — iterate with Module.encodeJson
	if !strings.Contains(output, "MySchema.encodeJsonChild(item)") {
		t.Error("encode should use MySchema.encodeJsonChild for repeated local children")
	}
	if !strings.Contains(output, "local childrenEncoded: {any} = {}") {
		t.Error("encode should create childrenEncoded array")
	}

	// Encode: repeated DataTypes message — iterate with DataTypes.encodeJson
	if !strings.Contains(output, "DataTypes.encodeJsonNumberType(item)") {
		t.Error("encode should use DataTypes.encodeJsonNumberType for repeated DataTypes nums")
	}
	if !strings.Contains(output, "local numsEncoded: {any} = {}") {
		t.Error("encode should create numsEncoded array")
	}

	// Decode: optional local message — if-then-else
	if !strings.Contains(output, "if json.onlyChild then MySchema.decodeJsonChild(json.onlyChild) else nil") {
		t.Error("decode should use if-then-else for optional local ref onlyChild")
	}

	// Decode: repeated local message — pre-decode loop
	if !strings.Contains(output, "local children: {Child} = {}") {
		t.Error("decode should create children local var")
	}
	if !strings.Contains(output, "MySchema.decodeJsonChild(item)") {
		t.Error("decode should use MySchema.decodeJsonChild for repeated local children")
	}

	// Decode: repeated DataTypes message — pre-decode loop
	if !strings.Contains(output, "local nums: {DataTypes.NumberType} = {}") {
		t.Error("decode should create nums local var")
	}
	if !strings.Contains(output, "DataTypes.decodeJsonNumberType(item)") {
		t.Error("decode should use DataTypes.decodeJsonNumberType for repeated DataTypes nums")
	}
}
