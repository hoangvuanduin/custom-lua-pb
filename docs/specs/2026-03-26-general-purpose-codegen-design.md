# General-Purpose Proto-to-Luau Codegen Design

## Goal

Refactor protoc-gen-luau from a hardcoded 4-module generator into an extensible, general-purpose proto-to-Luau code generator. The `data_types` module (common types) remains a dedicated generator. All other schemas — tables, subscription data, fund interfaces, or any future shape — use a single generic `SchemaGenerator` driven by proto file options.

## Context

### Current State

The plugin generates 4 Luau modules from 4 hardcoded proto file basenames:

```
data_types.proto   → data_types.luau    (DataTypesGenerator — classifier, dispatch tables, oneof)
table_schema.proto → table_schema.luau  (hardcoded SingleFieldType/FieldGroup/TableSchema shapes)
source_table.proto → source_table.luau  (suffix-matching for *FieldsMap/*Schema messages)
target_table.proto → target_table.luau  (same as source_table)
```

The dispatcher in `generate.go` is a flat if/else chain on basename strings. Adding any new schema requires editing the dispatcher and writing a new generator function.

### Why Generalize

The `data_types` library defines stable common types (StringType, IndividualName, Address, etc.). Multiple structurally different schemas are built on top of these types:

- **Table schemas** — FieldsMap + Schema pattern (source_table, target_table)
- **Domain schemas** — trees of typed messages (SubscriptionData, FundSubInterface, IndividualInvestorInfo)
- **Future schemas** — unknown shapes that reference data_types

All share the same codegen needs: type definitions, constructors, JSON encode/decode. The difference is message structure, not codegen mechanics.

## Architecture

### Strategy Pattern with Proto Option Dispatch

Each proto file declares its codegen strategy via a custom file-level option. A registry maps option values to `LuauGenerator` implementations.

```
luau_options.proto defines:
  luau_generator      → "data_types" | "schema"
  luau_data_types_require → override require path (default: "@lib/data_types")

Generator registry:
  "data_types" → DataTypesGenerator (unchanged, dedicated)
  "schema"     → SchemaGenerator (new, general-purpose)
```

### Proto Options

`luau_options.proto`:

```proto
syntax = "proto3";
package luau;
option go_package = "protoc-gen-luau/luauoptions";
import "google/protobuf/descriptor.proto";

extend google.protobuf.FileOptions {
  string luau_generator = 51000;
  string luau_data_types_require = 51001;
}
```

Any proto file that wants Luau codegen imports this and sets `option (luau.luau_generator) = "schema";`.

### Generator Interface

```go
type LuauGenerator interface {
    Generate(plugin *protogen.Plugin, file *protogen.File)
}

var generators = map[string]LuauGenerator{
    "data_types": &DataTypesGenerator{},
    "schema":     &SchemaGenerator{},
}
```

### Dispatch (revised `generate.go`)

```go
func generate(plugin *protogen.Plugin) error {
    for _, f := range plugin.Files {
        if !f.Generate { continue }

        genType := getFileOption(f, luauoptions.E_LuauGenerator)
        if gen, ok := generators[genType]; ok {
            gen.Generate(plugin, f)
        } else if genType != "" {
            return fmt.Errorf("unknown luau_generator %q in %s", genType, f.Desc.Path())
        }
        // No option = skip
    }
    return nil
}
```

## SchemaGenerator Design

### What It Replaces

The SchemaGenerator subsumes both `gen_table_schema.go` (hardcoded type shapes) and `gen_table_files.go` (FieldsMap/Schema suffix matching). Both are deleted.

### Algorithm

For any proto file with `option (luau.luau_generator) = "schema"`:

1. **Classify messages** — reuse `classifyMessages()` to detect compound pairs, simple types, custom compounds, oneof messages
2. **Build local dependency graph** — for each message, record which other local messages it references via fields
3. **Topological sort** — emit messages in dependency order (dependencies before dependents)
4. **Header** — emit `--!strict`, `require()` for data_types (and any cross-schema imports), module table declaration
5. **For each message** (in sorted order), emit:
   - Type definition
   - Constructor
   - JSON encoder
   - JSON decoder
