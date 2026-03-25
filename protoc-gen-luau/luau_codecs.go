package main

import (
	"strings"

	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func emitSimpleCodecs(g *protogen.GeneratedFile, msgs []*protogen.Message) {
	for _, msg := range msgs {
		emitSimpleEncode(g, msg)
		emitSimpleDecode(g, msg)
	}
}

func emitSimpleEncode(g *protogen.GeneratedFile, msg *protogen.Message) {
	name := msg.GoIdent.GoName
	paramName := strings.ToLower(name[:1])

	g.P("function DataTypes.encodeJson", name, "(", paramName, ": ", name, "): {[string]: any}")

	var required, optional []*protogen.Field
	for _, field := range msg.Fields {
		if isSimpleOptionalField(name, field) {
			optional = append(optional, field)
		} else {
			required = append(required, field)
		}
	}

	if len(optional) == 0 {
		g.P("\treturn {")
		for _, field := range required {
			luauName := snakeToCamel(string(field.Desc.Name()))
			g.P("\t\t", luauName, " = ", paramName, ".", luauName, ",")
		}
		g.P("\t}")
	} else {
		g.P("\tlocal result: {[string]: any} = {")
		for _, field := range required {
			luauName := snakeToCamel(string(field.Desc.Name()))
			g.P("\t\t", luauName, " = ", paramName, ".", luauName, ",")
		}
		g.P("\t}")
		for _, field := range optional {
			luauName := snakeToCamel(string(field.Desc.Name()))
			g.P("\tif ", paramName, ".", luauName, " ~= nil then")
			g.P("\t\tresult.", luauName, " = ", paramName, ".", luauName)
			g.P("\tend")
		}
		g.P("\treturn result")
	}

	g.P("end")
	g.P()
}

func emitSimpleDecode(g *protogen.GeneratedFile, msg *protogen.Message) {
	name := msg.GoIdent.GoName

	g.P("function DataTypes.decodeJson", name, "(json: {[string]: any}): ", name)
	g.P("\treturn {")
	for _, field := range msg.Fields {
		luauName := snakeToCamel(string(field.Desc.Name()))
		if isSimpleOptionalField(name, field) {
			g.P("\t\t", luauName, " = json.", luauName, ",")
		} else if field.Desc.Kind() == protoreflect.BoolKind {
			g.P("\t\t", luauName, " = if json.", luauName, " ~= nil then json.", luauName, " else false,")
		} else {
			dflt := luauDefaultValue(field)
			if dflt != "" {
				g.P("\t\t", luauName, " = json.", luauName, " or ", dflt, ",")
			} else {
				g.P("\t\t", luauName, " = json.", luauName, ",")
			}
		}
	}
	g.P("\t}")
	g.P("end")
	g.P()
}

func emitOptionalHelpers(g *protogen.GeneratedFile, pairs []compoundPair) {
	usedTypes := make(map[string]bool)
	for _, pair := range pairs {
		for _, field := range pair.subFields.Fields {
			if field.Desc.Kind() == protoreflect.MessageKind {
				usedTypes[field.Message.GoIdent.GoName] = true
			}
		}
	}

	for _, typeName := range []string{"StringType", "NumberType", "MultipleCheckboxType"} {
		if !usedTypes[typeName] {
			continue
		}
		shortName := strings.TrimSuffix(typeName, "Type")

		g.P("local function encodeOptional", shortName, "Field(field: ", typeName, "?): {[string]: any}?")
		g.P("\tif field == nil then")
		g.P("\t\treturn nil")
		g.P("\tend")
		g.P("\treturn DataTypes.encodeJson", typeName, "(field)")
		g.P("end")
		g.P()
		g.P("local function decodeOptional", shortName, "Field(json: any): ", typeName, "?")
		g.P("\tif json == nil then")
		g.P("\t\treturn nil")
		g.P("\tend")
		g.P("\treturn DataTypes.decodeJson", typeName, "(json)")
		g.P("end")
		g.P()
	}
}

func optionalHelperName(field *protogen.Field) string {
	if field.Desc.Kind() != protoreflect.MessageKind {
		return ""
	}
	return strings.TrimSuffix(field.Message.GoIdent.GoName, "Type")
}

func emitCompoundSubFieldsEncode(g *protogen.GeneratedFile, msg *protogen.Message) {
	name := msg.GoIdent.GoName
	g.P("function DataTypes.encodeJson", name, "(sf: ", name, "): {[string]: any}")
	g.P("\tlocal result: {[string]: any} = {}")
	for _, field := range msg.Fields {
		luauName := snakeToCamel(string(field.Desc.Name()))
		helper := optionalHelperName(field)
		g.P("\tresult.", luauName, " = encodeOptional", helper, "Field(sf.", luauName, ")")
	}
	g.P("\treturn result")
	g.P("end")
	g.P()
}

func emitCompoundSubFieldsDecode(g *protogen.GeneratedFile, msg *protogen.Message) {
	name := msg.GoIdent.GoName
	g.P("function DataTypes.decodeJson", name, "(json: {[string]: any}): ", name)
	g.P("\treturn {")
	for _, field := range msg.Fields {
		luauName := snakeToCamel(string(field.Desc.Name()))
		helper := optionalHelperName(field)
		g.P("\t\t", luauName, " = decodeOptional", helper, "Field(json.", luauName, "),")
	}
	g.P("\t}")
	g.P("end")
	g.P()
}

func emitCompoundWrapperEncode(g *protogen.GeneratedFile, wrapper *protogen.Message, subFieldsName string) {
	name := wrapper.GoIdent.GoName
	g.P("function DataTypes.encodeJson", name, "(t: ", name, "): {[string]: any}")
	g.P("\tlocal result: {[string]: any} = {")
	g.P("\t\ttypeId = t.typeId,")
	g.P("\t\tsubFieldKeysInOrder = t.subFieldKeysInOrder,")
	g.P("\t}")
	g.P("\tif t.label ~= nil then")
	g.P("\t\tresult.label = t.label")
	g.P("\tend")
	g.P("\tif t.valueSubFields ~= nil then")
	g.P("\t\tresult.valueSubFields = DataTypes.encodeJson", subFieldsName, "(t.valueSubFields)")
	g.P("\tend")
	g.P("\treturn result")
	g.P("end")
	g.P()
}

func emitCompoundWrapperDecode(g *protogen.GeneratedFile, wrapper *protogen.Message, subFieldsName string) {
	name := wrapper.GoIdent.GoName
	g.P("function DataTypes.decodeJson", name, "(json: {[string]: any}): ", name)
	g.P("\treturn {")
	g.P(`		typeId = json.typeId or "",`)
	g.P("\t\tvalueSubFields = if json.valueSubFields then DataTypes.decodeJson", subFieldsName, "(json.valueSubFields) else nil,")
	g.P("\t\tsubFieldKeysInOrder = json.subFieldKeysInOrder or {},")
	g.P("\t\tlabel = json.label,")
	g.P("\t}")
	g.P("end")
	g.P()
}

func emitCustomCompoundCodecs(g *protogen.GeneratedFile) {
	g.P("function DataTypes.encodeJsonCustomCompoundType(t: CustomCompoundType): {[string]: any}")
	g.P("\tlocal result: {[string]: any} = {")
	g.P("\t\ttypeId = t.typeId,")
	g.P("\t\tsubFieldKeysInOrder = t.subFieldKeysInOrder,")
	g.P("\t}")
	g.P("\tif t.label ~= nil then")
	g.P("\t\tresult.label = t.label")
	g.P("\tend")
	g.P("\tif t.valueSubFields ~= nil then")
	g.P("\t\tlocal encoded: {[string]: any} = {}")
	g.P("\t\tfor key, value in t.valueSubFields do")
	g.P("\t\t\tencoded[key] = DataTypes.encodeFieldValueByTypeId(value)")
	g.P("\t\tend")
	g.P("\t\tresult.valueSubFields = encoded")
	g.P("\tend")
	g.P("\treturn result")
	g.P("end")
	g.P()

	g.P("function DataTypes.decodeJsonCustomCompoundType(json: {[string]: any}): CustomCompoundType")
	g.P("\tlocal subFields: {[string]: FieldValue}? = nil")
	g.P("\tif json.valueSubFields ~= nil then")
	g.P("\t\tlocal decoded: {[string]: FieldValue} = {}")
	g.P("\t\tfor key, value in json.valueSubFields :: {[string]: any} do")
	g.P("\t\t\tdecoded[key] = DataTypes.decodeFieldValueByTypeId(value)")
	g.P("\t\tend")
	g.P("\t\tsubFields = decoded")
	g.P("\tend")
	g.P("\treturn {")
	g.P(`		typeId = json.typeId or "",`)
	g.P("\t\tvalueSubFields = subFields,")
	g.P("\t\tsubFieldKeysInOrder = json.subFieldKeysInOrder or {},")
	g.P("\t\tlabel = json.label,")
	g.P("\t}")
	g.P("end")
	g.P()
}
