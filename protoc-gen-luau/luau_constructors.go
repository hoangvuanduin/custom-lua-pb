package main

import (
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func emitSimpleConstructors(g *protogen.GeneratedFile, msgs []*protogen.Message) {
	for _, msg := range msgs {
		emitSimpleConstructor(g, msg)
	}
}

func emitSimpleConstructor(g *protogen.GeneratedFile, msg *protogen.Message) {
	name := msg.GoIdent.GoName

	g.P("function DataTypes.make", name, "(config: {")
	for _, field := range msg.Fields {
		luauName := snakeToCamel(string(field.Desc.Name()))
		luauType := protoFieldToLuauType(field)
		g.P("\t", luauName, ": ", luauType, "?,")
	}
	g.P("}?): ", name)

	g.P("\tlocal c = config or {}")
	g.P("\treturn {")
	for _, field := range msg.Fields {
		luauName := snakeToCamel(string(field.Desc.Name()))
		dflt := luauDefaultValue(field)
		if field.Desc.Kind() == protoreflect.BoolKind {
			g.P("\t\t", luauName, " = if c.", luauName, " ~= nil then c.", luauName, " else false,")
		} else if isSimpleOptionalField(name, field) || dflt == "" {
			g.P("\t\t", luauName, " = c.", luauName, ",")
		} else {
			g.P("\t\t", luauName, " = c.", luauName, " or ", dflt, ",")
		}
	}
	g.P("\t}")
	g.P("end")
	g.P()
}

func emitCompoundSubFieldsConstructor(g *protogen.GeneratedFile, msg *protogen.Message, moduleName string, currentFile *protogen.File) {
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

func emitCompoundWrapperConstructor(g *protogen.GeneratedFile, wrapper *protogen.Message, subFieldsName string, moduleName string) {
	name := wrapper.GoIdent.GoName
	g.P("function ", moduleName, ".make", name, "(config: {")
	g.P("\ttypeId: string?,")
	g.P("\tvalueSubFields: ", subFieldsName, "?,")
	g.P("\tsubFieldKeysInOrder: {string}?,")
	g.P("\tlabel: string?,")
	g.P("}?): ", name)
	g.P("\tlocal c = config or {}")
	g.P("\treturn {")
	g.P(`		typeId = c.typeId or "",`)
	g.P("\t\tvalueSubFields = c.valueSubFields,")
	g.P("\t\tsubFieldKeysInOrder = c.subFieldKeysInOrder or {},")
	g.P("\t\tlabel = c.label,")
	g.P("\t}")
	g.P("end")
	g.P()
}

func emitCustomCompoundConstructor(g *protogen.GeneratedFile) {
	g.P("function DataTypes.makeCustomCompoundType(config: {")
	g.P("\ttypeId: string?,")
	g.P("\tvalueSubFields: {[string]: FieldValue}?,")
	g.P("\tsubFieldKeysInOrder: {string}?,")
	g.P("\tlabel: string?,")
	g.P("}?): CustomCompoundType")
	g.P("\tlocal c = config or {}")
	g.P("\treturn {")
	g.P(`		typeId = c.typeId or "",`)
	g.P("\t\tvalueSubFields = c.valueSubFields,")
	g.P("\t\tsubFieldKeysInOrder = c.subFieldKeysInOrder or {},")
	g.P("\t\tlabel = c.label,")
	g.P("\t}")
	g.P("end")
	g.P()
}