6. **Compound pairs** — for detected compound SubFields/Wrapper pairs, use existing compound emitters
7. **Per-module optional helpers** — for DataTypes-imported types used across messages, emit `decodeOptional*`/`encodeOptional*` helpers
8. **Footer** — `return Module`

### Field Classification & Codec Dispatch

Each field is classified and gets appropriate codegen:

| Field kind | Proto example | Luau type | Encode | Decode |
|---|---|---|---|---|
| Scalar string | `string name = 1;` | `string` | direct | direct |
| Scalar int | `int32 share_unit = 1;` | `number` | direct | direct |
| Scalar bool | `bool active = 1;` | `boolean` | direct | direct |
| DataTypes ref | `IndividualName name = 1;` | `DataTypes.IndividualName` | `DataTypes.encodeJson*()` | `DataTypes.decodeJson*()` |
| Local msg ref | `IndividualInvestorInfo inv = 1;` | `IndividualInvestorInfo` | `Module.encodeJson*()` | `Module.decodeJson*()` |
| Repeated scalar | `repeated string tags = 1;` | `{string}` | direct (array) | direct (array) |
| Repeated DataTypes ref | `repeated Signatory signers = 1;` | `{DataTypes.Signatory}` | iterate + DataTypes codec | iterate + DataTypes codec |
| Repeated local ref | `repeated SubscriptionData subs = 1;` | `{SubscriptionData}` | iterate + local codec | iterate + local codec |
| Map field | `map<string, X> m = 1;` | `{[string]: X}` | iterate entries | iterate entries |

### Required vs Optional (proto3 semantics)

- Field **without** `optional` keyword: required in Luau — no `?` suffix, no default in constructor, must be provided
- Field **with** `optional` keyword: optional in Luau — `?` suffix, default value in constructor (`""`, `0`, `false`, `nil`, `{}`)
- `repeated` fields: always present, default to `{}`

### `typeId` Handling

The `type_id` field that appears on domain messages (e.g., `string type_id = 1;`) is treated as a regular string field. The generator does not auto-emit or treat it specially. Its default value in the constructor is `""` (standard string default). If a specific default is needed, that's a future enhancement via field-level options.

### Cross-Schema Imports

When a schema proto imports another schema proto (not data_types), the generator:

1. Detects the import via proto file dependencies
2. Emits a `require()` for the imported module (derive path from proto file path)
3. Prefixes types from the imported module with its module name (e.g., `OtherModule.SomeType`)

### Topological Sort

Messages within a file are sorted by dependency:

```go
func topologicalSort(messages []*protogen.Message, file *protogen.File) []*protogen.Message {
    // Build adjacency: msg A depends on msg B if A has a field of type B (same file)
    // Kahn's algorithm or DFS-based topo sort
    // Cycle detection: error if circular references found
}
```

## File Changes

| File | Action | Notes |
|---|---|---|
| `luau_options.proto` | NEW | Custom file-level options |
| `luau_options.pb.go` | NEW | Generated from luau_options.proto |
| `generator.go` | NEW | LuauGenerator interface + registry |
| `generate.go` | MODIFIED | Registry dispatch replaces if/else chain |
| `gen_data_types.go` | MINOR | Wrap in DataTypesGenerator struct, implement interface |
| `gen_schema.go` | NEW | SchemaGenerator — replaces gen_table_schema.go + gen_table_files.go |
| `gen_table_schema.go` | DELETED | Absorbed into SchemaGenerator |
| `gen_table_files.go` | DELETED | Absorbed into SchemaGenerator |
| `luau_types.go` | MINOR | Add repeated field type emission |
| `luau_constructors.go` | MINOR | Add repeated field constructor support |
| `luau_codecs.go` | MINOR | Add repeated + scalar field codec support |
| `luau_naming.go` | MINOR | Add repeated field utilities, remove hardcoded `simpleOptionalFields` map |

### Shared Emitter Toolkit (existing, reused)

These files remain as a composable toolkit both generators use:

