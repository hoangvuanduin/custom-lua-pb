#!/usr/bin/env bash
set -euo pipefail

PROTO_DIR="${1:?Usage: generate.sh <proto_dir> [output_dir]}"
OUTPUT_DIR="${2:-.}"

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
cd "$SCRIPT_DIR"
go build -o protoc-gen-luau .

protoc \
  --plugin=protoc-gen-luau="$SCRIPT_DIR/protoc-gen-luau" \
  "--luau_out=Mdata_types.proto=placeholder/data_types,Mtable_schema.proto=placeholder/table_schema,Mtables/source_table.proto=placeholder/tables/source_table,Mtables/target_table.proto=placeholder/tables/target_table:$OUTPUT_DIR" \
  -I "$PROTO_DIR" \
  data_types.proto table_schema.proto \
  tables/source_table.proto tables/target_table.proto
