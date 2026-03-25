# Table Schema Codegen Design

## Goal

Implement Luau code generation for `table_schema.proto`, `tables/source_table.proto`, and `tables/target_table.proto`, producing output functionally equivalent to the 3 hand-written modules (581 lines total).

## Decisions

| Decision | Choice | Rationale |
|----------|--------|-----------|
| Code reuse | Generalize existing emitters | Add typePrefix/moduleName params to existing functions; avoids duplication |
| Cross-file refs | Auto-detect via ParentFile | `field.Message.ParentFile == currentFile` determines local vs DataTypes prefix |
| table_schema.proto | Dedicated orchestrator | Unique patterns (oneof→FieldValue, map<string,SingleFieldType>, FieldGroup) |
| source/target tables | Shared orchestrator | Same pattern parameterized by filename/module name |
| Output fidelity | Functionally equivalent | Same as data-types-codegen decision |

## Files

### New Files

| File | Purpose |
|------|---------|
| `gen_table_schema.go` | Orchestrator for table_schema.proto |
| `gen_table_files.go` | Shared orchestrator for source_table.proto and target_table.proto |
| `gen_table_schema_test.go` | Structural tests for all 3 generated files |

### Modified Files

| File | Changes |
|------|---------|
| `generate.go` | Route 3 new proto files to their orchestrators |
| `luau_types.go` | Add typePrefix param to compound defs; add emitFieldsMapDef |
| `luau_constructors.go` | Add typePrefix/moduleName params; add emitFieldsMapConstructor |
| `luau_codecs.go` | Add typePrefix/moduleName params; add emitPerModuleOptionalHelpers, emitFieldsMapCodecs |

## Cross-File Type Reference Detection

When emitting a field's Luau type, the codegen checks whether the field's message is defined in the current proto file:

```go
func luauTypeRef(field *protogen.Field, currentFile *protogen.File) string {
    if field.Desc.Kind() != protoreflect.MessageKind {
        return protoFieldToLuauType(field) // scalar — no prefix needed
    }
    if field.Message.Desc.ParentFile() == currentFile.Desc {
        return field.Message.GoIdent.GoName // local: "LpSignatoryType"
    }
    return "DataTypes." + field.Message.GoIdent.GoName // imported: "DataTypes.StringType"
}
```

This replaces hardcoded prefix parameters for type references. The `moduleName` parameter is still needed for function names (e.g., `SourceTable.makeXxx` vs `DataTypes.makeXxx`).

## table_schema.proto Patterns

### SingleFieldType

Proto has a real oneof with 22 variants. The hand-written Luau flattens this to:

```luau
export type SingleFieldType = {
    value: DataTypes.FieldValue?,
    label: string,
}
```

The codegen detects the real oneof (via `oneof.Desc.IsSynthetic() == false`, per F007) and emits `DataTypes.FieldValue?` for the value field.

Encode uses `DataTypes.encodeFieldValueByTypeId(sft.value)`. Decode uses `DataTypes.decodeFieldValueByTypeId(json.value)`.

### FieldGroup

Simple struct, no message-typed fields:
```luau
export type FieldGroup = { label: string, startIdx: number, endIdx: number }
```

Standard constructor and inline encode/decode (no optional helpers needed).

### TableSchema

```luau
export type TableSchema = {
    fieldsMap: {[string]: SingleFieldType},
    fieldKeysInOrder: {string},
    label: string,
    groups: {FieldGroup},
}
```

Encode iterates `fieldsMap` map calling `encodeJsonSingleFieldType` per entry. Decode iterates `json.fieldsMap` map. `groups` is a repeated message encoded/decoded with inline `table.insert` loops.

### Emission Order

1. `--!strict` + `local DataTypes = require("@lib/data_types")` + `local TableSchema = {}`
2. SingleFieldType, FieldGroup, TableSchema type definitions
3. Constructors for all 3
4. SingleFieldType encode/decode (uses FieldValue dispatch)
5. TableSchema encode/decode (iterates map + groups)
6. `return TableSchema`

## source_table.proto / target_table.proto Patterns

### Local Compound Types

source_table defines LpSignatoryFields/LpSignatoryType and W9Fields/W9Type locally. target_table has none. The codegen classifies messages the same way as data_types — SubFields+wrapper pairs — but scoped to the current file. Compound emitters are reused with the `moduleName` parameter for function names.

### Typed FieldsMap