- `luau_types.go` — compound SubFields/Wrapper type definitions
- `luau_constructors.go` — compound constructors
- `luau_codecs.go` — compound encode/decode
- `luau_naming.go` — `snakeToCamel`, `luauTypeRef`, `luauDefaultValue`, `isFieldFromDifferentFile`
- `luau_dispatch.go` — data_types specific (dispatch tables, FieldValue codec)
- `type_id_tables.go` — data_types specific (type ID sets)

### What Stays Unchanged

- `gen_data_types.go` logic (wrapped in struct but same algorithm)
- `luau_dispatch.go` (data_types specific)
- `type_id_tables.go` (data_types specific)
- `classifyMessages()` (reused by both generators)
- Golden file test mechanism (new golden files added for new schemas)
- `main.go` (calls `generate()` which now uses registry)

## Proto Migration

All proto files add the option import and annotation. `data_types.proto` uses `"data_types"`; all other schema files use `"schema"`:

```proto
// data_types.proto — add option
syntax = "proto3";
import "luau_options.proto";
option (luau.luau_generator) = "data_types";
// ... messages unchanged ...

// source_table.proto — add option
syntax = "proto3";
import "luau_options.proto";
import "data_types.proto";
option (luau.luau_generator) = "schema";
// ... messages unchanged ...

// table_schema.proto — add option
syntax = "proto3";
import "luau_options.proto";
import "data_types.proto";
option (luau.luau_generator) = "schema";
// ... messages unchanged ...
```

New schemas (e.g., subscription_data.proto):

```proto
syntax = "proto3";
import "luau_options.proto";
import "data_types.proto";
option (luau.luau_generator) = "schema";

message IndividualInvestorInfo {
  string type_id = 1;
  IndividualName name = 2;
  optional string initials = 3;
  optional DateTimeString date_of_birth = 4;
  optional string nationality = 5;
  // ...
}

message SubscriptionData {
  string type_id = 1;
  optional RadioGroup investor_type = 2;
  optional IndividualInvestorInfo individual_investor = 3;
  repeated ContactInfo primary_contacts = 4;
  repeated Signatory lp_signers = 5;
  // ...
}
```

## Testing Strategy

1. **Existing golden tests** — update `testdata/request.pb` to include `luau_options` on existing proto files. Golden output should be identical (proves no regression).
2. **Subscription schema golden test** — add a full test proto based on the subscription/fund domain example that exercises all field kinds the SchemaGenerator handles. This proto must include:
   - Messages with local cross-references (e.g., `SubscriptionData` → `IndividualInvestorInfo`)
   - Required fields (no `optional` keyword): `string type_id`, `IndividualName name`
   - Optional scalar fields: `optional string initials`, `optional int32 share_unit`
   - DataTypes references (optional + required): `optional RadioGroup investor_type`, `IndividualName name`
   - Repeated DataTypes references: `repeated Signatory lp_signers`, `repeated ContactInfo primary_contacts`
   - Repeated local message references: `repeated SubscriptionData subscriptions`
   - Nested anonymous-like structs (proto nested messages): e.g., `FundInfo.KeyLabel` with `string key`, `optional string label`
   - Repeated nested messages: `repeated KeyLabel communication_types`
   - Multi-level nesting: `FundSubInterface` → `SubscriptionData` → `IndividualInvestorInfo` (3+ levels, tests topological sort)
   - A `typeId` string discriminator field on each message (treated as regular field)
   - Verify golden Luau output has correct dependency ordering, type refs, codecs
3. **Unit tests** — test topological sort, field classification, repeated field handling in isolation.
4. **Integration** — generated Luau modules should pass `luau-analyze` strict mode type checking.

## Scope Boundaries

### In scope
- Proto file-level options for generator dispatch
- SchemaGenerator handling all field kinds (scalar, message ref, repeated, map)
- Required/optional via proto3 `optional` keyword
- Topological sort of messages within a file
- Cross-file imports (data_types + other schema modules)
- Backward compatibility (existing table protos produce same Luau output with new annotations)

### Out of scope (future enhancements)
- Field-level options (custom defaults, codec overrides)
- Message-level options (auto-typeId generation)
- Proto enum → Luau string literal union codegen
- Binary protobuf encode/decode (current scope is JSON only)
- Proto `oneof` in schema files (only data_types uses oneof currently)
