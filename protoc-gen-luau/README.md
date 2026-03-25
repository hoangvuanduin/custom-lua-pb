# protoc-gen-luau

A protoc plugin that generates Luau type definition modules from `.proto` files.

## Prerequisites

- Go 1.22+
- protoc 3.20+

## Build

```bash
cd protoc-gen-luau
go build -o protoc-gen-luau .
```

## Usage

### Via generate.sh

```bash
./generate.sh <proto_dir> [output_dir]
```

Example:
```bash
./generate.sh ../luau-data-transform/proto ./output
```

### Manual protoc invocation

```bash
protoc \
  --plugin=protoc-gen-luau=./protoc-gen-luau \
  "--luau_out=Mdata_types.proto=placeholder/data_types,Mtable_schema.proto=placeholder/table_schema,Mtables/source_table.proto=placeholder/tables/source_table,Mtables/target_table.proto=placeholder/tables/target_table:./output" \
  -I <proto_dir> \
  data_types.proto table_schema.proto \
  tables/source_table.proto tables/target_table.proto
```

Note: The `M` parameters provide placeholder Go import paths required by the `protogen` framework. They are not used by the Luau code generator.

## Testing

```bash
go test -v ./...
```

## Debug: Capture CodeGeneratorRequest

To capture a raw CodeGeneratorRequest for replay testing:

```bash
DUMP_REQUEST=testdata/request.pb protoc \
  --plugin=protoc-gen-luau=./protoc-gen-luau \
  "--luau_out=Mdata_types.proto=placeholder/data_types,Mtable_schema.proto=placeholder/table_schema,Mtables/source_table.proto=placeholder/tables/source_table,Mtables/target_table.proto=placeholder/tables/target_table:/tmp" \
  -I <proto_dir> \
  data_types.proto table_schema.proto \
  tables/source_table.proto tables/target_table.proto
```
