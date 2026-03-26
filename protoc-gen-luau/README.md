# protoc-gen-luau

A `protoc` compiler plugin that generates strictly-typed [Luau](https://luau-lang.org/) modules from Protocol Buffer definitions.

## Overview

`protoc-gen-luau` reads `.proto` files via protoc's standard plugin interface (`CodeGeneratorRequest` on stdin) and emits Luau modules containing:

- **Type definitions** (`export type`) with full Luau type annotations
- **Constructors** (`makeXxx`) with typed optional configuration
- **JSON codecs** (`encodeJsonXxx` / `decodeJsonXxx`) for serialization round-trips
- **Dispatch tables and functions** for runtime type-ID-based encoding/decoding (data_types only)

## Architecture

The plugin uses a **strategy pattern** with proto file-level options for dispatch. Each `.proto` file declares which generator handles it via a custom option:

```proto
import "luauoptions/luau_options.proto";
option (luau.luau_generator) = "schema";  // or "data_types"
```

Two generator strategies are registered:

| Strategy | Option Value | Handles |
|---|---|---|
| `DataTypesGenerator` | `"data_types"` | The foundational data_types module with compound pairs, dispatch tables, and FieldValue union |
| `SchemaGenerator` | `"schema"` | Any schema module built on data_types — tables, subscription data, fund interfaces, or any future shape |

Files without a `luau_generator` option are skipped (no output generated).

```
stdin (CodeGeneratorRequest)
  |
  v
main.go --- protogen.Options{}.Run(generate)
  |
  v
generate.go --- reads luau_generator option, looks up strategy in registry
  |
  +-- DataTypesGenerator (gen_data_types.go)
  |     classifyMessages() -> simpleTypes, compoundPairs,
  |     customCompound, oneofMsg
  |
  +-- SchemaGenerator (gen_schema.go + gen_schema_emitters.go)
  |     topoSortMessages() -> dependency-ordered emission
  |     generic per-message type defs, constructors, codecs
  |
  v
Shared emitters (called by both generators):
  +-- luau_types.go            export type blocks
  +-- luau_constructors.go     makeXxx functions
  +-- luau_codecs.go           encodeJson/decodeJson + optional helpers
  +-- luau_dispatch.go         dispatch tables + functions (data_types only)
  +-- luau_naming.go           naming conventions + type mapping
```

### Proto Options

Custom file-level options are defined in `luauoptions/luau_options.proto`:

| Option | Type | Default | Purpose |
|---|---|---|---|
| `luau_generator` | string | (none) | Which generator strategy to use: `"data_types"` or `"schema"` |
| `luau_data_types_require` | string | `"@lib/data_types"` | Override the `require()` path for data_types imports |

### SchemaGenerator

The `SchemaGenerator` handles any proto file with `option (luau.luau_generator) = "schema"`. It generates Luau types, constructors, and JSON codecs for every message in the file, supporting:

| Field Kind | Luau Type | Encode | Decode |
|---|---|---|---|
| Required scalar | `string` / `number` / `boolean` | direct | `or ""` / `or 0` / `if ~= nil` |
| Optional scalar (proto3 `optional`) | `type?` | nil check | pass through |
| Message ref (DataTypes) | `DataTypes.Type?` | optional helper | optional helper |
| Message ref (local) | `Type?` | `if ~= nil then Module.encode` | `if then Module.decode else nil` |
| Repeated scalar | `{type}` | direct | `or {}` |
| Repeated message (DataTypes) | `{DataTypes.Type}` | iterate + DataTypes codec | iterate + DataTypes codec |
| Repeated message (local) | `{Type}` | iterate + Module codec | iterate + Module codec |
| Map field | `{[string]: Type}` | direct | `or {}` |

Messages are emitted in **topological order** (dependencies before dependents) using Kahn's algorithm, so local cross-references resolve correctly.

**Required/optional rules** (proto3 semantics):
- Scalar without `optional` keyword: required (no `?`, has default)
- Scalar with `optional` keyword: optional (`?`, nil OK)
- Message fields: always optional (`?`) — proto3 messages are nullable
- Repeated/map fields: always present, default `{}`

### DataTypesGenerator

The `DataTypesGenerator` handles `data_types.proto` — the foundational module. It uses `classifyMessages()` to categorize proto messages:

- **Simple types**: `StringType`, `NumberType`, etc.
- **Compound pairs**: Matched `*SubFields` + `*Type` pairs (e.g., `AddressSubFields` + `AddressType`)
- **Custom compound**: Message with a map field (`CustomCompoundType`)
- **Oneof message**: Message with a non-synthetic oneof (used for dispatch detection)

It also emits dispatch tables (`STRING_TYPE_IDS`, `NUMBER_TYPE_IDS`, `ENUM_TYPE_IDS`, `COMPOUND_TYPE_CODECS`) and `decodeFieldValueByTypeId`/`encodeFieldValueByTypeId` functions.

## Generated Output

Every generated file follows this structure:

```lua
--!strict
local DataTypes = require("@lib/data_types")  -- schema files only
local ModuleName = {}

-- export type definitions
-- constructors (makeXxx)
-- local optional encode/decode helpers
-- JSON encode/decode functions
-- dispatch tables and functions (data_types only)

return ModuleName
```

The module name is derived from the proto filename: `source_table.proto` becomes `SourceTable`.

## Prerequisites

- **Go** 1.22 or later
- **protoc** 3.20 or later

## Quick Start

```bash
cd protoc-gen-luau
./generate.sh <proto_dir> [output_dir]
```

`generate.sh` builds the plugin binary and runs protoc in a single step.

### Build Only

```bash
cd protoc-gen-luau
go build -o protoc-gen-luau .
```

### Manual protoc Invocation

Proto files must import `luauoptions/luau_options.proto` and set the `luau_generator` option. Then:

```bash
protoc \
  --plugin=protoc-gen-luau=./protoc-gen-luau \
  --luau_out=./output \
  -I <proto_dir> \
  -I <path_to_luau_options_proto> \
  your_schema.proto
```

## Testing

### Run All Tests

```bash
cd protoc-gen-luau
go test -v ./...
```

The test suite uses a **replay testing pattern** — protoc is not required to run tests. A serialized `CodeGeneratorRequest` captured from a real protoc run is stored at `testdata/request.pb` and replayed during tests. Proto options are injected programmatically at test time.

### Test Categories

| Test | File | What It Verifies |
|---|---|---|
| `TestGenerate` | `generate_test.go` | 4 output files produced with correct filenames (with injected options) |
| `TestGenerateGolden` | `generate_test.go` | Byte-for-byte match against `testdata/golden/` |
| `TestGenerateWithOptions` | `generate_test.go` | Option-based dispatch routes `data_types` correctly |
| `TestGenerateUnknownOption` | `generate_test.go` | Unknown generator type returns error |
| `TestGenerateNoOption` | `generate_test.go` | Files without options are skipped |
| `TestDataTypesSimpleTypes` | `gen_data_types_test.go` | Simple type names, constructors, codecs present |
| `TestDataTypesCompoundTypes` | `gen_data_types_test.go` | Compound pairs, SubFields/Type wrappers present |
| `TestDataTypesDispatch` | `gen_data_types_test.go` | Dispatch tables and functions present |
| `TestDataTypesHeader` | `gen_data_types_test.go` | `--!strict` header and `return DataTypes` footer |
| `TestSchemaTypeDefsSimple` | `gen_schema_test.go` | Required/optional scalar field type definitions |
| `TestSchemaTypeDefsWithRefs` | `gen_schema_test.go` | Local + DataTypes message refs with topological ordering |
| `TestSchemaCodecsSimple` | `gen_schema_test.go` | Encode/decode for scalars, DataTypes refs, repeated strings |
| `TestSchemaCodecsComplex` | `gen_schema_test.go` | Encode/decode for local refs, repeated local/DataTypes refs |
| `TestSubscriptionSchemaGolden` | `gen_schema_test.go` | Full subscription domain: 4 messages, 3-level nesting, all field kinds |
| `TestTopoSortNoDeps` | `topo_sort_test.go` | Independent messages sort without error |
| `TestTopoSortLinearChain` | `topo_sort_test.go` | C->B->A chain produces correct order |
| `TestTopoSortDiamond` | `topo_sort_test.go` | Diamond dependency resolves correctly |
| `TestSnakeToCamel` | `luau_naming_test.go` | 9 table-driven naming conversion cases |

### Update Golden Files

After making codegen changes, regenerate the golden files:

```bash
UPDATE_GOLDEN=1 go test ./...
```

## Debug: Capture CodeGeneratorRequest

To capture a raw `CodeGeneratorRequest` for replay testing or debugging:

```bash
DUMP_REQUEST=testdata/request.pb protoc \
  --plugin=protoc-gen-luau=./protoc-gen-luau \
  --luau_out=/tmp \
  -I <proto_dir> \
  your_protos.proto
```

When `DUMP_REQUEST` is set, the plugin writes the serialized request to the specified path and returns an empty response (no files generated).

## Project Structure

```
protoc-gen-luau/
+-- main.go                   # Entrypoint: protogen.Options{}.Run + DUMP_REQUEST debug mode
+-- generate.go               # Dispatch loop: reads proto option, looks up generator in registry
+-- generator.go              # LuauGenerator interface, registry, option readers
+-- classify.go               # classifyMessages(): categorizes proto messages by structure
+-- gen_data_types.go         # DataTypesGenerator: orchestrator for data_types.proto
+-- gen_schema.go             # SchemaGenerator: orchestrator for any schema proto
+-- gen_schema_emitters.go    # Emitter functions for SchemaGenerator (type defs, constructors, codecs)
+-- topo_sort.go              # Topological sort for message dependency ordering
+-- luau_naming.go            # Naming utilities: snakeToCamel, type mapping, optionality, defaults
+-- luau_types.go             # Shared emitters: export type definitions
+-- luau_constructors.go      # Shared emitters: makeXxx constructor functions
+-- luau_codecs.go            # Shared emitters: encodeJson/decodeJson + optional helpers
+-- luau_dispatch.go          # Emits dispatch tables + dispatch functions (data_types only)
+-- type_id_tables.go         # Hardcoded string/number/enum type ID sets
+-- luauoptions/
|   +-- luau_options.proto    # Custom file-level proto options (luau_generator, luau_data_types_require)
|   +-- luau_options.pb.go    # Generated Go code for proto options
+-- generate_test.go          # Dispatch tests + golden replay test
+-- gen_data_types_test.go    # Structural tests for data_types.luau output
+-- gen_schema_test.go        # SchemaGenerator tests + subscription domain golden test
+-- topo_sort_test.go         # Unit tests for topological sort
+-- luau_naming_test.go       # Unit tests for snakeToCamel
+-- generate.sh               # Build + run protoc in one step
+-- go.mod                    # Single dependency: google.golang.org/protobuf
+-- go.sum
+-- testdata/
    +-- request.pb            # Captured CodeGeneratorRequest for replay tests
    +-- golden/               # Expected output for byte-for-byte comparison
        +-- data_types.luau
        +-- table_schema.luau
        +-- source_table.luau
        +-- target_table.luau
```

## Adding a New Schema

To generate Luau from a new `.proto` file:

1. Import the options proto and set the generator:
   ```proto
   syntax = "proto3";
   import "luauoptions/luau_options.proto";
   import "data_types.proto";
   option (luau.luau_generator) = "schema";
   ```

2. Define your messages referencing data_types as needed:
   ```proto
   message InvestorInfo {
     string type_id = 1;
     IndividualNameType name = 2;
     optional string initials = 3;
     repeated ContactInfoType contacts = 4;
   }
   ```

3. Run protoc — the SchemaGenerator will produce `investor_info.luau` with typed definitions, constructors, and JSON codecs.

No Go code changes required.

## Known Limitations

1. **Hardcoded type ID tables.** The 28 type IDs in `type_id_tables.go` are Go constants, not derived from the proto schema. Adding a new field type to the proto `oneof` requires a corresponding edit to this file.

2. **Hardcoded dispatch logic.** `luau_dispatch.go` contains hardcoded string literals (`"Boolean"`, `"MultipleCheckbox"`, etc.) in the dispatch chain. These are not derived from proto.

3. **No proto enum support.** Proto enums are not mapped to Luau string literal unions. Enum fields are treated as `any`.

4. **No proto oneof support in schemas.** Only `data_types.proto` uses oneof. The SchemaGenerator does not handle oneof in generic schema files.

5. **No binary protobuf codecs.** Only JSON encode/decode is generated. Binary protobuf serialization is not supported.

## Dependencies

| Dependency | Version | Purpose |
|---|---|---|
| `google.golang.org/protobuf` | v1.36.11 | Proto descriptor parsing, protogen plugin framework, proto extensions |
