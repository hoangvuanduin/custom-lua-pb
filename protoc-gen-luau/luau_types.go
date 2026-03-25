package main

import (
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func emitSimpleTypeDefs(g *protogen.GeneratedFile, msgs []*protogen.Message) {
	for _, msg := range msgs {
		name := msg.GoIdent.GoName
		g.P("export type ", name, " = {")
		for _, field := range msg.Fields {
			luauName := snakeToCamel(string(field.Desc.Name()))
			luauType := protoFieldToLuauType(field)
			if isSimpleOptionalField(name, field) {
				g.P("\t", luauName, ": ", luauType, "?,")
			} else {
				g.P("\t", luauName, ": ", luauType, ",")
			}
		}
		g.P("}")
		g.P()
	}
}

func emitCompoundSubFieldsDef(g *protogen.GeneratedFile, msg *protogen.Message) {
	name := msg.GoIdent.GoName
	g.P("export type ", name, " = {")
	for _, field := range msg.Fields {
		luauName := snakeToCamel(string(field.Desc.Name()))
		luauType := protoFieldToLuauType(field)
		if field.Desc.Kind() == protoreflect.MessageKind {
			g.P("\t", luauName, ": ", luauType, "?,")
		} else {
			g.P("\t", luauName, ": ", luauType, ",")
		}
	}
	g.P("}")
}

func emitCompoundWrapperDef(g *protogen.GeneratedFile, wrapper *protogen.Message, subFieldsName string) {
	name := wrapper.GoIdent.GoName
	g.P("export type ", name, " = {")
	g.P("\ttypeId: string,")
	g.P("\tvalueSubFields: ", subFieldsName, "?,")
	g.P("\tsubFieldKeysInOrder: {string},")
	g.P("\tlabel: string?,")
	g.P("}")
}

func emitFieldValueType(g *protogen.GeneratedFile) {
	g.P("export type FieldValue = {")
	g.P("\ttypeId: string,")
	g.P("\t[string]: any,")
	g.P("}")
	g.P()
}

func emitCustomCompoundDef(g *protogen.GeneratedFile) {
	g.P("export type CustomCompoundType = {")
	g.P("\ttypeId: string,")
	g.P("\tvalueSubFields: {[string]: FieldValue}?,")
	g.P("\tsubFieldKeysInOrder: {string},")
	g.P("\tlabel: string?,")
	g.P("}")
	g.P()
}