Both files have a `*FieldsMap` message where every field is a message type. In proto3, message-typed fields are implicitly nullable (zero value is absent). The codegen treats ALL message-typed FieldsMap fields as optional in Luau (`: MessageType?`), matching the hand-written pattern. Fields referencing DataTypes types get `DataTypes.` prefix; fields referencing local types get no prefix.

### Per-Module Optional Helpers

Each file generates local optional helpers ONLY for DataTypes base types actually used in its FieldsMap:

- source_table: StringType, MultipleCheckboxType (2 pairs)
- target_table: StringType, MultipleCheckboxType, RadioGroupType, MoneyType (4 pairs)

These helpers are named differently from data_types helpers: `decodeOptionalStringType` (not `decodeOptionalStringField`). The codegen scans the FieldsMap message to determine which base types are used.

### FieldsMap Encode/Decode

For each field in the FieldsMap:
- If field type is from DataTypes (base type): use the optional helper (`encodeOptionalStringType(fm.fieldName)`)
- If field type is a local compound: use conditional dispatch (`if fm.field then Module.encodeJsonXxxType(fm.field) end`)

### Schema Encode/Decode

Same pattern as TableSchema but with a typed `fieldsMap` (not a generic map):
- `fieldsMap` is optional — conditional encode/decode
- `fieldKeysInOrder` and `label` are direct

### Emission Order

1. `--!strict` + `local DataTypes = require("@lib/data_types")` + `local ModuleName = {}`
2. Local compound type definitions (if any)
3. FieldsMap type definition
4. Schema type definition
5. Constructors for all types
6. Per-module optional helpers
7. Local compound encode/decode (if any)
8. FieldsMap encode/decode
9. Schema encode/decode
10. `return ModuleName`

## Emitter Generalization

### Type definitions

`emitCompoundSubFieldsDef(g, msg)` and `emitCompoundWrapperDef(g, wrapper, sfName)` gain a `currentFile *protogen.File` parameter so `luauTypeRef` can determine prefixes. Existing data_types calls pass their own file (fields are always local, so behavior unchanged).

### Constructors

`emitCompoundSubFieldsConstructor` and `emitCompoundWrapperConstructor` gain a `moduleName string` parameter for the function prefix (e.g., `"SourceTable"` → `function SourceTable.makeXxx`). Existing data_types calls pass `"DataTypes"`.

### Codecs

Same pattern: `moduleName` for function names, `currentFile` for type references. The compound SubFields/wrapper encode/decode functions already work generically — they just need the correct module name prefix.

### New emitters

- `emitPerModuleOptionalHelpers(g, fieldsMapMsg, currentFile)` — scans FieldsMap fields; for each field whose message type is from a DataTypes import (not a local compound), adds that type to the helper set. Emits one encode+decode helper pair per unique DataTypes type.
- `emitFieldsMapDef(g, msg, currentFile)` — typed struct with all-optional fields
- `emitFieldsMapConstructor(g, msg, moduleName, currentFile)` — all fields passthrough
- `emitFieldsMapCodecs(g, msg, moduleName, currentFile, localCompounds)` — per-field encode/decode using helpers or conditional dispatch

## Testing

### Go unit tests

- **TestTableSchemaOutput** — structural checks: SingleFieldType, FieldGroup, TableSchema types; constructors; encode/decode with FieldValue dispatch
- **TestSourceTableOutput** — local compound types (LpSignatory, W9); FieldsMap; optional helpers (StringType, MultipleCheckbox); Schema
- **TestTargetTableOutput** — FieldsMap; optional helpers (StringType, MultipleCheckbox, RadioGroup, MoneyType); Schema; requires `DataTypes.` import

### Golden file updates

Update `testdata/golden/table_schema.luau`, `source_table.luau`, `target_table.luau` with generated output.

### Backward compatibility

All existing data_types tests must continue to pass. The emitter parameter changes are backward-compatible (data_types orchestrator passes its own file + "DataTypes" module name).

## Acceptance Criteria

1. `protoc --luau_out` generates all 3 files: table_schema.luau, source_table.luau, target_table.luau
2. Each file has `local DataTypes = require("@lib/data_types")` import
3. SingleFieldType uses `DataTypes.FieldValue?` (not oneof variants)
4. FieldsMap messages generate typed structs with per-field encode/decode
5. Optional helpers generated only for base types actually used in each module
6. All generated files are functionally equivalent to hand-written versions
