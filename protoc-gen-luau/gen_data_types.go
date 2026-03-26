package main

import (
	"google.golang.org/protobuf/compiler/protogen"
)

func generateDataTypes(plugin *protogen.Plugin, file *protogen.File) {
	g := plugin.NewGeneratedFile("data_types.luau", "")
	cm := classifyMessages(file)

	g.P("--!strict")
	g.P("local DataTypes = {}")
	g.P()

	emitSimpleTypeDefs(g, cm.simpleTypes)
	emitSimpleConstructors(g, cm.simpleTypes)
	emitSimpleCodecs(g, cm.simpleTypes)
	emitFieldValueType(g)
	emitOptionalHelpers(g, cm.compoundPairs)

	for _, pair := range cm.compoundPairs {
		sfName := pair.subFields.GoIdent.GoName
		emitCompoundSubFieldsDef(g, pair.subFields, file)
		emitCompoundWrapperDef(g, pair.wrapper, sfName)
		g.P()
		emitCompoundSubFieldsConstructor(g, pair.subFields, "DataTypes", file)
		emitCompoundWrapperConstructor(g, pair.wrapper, sfName, "DataTypes")
		emitCompoundSubFieldsEncode(g, pair.subFields, "DataTypes")
		emitCompoundSubFieldsDecode(g, pair.subFields, "DataTypes")
		emitCompoundWrapperEncode(g, pair.wrapper, sfName, "DataTypes")
		emitCompoundWrapperDecode(g, pair.wrapper, sfName, "DataTypes")
	}

	emitCustomCompoundDef(g)
	emitCustomCompoundConstructor(g)
	emitDispatchTables(g, cm.compoundPairs)
	emitDispatchFunctions(g)
	emitCustomCompoundCodecs(g)

	g.P("return DataTypes")
}
