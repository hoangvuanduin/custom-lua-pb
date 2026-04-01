#!/usr/bin/env bash
set -euo pipefail

PROTO_DIR="${1:?Usage: generate.sh <proto_dir> [output_dir]}"
OUTPUT_DIR="${2:-.}"

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
cd "$SCRIPT_DIR"
go build -o protoc-gen-luau .

protoc \
  --plugin=protoc-gen-luau="$SCRIPT_DIR/protoc-gen-luau" \
  "--luau_out=Mdata_types.proto=placeholder/data_types,Mtable_schema.proto=placeholder/table_schema,Mtables/source_table.proto=placeholder/tables/source_table,Mtables/target_table.proto=placeholder/tables/target_table,Mtables/source_contact_table.proto=placeholder/tables/source_contact_table,Mtables/source_w9_table.proto=placeholder/tables/source_w9_table,Mtables/source_feeder_table.proto=placeholder/tables/source_feeder_table,Mtables/source_master_table.proto=placeholder/tables/source_master_table,Mtables/full_target_table.proto=placeholder/tables/full_target_table:$OUTPUT_DIR" \
  -I "$PROTO_DIR" \
  -I "$SCRIPT_DIR" \
  data_types.proto table_schema.proto \
  tables/source_table.proto tables/target_table.proto \
  tables/source_contact_table.proto tables/source_w9_table.proto \
  tables/source_feeder_table.proto tables/source_master_table.proto \
  tables/full_target_table.proto

# --- Mapping protos (optional) ---
MAPPINGS_DIR="$PROTO_DIR/mappings"
if [ -d "$MAPPINGS_DIR" ]; then
  MAPPING_PROTOS=()
  MAPPING_MFLAGS="Mdata_types.proto=placeholder/data_types"
  for f in "$MAPPINGS_DIR"/*.proto; do
    [ -f "$f" ] || continue
    rel="mappings/$(basename "$f")"
    MAPPING_PROTOS+=("$rel")
    MAPPING_MFLAGS="$MAPPING_MFLAGS,M${rel}=placeholder/${rel%.proto}"
  done
  if [ ${#MAPPING_PROTOS[@]} -gt 0 ]; then
    protoc \
      --plugin=protoc-gen-luau="$SCRIPT_DIR/protoc-gen-luau" \
      "--luau_out=${MAPPING_MFLAGS}:$OUTPUT_DIR" \
      -I "$PROTO_DIR" \
      -I "$SCRIPT_DIR" \
      "${MAPPING_PROTOS[@]}"
  fi
fi
