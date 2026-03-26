package main

import (
	"path"
	"sort"
	"strings"

	"google.golang.org/protobuf/compiler/protogen"
)

// generateTableFile generates a Luau module for a source_table or target_table proto file.
func generateTableFile(plugin *protogen.Plugin, file *protogen.File) {
	protoPath := file.Desc.Path()
	baseName := strings.TrimSuffix(path.Base(protoPath), ".proto")
	luauName := baseName + ".luau"
	moduleName := deriveModuleName(protoPath)

	g := plugin.NewGeneratedFile(luauName, "")
	cm := classifyMessages(file)

	// Find FieldsMap and Schema messages from simpleTypes
	var fieldsMapMsg, schemaMsg *protogen.Message
	for _, msg := range cm.simpleTypes {
		name := msg.GoIdent.GoName
		if strings.HasSuffix(name, "FieldsMap") {
			fieldsMapMsg = msg
		} else if strings.HasSuffix(name, "Schema") {
			schemaMsg = msg
		}
	}

	// Header
	g.P("--!strict")
	g.P(`local DataTypes = require("@lib/data_types")`)
	g.P()
	g.P("local ", moduleName, " = {}")
	g.P()

	// Local compound type definitions
	for _, pair := range cm.compoundPairs {
		sfName := pair.subFields.GoIdent.GoName
		emitCompoundSubFieldsDef(g, pair.subFields, file)
		g.P()
		emitCompoundWrapperDef(g, pair.wrapper, sfName)
		g.P()
	}

	// FieldsMap type definition (all fields optional)
	if fieldsMapMsg != nil {
		emitFieldsMapDef(g, fieldsMapMsg, file)
	}

	// Schema type definition
	if schemaMsg != nil && fieldsMapMsg != nil {
		emitTableFileSchemaTypeDef(g, schemaMsg, fieldsMapMsg)
	}

	// Local compound constructors
	for _, pair := range cm.compoundPairs {
		sfName := pair.subFields.GoIdent.GoName
		emitCompoundSubFieldsConstructor(g, pair.subFields, moduleName, file)
		emitCompoundWrapperConstructor(g, pair.wrapper, sfName, moduleName)
	}

	// FieldsMap constructor
	if fieldsMapMsg != nil {
		emitFieldsMapConstructor(g, fieldsMapMsg, moduleName, file)
	}

	// Schema constructor
	if schemaMsg != nil && fieldsMapMsg != nil {
		emitTableFileSchemaConstructor(g, schemaMsg, moduleName, fieldsMapMsg)
	}

	// Per-module optional helpers (for DataTypes types used in FieldsMap)
	if fieldsMapMsg != nil {
		emitPerModuleOptionalHelpers(g, fieldsMapMsg, file)
	}

	// Local compound codecs (use table-file helper naming: GoName instead of shortName+Field)
	for _, pair := range cm.compoundPairs {
		sfName := pair.subFields.GoIdent.GoName
		emitCompoundSubFieldsEncodeWith(g, pair.subFields, moduleName, tableFileHelperName)
		emitCompoundSubFieldsDecodeWith(g, pair.subFields, moduleName, tableFileHelperName)
		emitCompoundWrapperEncode(g, pair.wrapper, sfName, moduleName)
		emitCompoundWrapperDecode(g, pair.wrapper, sfName, moduleName)
	}

	// FieldsMap encode/decode
	if fieldsMapMsg != nil {
		emitFieldsMapEncode(g, fieldsMapMsg, moduleName, file)
		emitFieldsMapDecode(g, fieldsMapMsg, moduleName, file)
	}

	// Schema encode/decode
	if schemaMsg != nil && fieldsMapMsg != nil {
		emitTableFileSchemaEncode(g, schemaMsg, moduleName, fieldsMapMsg)
		emitTableFileSchemaDecode(g, schemaMsg, moduleName, fieldsMapMsg)
	}

	// Footer
	g.P("return ", moduleName)
}

