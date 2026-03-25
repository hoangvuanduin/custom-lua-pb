package main

import (
	"google.golang.org/protobuf/compiler/protogen"
)

func generateTableSchema(plugin *protogen.Plugin, file *protogen.File) {
	g := plugin.NewGeneratedFile("table_schema.luau", "")

	g.P("--!strict")
	g.P(`local DataTypes = require("@lib/data_types")`)
	g.P()
	g.P("local TableSchema = {}")
	g.P()

	// Find the FieldGroup message for generic field walking
	var fieldGroupMsg *protogen.Message
	for _, msg := range file.Messages {
		if msg.GoIdent.GoName == "FieldGroup" {
			fieldGroupMsg = msg
			break
		}
	}

	// Type definitions
	emitSingleFieldTypeDef(g)
	if fieldGroupMsg != nil {
		emitFieldGroupDef(g, fieldGroupMsg)
	}
	emitTableSchemaDef(g)

	// Constructors
	emitSingleFieldTypeConstructor(g)
	if fieldGroupMsg != nil {
		emitFieldGroupConstructor(g, fieldGroupMsg)
	}
	emitTableSchemaConstructor(g)

	// Codecs
	emitSingleFieldTypeCodecs(g)
	emitTableSchemaCodecs(g)

	g.P("return TableSchema")
}

func emitSingleFieldTypeDef(g *protogen.GeneratedFile) {
	g.P("export type SingleFieldType = {")
	g.P("\tvalue: DataTypes.FieldValue?,")
	g.P("\tlabel: string,")
	g.P("}")
	g.P()
}

func emitFieldGroupDef(g *protogen.GeneratedFile, msg *protogen.Message) {
	g.P("export type FieldGroup = {")
	for _, field := range msg.Fields {
		luauName := snakeToCamel(string(field.Desc.Name()))
		luauType := protoFieldToLuauType(field)
		g.P("\t", luauName, ": ", luauType, ",")
	}
	g.P("}")
	g.P()
}

func emitTableSchemaDef(g *protogen.GeneratedFile) {
	g.P("export type TableSchema = {")
	g.P("\tfieldsMap: {[string]: SingleFieldType},")
	g.P("\tfieldKeysInOrder: {string},")
	g.P("\tlabel: string,")
	g.P("\tgroups: {FieldGroup},")
	g.P("}")
	g.P()
}

func emitSingleFieldTypeConstructor(g *protogen.GeneratedFile) {
	g.P("function TableSchema.makeSingleFieldType(config: {")
	g.P("\tvalue: DataTypes.FieldValue?,")
	g.P("\tlabel: string?,")
	g.P("}?): SingleFieldType")
	g.P("\tlocal c = config or {}")
	g.P("\treturn {")
	g.P(`		value = c.value,`)
	g.P(`		label = c.label or "",`)
	g.P("\t}")
	g.P("end")
	g.P()
}

func emitFieldGroupConstructor(g *protogen.GeneratedFile, msg *protogen.Message) {
	g.P("function TableSchema.makeFieldGroup(config: {")
	for _, field := range msg.Fields {
		luauName := snakeToCamel(string(field.Desc.Name()))
		luauType := protoFieldToLuauType(field)
		g.P("\t", luauName, ": ", luauType, "?,")
	}
	g.P("}?): FieldGroup")
	g.P("\tlocal c = config or {}")
	g.P("\treturn {")
	for _, field := range msg.Fields {
		luauName := snakeToCamel(string(field.Desc.Name()))
		dflt := luauDefaultValue(field)
		if dflt == "" {
			g.P("\t\t", luauName, " = c.", luauName, ",")
		} else {
			g.P("\t\t", luauName, " = c.", luauName, " or ", dflt, ",")
		}
	}
	g.P("\t}")
	g.P("end")
	g.P()
}

