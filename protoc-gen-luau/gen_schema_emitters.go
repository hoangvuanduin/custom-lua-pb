package main

import (
	"sort"
	"strings"

	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// emitSchemaTypeDef emits a generic type definition for a message in a schema file.
func emitSchemaTypeDef(g *protogen.GeneratedFile, msg *protogen.Message, currentFile *protogen.File) {
	name := msg.GoIdent.GoName
	g.P("export type ", name, " = {")
	for _, field := range msg.Fields {
		luauName := snakeToCamel(string(field.Desc.Name()))
		luauType := luauTypeRefGeneric(field, currentFile)
		if isFieldOptional(field) {
			g.P("\t", luauName, ": ", luauType, "?,")
		} else {
			g.P("\t", luauName, ": ", luauType, ",")
		}
	}
	g.P("}")
	g.P()
}

// emitSchemaConstructor emits a generic constructor for a message in a schema file.
// All constructor params are optional (with ?), using the config-or-{} pattern.
func emitSchemaConstructor(g *protogen.GeneratedFile, msg *protogen.Message, moduleName string, currentFile *protogen.File) {
	name := msg.GoIdent.GoName
	g.P("function ", moduleName, ".make", name, "(config: {")
	for _, field := range msg.Fields {
		luauName := snakeToCamel(string(field.Desc.Name()))
		luauType := luauTypeRefGeneric(field, currentFile)
		g.P("\t", luauName, ": ", luauType, "?,")
	}
	g.P("}?): ", name)
	g.P("\tlocal c = config or {}")
	g.P("\treturn {")
	for _, field := range msg.Fields {
		luauName := snakeToCamel(string(field.Desc.Name()))
		dflt := luauDefaultValueGeneric(field)
		if field.Desc.Kind() == protoreflect.BoolKind && !field.Desc.HasOptionalKeyword() {
			g.P("\t\t", luauName, " = if c.", luauName, " ~= nil then c.", luauName, " else false,")
		} else if dflt == "" {
			g.P("\t\t", luauName, " = c.", luauName, ",")
		} else {
			g.P("\t\t", luauName, " = c.", luauName, " or ", dflt, ",")
		}
	}
	g.P("\t}")
	g.P("end")
	g.P()
}

// emitSchemaEncode emits the encode function for a message in a schema file.
func emitSchemaEncode(g *protogen.GeneratedFile, msg *protogen.Message, moduleName string, currentFile *protogen.File) {
	name := msg.GoIdent.GoName
	paramName := strings.ToLower(name[:1])
	// Avoid param name collision with Luau builtins
	if paramName == "t" || paramName == "s" {
		paramName = strings.ToLower(name[:2])
		if len(name) > 2 {
			paramName = strings.ToLower(name[:1]) + name[1:2]
		}
	}

	g.P("function ", moduleName, ".encodeJson", name, "(", paramName, ": ", name, "): {[string]: any}")

	// Classify fields into those needing pre-processing (maps, repeated messages)
	// and those that can go directly into the result literal or need post-processing
	var preProcess []*protogen.Field // fields needing loop before result
	var directFields []*protogen.Field
	var postProcess []*protogen.Field // optional fields needing if-then after result

	for _, field := range msg.Fields {
		if field.Desc.IsMap() && isMapValueMessage(field) {
			preProcess = append(preProcess, field)
		} else if field.Desc.IsList() && field.Desc.Kind() == protoreflect.MessageKind {
			preProcess = append(preProcess, field)
		} else if isFieldOptional(field) {
			postProcess = append(postProcess, field)
		} else {
			directFields = append(directFields, field)
		}
	}

	// Emit pre-processing loops
	for _, field := range preProcess {
		luauName := snakeToCamel(string(field.Desc.Name()))
		if field.Desc.IsMap() {
			valField := field.Message.Fields[1]
			valTypeName := valField.Message.GoIdent.GoName
			g.P("\tlocal ", luauName, "Encoded: {[string]: any} = {}")
			g.P("\tfor key, val in ", paramName, ".", luauName, " do")
			if isFieldFromDifferentFile(valField, currentFile) {
				if isFieldValueType(valField) {
					g.P("\t\t", luauName, "Encoded[key] = DataTypes.encodeFieldValueByTypeId(val)")
				} else {
					g.P("\t\t", luauName, "Encoded[key] = DataTypes.encodeJson", valTypeName, "(val)")
				}
			} else {
				g.P("\t\t", luauName, "Encoded[key] = ", moduleName, ".encodeJson", valTypeName, "(val)")
			}
			g.P("\tend")
		} else {
			// Repeated message
			elemTypeName := field.Message.GoIdent.GoName
			g.P("\tlocal ", luauName, "Encoded: {any} = {}")
			g.P("\tfor _, item in ", paramName, ".", luauName, " do")
			if isFieldFromDifferentFile(field, currentFile) {
				if isFieldValueType(field) {
					g.P("\t\ttable.insert(", luauName, "Encoded, DataTypes.encodeFieldValueByTypeId(item))")
				} else {
					g.P("\t\ttable.insert(", luauName, "Encoded, DataTypes.encodeJson", elemTypeName, "(item))")
				}
			} else {
				g.P("\t\ttable.insert(", luauName, "Encoded, ", moduleName, ".encodeJson", elemTypeName, "(item))")
			}
			g.P("\tend")
		}
	}

	// Need result variable if we have postProcess fields
	hasOptional := len(postProcess) > 0
	if hasOptional {
		g.P("\tlocal result: {[string]: any} = {")
	} else {
		g.P("\treturn {")
	}

	// Direct fields
	for _, field := range directFields {
		luauName := snakeToCamel(string(field.Desc.Name()))
		g.P("\t\t", luauName, " = ", paramName, ".", luauName, ",")
	}

	// Pre-processed fields (use encoded variable)
	for _, field := range preProcess {
		luauName := snakeToCamel(string(field.Desc.Name()))
		g.P("\t\t", luauName, " = ", luauName, "Encoded,")
	}

	g.P("\t}")

	// Post-process optional fields
	for _, field := range postProcess {
		luauName := snakeToCamel(string(field.Desc.Name()))
		if field.Desc.Kind() == protoreflect.MessageKind {
			if isFieldFromDifferentFile(field, currentFile) {
				if isFieldValueType(field) {
					g.P("\tif ", paramName, ".", luauName, " ~= nil then")
					g.P("\t\tresult.", luauName, " = DataTypes.encodeFieldValueByTypeId(", paramName, ".", luauName, ")")
					g.P("\tend")
				} else {
					helperName := field.Message.GoIdent.GoName
					g.P("\tresult.", luauName, " = encodeOptional", helperName, "(", paramName, ".", luauName, ")")
				}
			} else {
				typeName := field.Message.GoIdent.GoName
				g.P("\tif ", paramName, ".", luauName, " ~= nil then")
				g.P("\t\tresult.", luauName, " = ", moduleName, ".encodeJson", typeName, "(", paramName, ".", luauName, ")")
				g.P("\tend")
			}
		} else {
			// Optional scalar
			g.P("\tif ", paramName, ".", luauName, " ~= nil then")
			g.P("\t\tresult.", luauName, " = ", paramName, ".", luauName)
			g.P("\tend")
		}
	}

	if hasOptional {
		g.P("\treturn result")
	}

	g.P("end")
	g.P()
}

// emitSchemaDecode emits the decode function for a message in a schema file.
func emitSchemaDecode(g *protogen.GeneratedFile, msg *protogen.Message, moduleName string, currentFile *protogen.File) {
	name := msg.GoIdent.GoName

	g.P("function ", moduleName, ".decodeJson", name, "(json: {[string]: any}): ", name)

	// Pre-decode loops for maps and repeated messages
	var preDecodeFields []*protogen.Field
	for _, field := range msg.Fields {
		if field.Desc.IsMap() && isMapValueMessage(field) {
			preDecodeFields = append(preDecodeFields, field)
		} else if field.Desc.IsList() && field.Desc.Kind() == protoreflect.MessageKind {
			preDecodeFields = append(preDecodeFields, field)
		}
	}

	for _, field := range preDecodeFields {
		luauName := snakeToCamel(string(field.Desc.Name()))
		if field.Desc.IsMap() {
			valField := field.Message.Fields[1]
			valTypeName := valField.Message.GoIdent.GoName
			luauValType := luauTypeRefGeneric(valField, currentFile)
			g.P("\tlocal ", luauName, ": {[string]: ", luauValType, "} = {}")
			g.P("\tif json.", luauName, " ~= nil then")
			g.P("\t\tfor key, val in json.", luauName, " :: {[string]: any} do")
			if isFieldFromDifferentFile(valField, currentFile) {
				if isFieldValueType(valField) {
					g.P("\t\t\t", luauName, "[key] = DataTypes.decodeFieldValueByTypeId(val)")
				} else {
					g.P("\t\t\t", luauName, "[key] = DataTypes.decodeJson", valTypeName, "(val)")
				}
			} else {
				g.P("\t\t\t", luauName, "[key] = ", moduleName, ".decodeJson", valTypeName, "(val)")
			}
			g.P("\t\tend")
			g.P("\tend")
		} else {
			// Repeated message
			elemType := luauElementType(field, currentFile)
			elemTypeName := field.Message.GoIdent.GoName
			g.P("\tlocal ", luauName, ": {", elemType, "} = {}")
			g.P("\tif json.", luauName, " ~= nil then")
			g.P("\t\tfor _, item in json.", luauName, " :: {{[string]: any}} do")
			if isFieldFromDifferentFile(field, currentFile) {
				if isFieldValueType(field) {
					g.P("\t\t\ttable.insert(", luauName, ", DataTypes.decodeFieldValueByTypeId(item))")
				} else {
					g.P("\t\t\ttable.insert(", luauName, ", DataTypes.decodeJson", elemTypeName, "(item))")
				}
			} else {
				g.P("\t\t\ttable.insert(", luauName, ", ", moduleName, ".decodeJson", elemTypeName, "(item))")
			}
			g.P("\t\tend")
			g.P("\tend")
		}
	}

	g.P("\treturn {")
	for _, field := range msg.Fields {
		luauName := snakeToCamel(string(field.Desc.Name()))

		// Pre-decoded fields: use local variable
		if field.Desc.IsMap() && isMapValueMessage(field) {
			g.P("\t\t", luauName, " = ", luauName, ",")
			continue
		}
		if field.Desc.IsList() && field.Desc.Kind() == protoreflect.MessageKind {
			g.P("\t\t", luauName, " = ", luauName, ",")
			continue
		}

		if field.Desc.IsList() || field.Desc.IsMap() {
			// Repeated scalar or map of scalars — use or {}
			g.P("\t\t", luauName, " = json.", luauName, " or {},")
			continue
		}

		if field.Desc.Kind() == protoreflect.MessageKind {
			if isFieldFromDifferentFile(field, currentFile) {
				if isFieldValueType(field) {
					g.P("\t\t", luauName, " = if json.", luauName, " ~= nil then DataTypes.decodeFieldValueByTypeId(json.", luauName, ") else nil,")
				} else {
					helperName := field.Message.GoIdent.GoName
					g.P("\t\t", luauName, " = decodeOptional", helperName, "(json.", luauName, "),")
				}
			} else {
				typeName := field.Message.GoIdent.GoName
				g.P("\t\t", luauName, " = if json.", luauName, " then ", moduleName, ".decodeJson", typeName, "(json.", luauName, ") else nil,")
			}
			continue
		}

		// Scalar fields
		if field.Desc.HasOptionalKeyword() {
			g.P("\t\t", luauName, " = json.", luauName, ",")
			continue
		}

		if field.Desc.Kind() == protoreflect.BoolKind {
			g.P("\t\t", luauName, " = if json.", luauName, " ~= nil then json.", luauName, " else false,")
			continue
		}

		dflt := luauDefaultValueGeneric(field)
		if dflt != "" {
			g.P("\t\t", luauName, " = json.", luauName, " or ", dflt, ",")
		} else {
			g.P("\t\t", luauName, " = json.", luauName, ",")
		}
	}
	g.P("\t}")
	g.P("end")
	g.P()
}

// isMapValueMessage returns true if the map field's value type is a message.
func isMapValueMessage(field *protogen.Field) bool {
	if !field.Desc.IsMap() {
		return false
	}
	valField := field.Message.Fields[1]
	return valField.Desc.Kind() == protoreflect.MessageKind
}

// emitSchemaOptionalHelpers scans all generic messages for DataTypes-imported
// optional message fields and emits decode/encode helper functions for each unique type.
func emitSchemaOptionalHelpers(g *protogen.GeneratedFile, msgs []*protogen.Message, currentFile *protogen.File) {
	typeSet := make(map[string]bool)
	for _, msg := range msgs {
		for _, field := range msg.Fields {
			if !isFieldFromDifferentFile(field, currentFile) {
				continue
			}
			// Skip FieldValue — it uses dispatch functions
			if isFieldValueType(field) {
				continue
			}
			// Only for optional (non-repeated, non-map) message fields
			if field.Desc.IsList() || field.Desc.IsMap() {
				continue
			}
			typeSet[field.Message.GoIdent.GoName] = true
		}
	}

	// Sort for deterministic output
	typeNames := make([]string, 0, len(typeSet))
	for name := range typeSet {
		typeNames = append(typeNames, name)
	}
	sort.Strings(typeNames)

	for _, typeName := range typeNames {
		g.P("local function decodeOptional", typeName, "(json: any): DataTypes.", typeName, "?")
		g.P("\tif json == nil then return nil end")
		g.P("\treturn DataTypes.decodeJson", typeName, "(json)")
		g.P("end")
		g.P("local function encodeOptional", typeName, "(v: DataTypes.", typeName, "?): {[string]: any}?")
		g.P("\tif v == nil then return nil end")
		g.P("\treturn DataTypes.encodeJson", typeName, "(v)")
		g.P("end")
	}
	if len(typeNames) > 0 {
		g.P()
	}
}