// emitFieldsMapDef emits the typed struct where all fields are optional.
func emitFieldsMapDef(g *protogen.GeneratedFile, msg *protogen.Message, currentFile *protogen.File) {
	name := msg.GoIdent.GoName
	g.P("export type ", name, " = {")
	for _, field := range msg.Fields {
		luauName := snakeToCamel(string(field.Desc.Name()))
		luauType := luauTypeRef(field, currentFile)
		g.P("\t", luauName, ": ", luauType, "?,")
	}
	g.P("}")
	g.P()
}

// emitFieldsMapConstructor emits the make function for a FieldsMap message.
func emitFieldsMapConstructor(g *protogen.GeneratedFile, msg *protogen.Message, moduleName string, currentFile *protogen.File) {
	name := msg.GoIdent.GoName
	g.P("function ", moduleName, ".make", name, "(config: {")
	for _, field := range msg.Fields {
		luauName := snakeToCamel(string(field.Desc.Name()))
		luauType := luauTypeRef(field, currentFile)
		g.P("\t", luauName, ": ", luauType, "?,")
	}
	g.P("}?): ", name)
	g.P("\tlocal c = config or {}")
	g.P("\treturn {")
	for _, field := range msg.Fields {
		luauName := snakeToCamel(string(field.Desc.Name()))
		g.P("\t\t", luauName, " = c.", luauName, ",")
	}
	g.P("\t}")
	g.P("end")
	g.P()
}

// emitTableFileSchemaTypeDef emits the Schema type with optional fieldsMap.
func emitTableFileSchemaTypeDef(g *protogen.GeneratedFile, schema *protogen.Message, fieldsMap *protogen.Message) {
	schemaName := schema.GoIdent.GoName
	fieldsMapName := fieldsMap.GoIdent.GoName
	g.P("export type ", schemaName, " = {")
	g.P("\tfieldsMap: ", fieldsMapName, "?,")
	g.P("\tfieldKeysInOrder: {string},")
	g.P("\tlabel: string,")
	g.P("}")
	g.P()
}

// emitTableFileSchemaConstructor emits the make function for a Schema message.
func emitTableFileSchemaConstructor(g *protogen.GeneratedFile, schema *protogen.Message, moduleName string, fieldsMap *protogen.Message) {
	schemaName := schema.GoIdent.GoName
	fieldsMapName := fieldsMap.GoIdent.GoName
	g.P("function ", moduleName, ".make", schemaName, "(config: {")
	g.P("\tfieldsMap: ", fieldsMapName, "?,")
	g.P("\tfieldKeysInOrder: {string}?,")
	g.P("\tlabel: string?,")
	g.P("}?): ", schemaName)
	g.P("\tlocal c = config or {}")
	g.P("\treturn {")
	g.P("\t\tfieldsMap = c.fieldsMap,")
	g.P("\t\tfieldKeysInOrder = c.fieldKeysInOrder or {},")
	g.P(`		label = c.label or "",`)
	g.P("\t}")
	g.P("end")
	g.P()
}