func emitTableSchemaConstructor(g *protogen.GeneratedFile) {
	g.P("function TableSchema.makeTableSchema(config: {")
	g.P("\tfieldsMap: {[string]: SingleFieldType}?,")
	g.P("\tfieldKeysInOrder: {string}?,")
	g.P("\tlabel: string?,")
	g.P("\tgroups: {FieldGroup}?,")
	g.P("}?): TableSchema")
	g.P("\tlocal c = config or {}")
	g.P("\treturn {")
	g.P("\t\tfieldsMap = c.fieldsMap or {},")
	g.P("\t\tfieldKeysInOrder = c.fieldKeysInOrder or {},")
	g.P(`		label = c.label or "",`)
	g.P("\t\tgroups = c.groups or {},")
	g.P("\t}")
	g.P("end")
	g.P()
}

func emitSingleFieldTypeCodecs(g *protogen.GeneratedFile) {
	// encode
	g.P("function TableSchema.encodeJsonSingleFieldType(sft: SingleFieldType): {[string]: any}")
	g.P("\tlocal result: {[string]: any} = { label = sft.label }")
	g.P("\tif sft.value ~= nil then")
	g.P("\t\tresult.value = DataTypes.encodeFieldValueByTypeId(sft.value)")
	g.P("\tend")
	g.P("\treturn result")
	g.P("end")
	g.P()

	// decode
	g.P("function TableSchema.decodeJsonSingleFieldType(json: {[string]: any}): SingleFieldType")
	g.P("\treturn {")
	g.P(`		label = json.label or "",`)
	g.P("\t\tvalue = if json.value ~= nil then DataTypes.decodeFieldValueByTypeId(json.value) else nil,")
	g.P("\t}")
	g.P("end")
	g.P()
}

func emitTableSchemaCodecs(g *protogen.GeneratedFile) {
	// encode
	g.P("function TableSchema.encodeJsonTableSchema(ts: TableSchema): {[string]: any}")
	g.P("\tlocal fieldsMapEncoded: {[string]: any} = {}")
	g.P("\tfor key, sft in ts.fieldsMap do")
	g.P("\t\tfieldsMapEncoded[key] = TableSchema.encodeJsonSingleFieldType(sft)")
	g.P("\tend")
	g.P("\tlocal groupsEncoded: {any} = {}")
	g.P("\tfor _, g in ts.groups do")
	g.P("\t\ttable.insert(groupsEncoded, { label = g.label, startIdx = g.startIdx, endIdx = g.endIdx })")
	g.P("\tend")
	g.P("\treturn {")
	g.P("\t\tfieldsMap = fieldsMapEncoded,")
	g.P("\t\tfieldKeysInOrder = ts.fieldKeysInOrder,")
	g.P("\t\tlabel = ts.label,")
	g.P("\t\tgroups = groupsEncoded,")
	g.P("\t}")
	g.P("end")
	g.P()

	// decode
	g.P("function TableSchema.decodeJsonTableSchema(json: {[string]: any}): TableSchema")
	g.P("\tlocal fieldsMap: {[string]: SingleFieldType} = {}")
	g.P("\tif json.fieldsMap ~= nil then")
	g.P("\t\tfor key, sft in json.fieldsMap :: {[string]: any} do")
	g.P("\t\t\tfieldsMap[key] = TableSchema.decodeJsonSingleFieldType(sft)")
	g.P("\t\tend")
	g.P("\tend")
	g.P("\tlocal groups: {FieldGroup} = {}")
	g.P("\tif json.groups ~= nil then")
	g.P("\t\tfor _, g in json.groups :: {{[string]: any}} do")
	g.P("\t\t\ttable.insert(groups, {")
	g.P(`			label = g.label or "",`)
	g.P("\t\t\t\tstartIdx = g.startIdx or 0,")
	g.P("\t\t\t\tendIdx = g.endIdx or 0,")
	g.P("\t\t\t})")
	g.P("\t\tend")
	g.P("\tend")
	g.P("\treturn {")
	g.P("\t\tfieldsMap = fieldsMap,")
	g.P("\t\tfieldKeysInOrder = json.fieldKeysInOrder or {},")
	g.P(`		label = json.label or "",`)
	g.P("\t\tgroups = groups,")
	g.P("\t}")
	g.P("end")
	g.P()
}
