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

// buildSubscriptionOutput builds a CodeGeneratorRequest with data_types.proto and
// subscription_data.proto, runs the generator, and returns the subscription_data.luau content.
func buildSubscriptionOutput(t *testing.T, dataTypesMessages []*descriptorpb.DescriptorProto, schemaMessages []*descriptorpb.DescriptorProto) string {
	t.Helper()

	dtOpts := &descriptorpb.FileOptions{GoPackage: proto.String("gen/datatypes;datatypes")}
	proto.SetExtension(dtOpts, luauoptions.E_LuauGenerator, "data_types")

	schemaOpts := &descriptorpb.FileOptions{GoPackage: proto.String("gen/subscription;subscription")}
	proto.SetExtension(schemaOpts, luauoptions.E_LuauGenerator, "schema")
	proto.SetExtension(schemaOpts, luauoptions.E_LuauDataTypesRequire, "@lib/data_types")

	req := &pluginpb.CodeGeneratorRequest{
		FileToGenerate: []string{"data_types.proto", "subscription_data.proto"},
		ProtoFile: []*descriptorpb.FileDescriptorProto{
			{
				Name:        proto.String("data_types.proto"),
				Package:     proto.String("data_types"),
				Syntax:      proto.String("proto3"),
				Options:     dtOpts,
				MessageType: dataTypesMessages,
			},
			{
				Name:        proto.String("subscription_data.proto"),
				Package:     proto.String("subscription_data"),
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
		if f.GetName() == "subscription_data.luau" {
			return f.GetContent()
		}
	}
	t.Fatal("subscription_data.luau not found in response")
	return ""
}

// TestSubscriptionSchemaGolden exercises ALL field kinds the SchemaGenerator handles:
// required string, optional string (proto3), message ref (DataTypes), optional local message,
// repeated string, repeated DataTypes message, repeated local message.
func TestSubscriptionSchemaGolden(t *testing.T) {
	// data_types.proto messages
	dtMessages := []*descriptorpb.DescriptorProto{
		makeScalarMessage("StringType",
			stringField("type_id", 1),
			stringField("value", 2),
		),
		makeScalarMessage("RadioGroupType",
			stringField("type_id", 1),
			stringField("value", 2),
		),
		makeScalarMessage("MoneyType",
			stringField("type_id", 1),
			&descriptorpb.FieldDescriptorProto{
				Name:   proto.String("value"),
				Number: proto.Int32(2),
				Type:   descriptorpb.FieldDescriptorProto_TYPE_DOUBLE.Enum(),
				Label:  descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL.Enum(),
			},
		),
		makeScalarMessage("IndividualNameType",
			stringField("type_id", 1),
		),
		makeScalarMessage("SignatoryType",
			stringField("type_id", 1),
		),
		makeScalarMessage("ContactInfoType",
			stringField("type_id", 1),
		),
		makeScalarMessage("DateTimeType",
			stringField("type_id", 1),
		),
	}

	// IndividualInvestorInfo: required type_id, message ref name (IndividualNameType),
	// optional string initials, optional string nationality
	individualInvestorInfo := &descriptorpb.DescriptorProto{
		Name: proto.String("IndividualInvestorInfo"),
		Field: []*descriptorpb.FieldDescriptorProto{
			stringField("type_id", 1),
			messageField("name", 2, ".data_types.IndividualNameType"),
			{
				Name:           proto.String("initials"),
				Number:         proto.Int32(3),
				Type:           descriptorpb.FieldDescriptorProto_TYPE_STRING.Enum(),
				Label:          descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL.Enum(),
				Proto3Optional: proto.Bool(true),
				OneofIndex:     proto.Int32(0),
			},
			{
				Name:           proto.String("nationality"),
				Number:         proto.Int32(4),
				Type:           descriptorpb.FieldDescriptorProto_TYPE_STRING.Enum(),
				Label:          descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL.Enum(),
				Proto3Optional: proto.Bool(true),
				OneofIndex:     proto.Int32(1),
			},
		},
		OneofDecl: []*descriptorpb.OneofDescriptorProto{
			{Name: proto.String("_initials")},
			{Name: proto.String("_nationality")},
		},
	}

	// SubscriptionWorkflowData: required type_id, message ref date_submitted (DateTimeType),
	// optional string client_matter_id, repeated string tags
	subscriptionWorkflowData := &descriptorpb.DescriptorProto{
		Name: proto.String("SubscriptionWorkflowData"),
		Field: []*descriptorpb.FieldDescriptorProto{
			stringField("type_id", 1),
			messageField("date_submitted", 2, ".data_types.DateTimeType"),
			{
				Name:           proto.String("client_matter_id"),
				Number:         proto.Int32(3),
				Type:           descriptorpb.FieldDescriptorProto_TYPE_STRING.Enum(),
				Label:          descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL.Enum(),
				Proto3Optional: proto.Bool(true),
				OneofIndex:     proto.Int32(0),
			},
			repeatedStringField("tags", 4),
		},
		OneofDecl: []*descriptorpb.OneofDescriptorProto{
			{Name: proto.String("_client_matter_id")},
		},
	}

	// SubscriptionData: required type_id, optional DataTypes refs, optional local refs,
	// repeated DataTypes refs, repeated local refs
	subscriptionData := &descriptorpb.DescriptorProto{
		Name: proto.String("SubscriptionData"),
		Field: []*descriptorpb.FieldDescriptorProto{
			stringField("type_id", 1),
			messageField("investor_type", 2, ".data_types.RadioGroupType"),
			messageField("individual_investor", 3, ".subscription_data.IndividualInvestorInfo"),
			messageField("commitment_amount", 4, ".data_types.MoneyType"),
			repeatedMessageField("primary_contacts", 5, ".data_types.ContactInfoType"),
			repeatedMessageField("lp_signers", 6, ".data_types.SignatoryType"),
			messageField("workflow_data", 7, ".subscription_data.SubscriptionWorkflowData"),
		},
	}

	// FundSubInterface: required type_id, repeated local SubscriptionData
	fundSubInterface := &descriptorpb.DescriptorProto{
		Name: proto.String("FundSubInterface"),
		Field: []*descriptorpb.FieldDescriptorProto{
			stringField("type_id", 1),
			repeatedMessageField("subscriptions", 2, ".subscription_data.SubscriptionData"),
		},
	}

	// Pass messages in reverse dependency order to verify topo sort works
	output := buildSubscriptionOutput(t, dtMessages, []*descriptorpb.DescriptorProto{
		fundSubInterface,
		subscriptionData,
		subscriptionWorkflowData,
		individualInvestorInfo,
	})

	// --- 1. Dependency ordering ---
	// IndividualInvestorInfo and SubscriptionWorkflowData must precede SubscriptionData
	iiIdx := strings.Index(output, "export type IndividualInvestorInfo = {")
	wdIdx := strings.Index(output, "export type SubscriptionWorkflowData = {")
	sdIdx := strings.Index(output, "export type SubscriptionData = {")
	fsiIdx := strings.Index(output, "export type FundSubInterface = {")

	if iiIdx < 0 {
		t.Error("missing IndividualInvestorInfo type def")
	}
	if wdIdx < 0 {
		t.Error("missing SubscriptionWorkflowData type def")
	}
	if sdIdx < 0 {
		t.Error("missing SubscriptionData type def")
	}
	if fsiIdx < 0 {
		t.Error("missing FundSubInterface type def")
	}
	if iiIdx >= sdIdx {
		t.Error("IndividualInvestorInfo should appear before SubscriptionData (topo sort)")
	}
	if wdIdx >= sdIdx {
		t.Error("SubscriptionWorkflowData should appear before SubscriptionData (topo sort)")
	}
	if sdIdx >= fsiIdx {
		t.Error("SubscriptionData should appear before FundSubInterface (topo sort)")
	}

	// --- 2. Type def correctness ---
	// Message ref (DataTypes) = optional
	if !strings.Contains(output, "\tname: DataTypes.IndividualNameType?,") {
		t.Error("name should be DataTypes.IndividualNameType? (message ref = optional)")
	}
	// proto3 optional string
	if !strings.Contains(output, "\tinitials: string?,") {
		t.Error("initials should be string? (proto3 optional)")
	}
	// Required string — no ?
	if !strings.Contains(output, "\ttypeId: string,") {
		t.Error("typeId should be required string (no ?)")
	}
	// Repeated DataTypes ref — no ?
	if !strings.Contains(output, "\tprimaryContacts: {DataTypes.ContactInfoType},") {
		t.Error("primaryContacts should be {DataTypes.ContactInfoType} (repeated, no ?)")
	}
	// Repeated local ref — no ?
	if !strings.Contains(output, "\tsubscriptions: {SubscriptionData},") {
		t.Error("subscriptions should be {SubscriptionData} (repeated local, no ?)")
	}
	// Optional local ref — has ?
	if !strings.Contains(output, "\tindividualInvestor: IndividualInvestorInfo?,") {
		t.Error("individualInvestor should be IndividualInvestorInfo? (optional local ref)")
	}

	// --- 3. Constructor existence ---
	if !strings.Contains(output, "function SubscriptionData.makeIndividualInvestorInfo(") {
		t.Error("missing makeIndividualInvestorInfo constructor")
	}
	if !strings.Contains(output, "function SubscriptionData.makeSubscriptionWorkflowData(") {
		t.Error("missing makeSubscriptionWorkflowData constructor")
	}
	if !strings.Contains(output, "function SubscriptionData.makeSubscriptionData(") {
		t.Error("missing makeSubscriptionData constructor")
	}
	if !strings.Contains(output, "function SubscriptionData.makeFundSubInterface(") {
		t.Error("missing makeFundSubInterface constructor")
	}

	// --- 4. Codec existence ---
	if !strings.Contains(output, "function SubscriptionData.encodeJsonIndividualInvestorInfo(") {
		t.Error("missing encodeJsonIndividualInvestorInfo")
	}
	if !strings.Contains(output, "function SubscriptionData.decodeJsonIndividualInvestorInfo(") {
		t.Error("missing decodeJsonIndividualInvestorInfo")
	}
	if !strings.Contains(output, "function SubscriptionData.encodeJsonSubscriptionWorkflowData(") {
		t.Error("missing encodeJsonSubscriptionWorkflowData")
	}
	if !strings.Contains(output, "function SubscriptionData.decodeJsonSubscriptionWorkflowData(") {
		t.Error("missing decodeJsonSubscriptionWorkflowData")
	}
	if !strings.Contains(output, "function SubscriptionData.encodeJsonSubscriptionData(") {
		t.Error("missing encodeJsonSubscriptionData")
	}
	if !strings.Contains(output, "function SubscriptionData.decodeJsonSubscriptionData(") {
		t.Error("missing decodeJsonSubscriptionData")
	}
	if !strings.Contains(output, "function SubscriptionData.encodeJsonFundSubInterface(") {
		t.Error("missing encodeJsonFundSubInterface")
	}
	if !strings.Contains(output, "function SubscriptionData.decodeJsonFundSubInterface(") {
		t.Error("missing decodeJsonFundSubInterface")
	}

	// --- 5. Encode patterns ---
	// Repeated DataTypes ref iteration
	if !strings.Contains(output, "DataTypes.encodeJsonContactInfoType(item)") {
		t.Error("encode should iterate primaryContacts with DataTypes.encodeJsonContactInfoType(item)")
	}
	// Repeated local ref iteration
	if !strings.Contains(output, "SubscriptionData.encodeJsonSubscriptionData(item)") {
		t.Error("encode should iterate subscriptions with SubscriptionData.encodeJsonSubscriptionData(item)")
	}
	// Optional local ref encode — param name for SubscriptionData is "su" (s-collision avoidance)
	if !strings.Contains(output, "SubscriptionData.encodeJsonIndividualInvestorInfo(su.individualInvestor)") {
		t.Error("encode should call SubscriptionData.encodeJsonIndividualInvestorInfo(su.individualInvestor) for optional local ref")
	}

	// --- 6. Optional helpers ---
	if !strings.Contains(output, "local function decodeOptionalRadioGroupType(") {
		t.Error("missing decodeOptionalRadioGroupType helper")
	}
	if !strings.Contains(output, "local function encodeOptionalMoneyType(") {
		t.Error("missing encodeOptionalMoneyType helper")
	}

	// --- 7. Footer ---
	if !strings.Contains(output, "return SubscriptionData") {
		t.Error("missing 'return SubscriptionData' footer")
	}
}