// emitPerModuleOptionalHelpers scans FieldsMap fields for DataTypes-imported types
// and emits decode/encode optional helpers for each unique type.
func emitPerModuleOptionalHelpers(g *protogen.GeneratedFile, fieldsMapMsg *protogen.Message, currentFile *protogen.File) {
	typeSet := make(map[string]bool)
	for _, field := range fieldsMapMsg.Fields {
		if isFieldFromDifferentFile(field, currentFile) {
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

// fieldsMapFieldHelperName returns the optional helper function name suffix for a FieldsMap field.
// For imported (DataTypes) fields, returns the GoName (e.g. "StringType").
// For local compound fields, returns "" (they use direct module-qualified calls).
func fieldsMapFieldHelperName(field *protogen.Field, currentFile *protogen.File) string {
	if isFieldFromDifferentFile(field, currentFile) {
		return field.Message.GoIdent.GoName
	}
	return ""
}

// emitFieldsMapEncode emits the encode function for a FieldsMap message.
func emitFieldsMapEncode(g *protogen.GeneratedFile, msg *protogen.Message, moduleName string, currentFile *protogen.File) {
	name := msg.GoIdent.GoName
	g.P("function ", moduleName, ".encodeJson", name, "(fm: ", name, "): {[string]: any}")
	g.P("\tlocal result: {[string]: any} = {}")
	for _, field := range msg.Fields {
		luauName := snakeToCamel(string(field.Desc.Name()))
		helperName := fieldsMapFieldHelperName(field, currentFile)
		if helperName != "" {
			// Imported DataTypes field — use optional helper
			g.P("\tresult.", luauName, " = encodeOptional", helperName, "(fm.", luauName, ")")
		} else {
			// Local compound — use if-then pattern with module-qualified call
			typeName := field.Message.GoIdent.GoName
			g.P("\tif fm.", luauName, " then result.", luauName, " = ", moduleName, ".encodeJson", typeName, "(fm.", luauName, ") end")
		}
	}
	g.P("\treturn result")
	g.P("end")
	g.P()
}

// emitFieldsMapDecode emits the decode function for a FieldsMap message.
func emitFieldsMapDecode(g *protogen.GeneratedFile, msg *protogen.Message, moduleName string, currentFile *protogen.File) {
	name := msg.GoIdent.GoName
	g.P("function ", moduleName, ".decodeJson", name, "(json: {[string]: any}): ", name)
	g.P("\treturn {")
	for _, field := range msg.Fields {
		luauName := snakeToCamel(string(field.Desc.Name()))
		helperName := fieldsMapFieldHelperName(field, currentFile)
		if helperName != "" {
			// Imported DataTypes field — use optional helper
			g.P("\t\t", luauName, " = decodeOptional", helperName, "(json.", luauName, "),")
		} else {
			// Local compound — use if-then-else pattern
			typeName := field.Message.GoIdent.GoName
			g.P("\t\t", luauName, " = if json.", luauName, " then ", moduleName, ".decodeJson", typeName, "(json.", luauName, ") else nil,")
		}
	}
	g.P("\t}")
	g.P("end")
	g.P()
}

// emitTableFileSchemaEncode emits the encode function for a table file Schema message.
func emitTableFileSchemaEncode(g *protogen.GeneratedFile, schema *protogen.Message, moduleName string, fieldsMap *protogen.Message) {
	schemaName := schema.GoIdent.GoName
	fieldsMapName := fieldsMap.GoIdent.GoName
	g.P("function ", moduleName, ".encodeJson", schemaName, "(schema: ", schemaName, "): {[string]: any}")
	g.P("\tlocal result: {[string]: any} = {")
	g.P("\t\tfieldKeysInOrder = schema.fieldKeysInOrder,")
	g.P("\t\tlabel = schema.label,")
	g.P("\t}")
	g.P("\tif schema.fieldsMap ~= nil then")
	g.P("\t\tresult.fieldsMap = ", moduleName, ".encodeJson", fieldsMapName, "(schema.fieldsMap)")
	g.P("\tend")
	g.P("\treturn result")
	g.P("end")
	g.P()
}

// emitTableFileSchemaDecode emits the decode function for a table file Schema message.
func emitTableFileSchemaDecode(g *protogen.GeneratedFile, schema *protogen.Message, moduleName string, fieldsMap *protogen.Message) {
	schemaName := schema.GoIdent.GoName
	fieldsMapName := fieldsMap.GoIdent.GoName
	g.P("function ", moduleName, ".decodeJson", schemaName, "(json: {[string]: any}): ", schemaName)
	g.P("\treturn {")
	g.P("\t\tfieldsMap = if json.fieldsMap then ", moduleName, ".decodeJson", fieldsMapName, "(json.fieldsMap) else nil,")
	g.P("\t\tfieldKeysInOrder = json.fieldKeysInOrder or {},")
	g.P(`		label = json.label or "",`)
	g.P("\t}")
	g.P("end")
	g.P()
}
