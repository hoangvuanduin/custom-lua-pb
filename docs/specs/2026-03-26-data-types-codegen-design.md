# Data Types Codegen Design

## Goal

Implement Luau code generation for `data_types.proto`, producing output that is functionally equivalent to the hand-written `data_types.luau` (2055 lines) and passes all existing tests that depend on it.

## Decisions

| Decision | Choice | Rationale |
|----------|--------|-----------|
| Codegen approach | Go string builder via `g.P()` | Already established in scaffold; output is mechanical and repetitive |
| Output fidelity | Functionally equivalent | Same types, functions, behavior; whitespace/comments/ordering may differ |
| typeId source | Hardcoded in Go | Known stable set; proto schema doesn't encode these |
| Code organization | Split by concern | Separate Go files per output section; manageable file sizes |
| Optional field rules | Semantic (match hand-written) | Not pure proto3 rules; hardcoded for simple types, derived for compounds |

## File Structure

```
protoc-gen-luau/
├── generate.go              # Router: dispatch per proto file (modified)
├── gen_data_types.go        # Orchestrator: emit data_types.luau sections in order
├── luau_naming.go           # snake_case → camelCase, proto→Luau type mappings
├── luau_types.go            # Emit export type blocks
├── luau_constructors.go     # Emit make* constructor functions
├── luau_codecs.go           # Emit encodeJson*/decodeJson* functions + optional helpers
├── luau_dispatch.go         # Emit dispatch tables + dispatch functions
├── type_id_tables.go        # Hardcoded typeId→category mappings
├── gen_data_types_test.go   # Tests for data types codegen
├── luau_naming_test.go      # Tests for naming conversion
├── main.go                  # (existing, unchanged)
├── generate_test.go         # (existing, golden file updated)
└── testdata/...             # (existing)
```

## Naming & Type Mapping

### Field names: snake_case → camelCase

`type_id` → `typeId`, `format_patterns` → `formatPatterns`, `value_sub_fields` → `valueSubFields`

### Proto types → Luau types

| Proto type | Luau type | Constructor default |
|-----------|-----------|---------------------|
| `string` | `string` | `""` |
| `double` / `int32` | `number` | `0` |
| `bool` | `boolean` | `if c.value ~= nil then c.value else false` |
| `repeated string` | `{string}` | `{}` |
| Message field (in SubFields) | `<MessageType>?` | `nil` (passthrough) |
| `map<string, NonCustomFieldValue>` | `{[string]: FieldValue}?` | `nil` |

### Optional field rules

Simple types — hardcoded set:
- StringType: `regex` is optional
- NumberType: `minValue`, `maxValue` are optional
- All other simple type fields are required

Compound types — derived:
- All message-typed fields in SubFields → optional (`?`)
- In wrapper types: `valueSubFields` → optional, `label` → optional
- `typeId` → required, `subFieldKeysInOrder` → required

## Message Classification

The codegen classifies each proto message by structural inspection:

| Category | Detection rule | Messages |
|----------|---------------|----------|
| Simple type | Has `type_id` field, no `value_sub_fields`, name doesn't end with `SubFields` | StringType, NumberType, BooleanType, EnumType, MultipleCheckboxType, RadioGroupType |
| Compound SubFields | Name ends with `SubFields` | PhoneFaxSubFields, DateTimeSubFields, ... (15 messages) |
| Compound wrapper | Has `type_id` + `value_sub_fields` + `sub_field_keys_in_order`, name ends with `Type` | PhoneFaxType, DateTimeType, ... (15 messages) |
| Oneof dispatch | Has a non-synthetic oneof field | NonCustomFieldValue |
| Custom compound | Has `map<string, ...>` field | CustomCompoundType |

## Codegen Data Flow

`generate.go` detects `data_types.proto` by filename and calls `generateDataTypes(plugin, file)`.

`gen_data_types.go` orchestrates emission in this order:

1. **Header** — `--!strict`, `local DataTypes = {}`
2. **Simple type definitions** — `export type StringType = { ... }` for each simple message
3. **Simple constructors** — `function DataTypes.makeStringType(config)` for each
4. **Simple codecs** — `encodeJsonStringType` / `decodeJsonStringType` for each
5. **FieldValue type** — `export type FieldValue = { typeId: string, [string]: any }`. This is the duck-typed union derived from the `NonCustomFieldValue` oneof in proto — each variant is identified at runtime by its `typeId` field rather than by Luau union types.
6. **Optional helpers** — Exactly 3 pairs, one per base type referenced in compound SubFields: `encodeOptionalStringField`/`decodeOptionalStringField`, `encodeOptionalNumberField`/`decodeOptionalNumberField`, `encodeOptionalMultipleCheckboxField`/`decodeOptionalMultipleCheckboxField`. Derived by scanning which base message types appear as fields in SubFields messages.
7. **Compound types** — For each SubFields+Wrapper pair:
   - SubFields type definition
   - Wrapper type definition
   - SubFields constructor
   - Wrapper constructor
   - SubFields encode/decode
   - Wrapper encode/decode
8. **CustomCompoundType** — Type, constructor, encode/decode (uses FieldValue dispatch)
9. **Dispatch tables** — STRING_TYPE_IDS, NUMBER_TYPE_IDS, ENUM_TYPE_IDS, COMPOUND_TYPE_CODECS
10. **Dispatch functions** — `decodeFieldValueByTypeId`, `encodeFieldValueByTypeId`
11. **Footer** — `return DataTypes`

## Codegen Patterns

### Simple type definition

