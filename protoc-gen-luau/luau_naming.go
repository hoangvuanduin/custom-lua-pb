package main

import (
	"path"
	"strings"
	"unicode"

	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// snakeToCamel converts snake_case to camelCase.
func snakeToCamel(s string) string {
	parts := strings.Split(s, "_")
	for i := 1; i < len(parts); i++ {
		if len(parts[i]) > 0 {
			runes := []rune(parts[i])
			runes[0] = unicode.ToUpper(runes[0])
			parts[i] = string(runes)
		}
	}
	return strings.Join(parts, "")
}

// protoFieldToLuauType returns the Luau type string for a proto field.
func protoFieldToLuauType(field *protogen.Field) string {
	if field.Desc.IsMap() {
		return "{[string]: FieldValue}"
	}
	if field.Desc.IsList() {
		switch field.Desc.Kind() {
		case protoreflect.StringKind:
			return "{string}"
		default:
			return "{any}"
		}
	}
	switch field.Desc.Kind() {
	case protoreflect.StringKind:
		return "string"
	case protoreflect.DoubleKind, protoreflect.Int32Kind, protoreflect.FloatKind, protoreflect.Int64Kind:
		return "number"
	case protoreflect.BoolKind:
		return "boolean"
	case protoreflect.MessageKind:
		return field.Message.GoIdent.GoName
	default:
		return "any"
	}
}

// luauTypeRef returns the Luau type reference for a proto field, prefixing with "DataTypes."
// when the field's message type lives in a different file than currentFile.
func luauTypeRef(field *protogen.Field, currentFile *protogen.File) string {
	if field.Desc.Kind() != protoreflect.MessageKind {
		return protoFieldToLuauType(field)
	}
	if field.Message.Desc.ParentFile().Path() == currentFile.Desc.Path() {
		return field.Message.GoIdent.GoName
	}
	return "DataTypes." + field.Message.GoIdent.GoName
}

// luauDefaultValue returns the Luau default for a proto field used in constructors.
// Returns "" (empty) for fields that should default to nil (no `or` clause).
func luauDefaultValue(field *protogen.Field) string {
	if field.Desc.Kind() == protoreflect.MessageKind {
		return "" // nil, no default
	}
	if field.Desc.IsMap() {
		return "" // nil
	}
	if field.Desc.IsList() {
		return "{}"
	}
	switch field.Desc.Kind() {
	case protoreflect.StringKind:
		return `""`
	case protoreflect.DoubleKind, protoreflect.Int32Kind, protoreflect.FloatKind, protoreflect.Int64Kind:
		return "0"
	case protoreflect.BoolKind:
		return "" // special handling: if c.value ~= nil then c.value else false
	default:
		return `""`
	}
}

// simpleOptionalFields defines which fields in simple types are optional (get `?` in type def).
var simpleOptionalFields = map[string]map[string]bool{
	"StringType": {"regex": true},
	"NumberType": {"min_value": true, "max_value": true},
}

// isSimpleOptionalField returns true if this field should be optional in the Luau type.
func isSimpleOptionalField(msgName string, field *protogen.Field) bool {
	if fields, ok := simpleOptionalFields[msgName]; ok {
		return fields[string(field.Desc.Name())]
	}
	return false
}

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
// e.g. "tables/source_table.proto" → "SourceTable"
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