For StringType (proto: `string type_id, string value, string regex, repeated string format_patterns`):

```luau
export type StringType = {
    typeId: string,
    value: string,
    regex: string?,
    formatPatterns: {string},
}
```

Rule: Walk fields, map type, apply optional rule, convert name to camelCase.

### Simple constructor

```luau
function DataTypes.makeStringType(config: {
    typeId: string?,
    value: string?,
    regex: string?,
    formatPatterns: {string}?,
}?): StringType
    local c = config or {}
    return {
        typeId = c.typeId or "",
        value = c.value or "",
        regex = c.regex,
        formatPatterns = c.formatPatterns or {},
    }
end
```

Rule: Config param has ALL fields optional. Body uses `c.field or <default>` for required fields, bare `c.field` for optional fields. Boolean special case: `if c.value ~= nil then c.value else false`.

### Simple encode

Required fields go in the table literal. Optional fields use conditional assignment. Special case: `repeated` field `formatPatterns` uses `if #field > 0`.

```luau
function DataTypes.encodeJsonStringType(st: StringType): {[string]: any}
    local result: {[string]: any} = {
        typeId = st.typeId,
        value = st.value,
    }
    if st.regex ~= nil then
        result.regex = st.regex
    end
    if #st.formatPatterns > 0 then
        result.formatPatterns = st.formatPatterns
    end
    return result
end
```

### Simple decode

Required fields use `json.field or <default>`. Optional fields use bare `json.field`.

### Compound SubFields encode/decode

Each SubFields field calls the appropriate optional helper:

```luau
result.countryCode = encodeOptionalStringField(sf.countryCode)
```

The helper to use is determined by the field's message type: StringType → `encodeOptionalStringField`, NumberType → `encodeOptionalNumberField`, etc.

### Compound wrapper encode/decode

Fixed pattern: encode omits nil `label` and `valueSubFields`. Decode uses conditional for `valueSubFields`:

```luau
valueSubFields = if json.valueSubFields then DataTypes.decodeJsonPhoneFaxSubFields(json.valueSubFields) else nil,
```

### Dispatch tables

Hardcoded in `type_id_tables.go` as Go slices. The codegen emits these as **Luau boolean maps** for O(1) lookup:

Go source (used to drive emission):
```go
var stringTypeIDs = []string{"String", "Ssn", "Ein", "Aba", "Itin", "Swift", "Iban", "Giin", "UsZip", "Email", "PhoneFaxString", "IsoCountryCode", "IsoCurrencyCode", "CountryString", "StateProvince", "City", "NumberAndStreet", "PostalZipCode", "MoneyString", "DateTimeString", "TimeZoneOffset"}
var numberTypeIDs = []string{"Integer", "Float", "Percentage", "Year"}
var enumTypeIDs = []string{"ShareClass", "TransactionType", "SubscriptionStatus"}
```

Generated Luau output:
```luau
local STRING_TYPE_IDS: {[string]: boolean} = {
    String = true, Ssn = true, Ein = true, ...
}
```

### Dispatch functions

Fixed structure with cascading if/elseif:
1. STRING_TYPE_IDS → decodeJsonStringType
2. NUMBER_TYPE_IDS → decodeJsonNumberType
3. "Boolean" → decodeJsonBooleanType
4. ENUM_TYPE_IDS → decodeJsonEnumType
5. "MultipleCheckbox" → decodeJsonMultipleCheckboxType
6. "RadioGroup" → decodeJsonRadioGroupType
7. "CustomCompound" or "CustomCompoundType" → decodeJsonCustomCompoundType
8. COMPOUND_TYPE_CODECS table lookup — returns `{decode: string, encode: string}` (function name strings), then dynamically resolves: `local decoder = (DataTypes :: any)[codec.decode]`
9. Fallback: return as-is with `:: any` cast

## Testing

### Go unit tests

- **TestSnakeToCamel** — Naming conversion edge cases
- **TestMessageClassification** — Verify classifier buckets each data_types.proto message correctly
- **TestGenerateDataTypes** — Feed real captured request through codegen, verify structural correctness:
  - `--!strict` header, `return DataTypes` footer
  - All expected `export type` names present
  - All expected function names present (make*, encodeJson*, decodeJson*, dispatch functions)
  - Dispatch table names present

### Golden file update

Update `testdata/golden/data_types.luau` with the generated output. `TestGenerateGolden` validates that codegen output is stable across runs.

### Integration test (manual during development)

1. `./generate.sh <proto_dir> /tmp/gen-output`
2. Copy `data_types.luau` to `luau-data-transform/lib/`
3. Run Luau test suite — `test_protobuf_json.luau`, `test_protobuf_binary.luau`, `test_example_mappings.luau` must pass

## Scope Boundaries

This spec does NOT include:
- Generation for `table_schema.proto`, `source_table.proto`, or `target_table.proto` (deferred to table-schema-codegen)
- The 3 other placeholder `.luau` files remain placeholders
- Runtime Luau code changes (tests are the acceptance gate, not the codegen target)

## Acceptance Criteria

1. `protoc --luau_out` generates a non-placeholder `data_types.luau`
2. Generated output includes all exported types matching hand-written signatures
3. Generated constructors have matching parameter types and default values
4. Generated JSON encode/decode functions handle all optional field patterns correctly
5. typeId dispatch tables and functions are generated from the oneof variants
6. `--!strict` header and `return DataTypes` footer present
7. Generated file passes `luau-analyze` without errors (if available; otherwise verified by test suite passing)
