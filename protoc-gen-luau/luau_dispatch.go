package main

import (
	"strings"

	"google.golang.org/protobuf/compiler/protogen"
)

func emitDispatchTables(g *protogen.GeneratedFile, pairs []compoundPair) {
	g.P("local STRING_TYPE_IDS: {[string]: boolean} = {")
	for i, id := range stringTypeIDs {
		comma := ","
		if i == len(stringTypeIDs)-1 {
			comma = ""
		}
		g.P("\t", id, " = true", comma)
	}
	g.P("}")
	g.P()

	g.P("local NUMBER_TYPE_IDS: {[string]: boolean} = {")
	for i, id := range numberTypeIDs {
		comma := ","
		if i == len(numberTypeIDs)-1 {
			comma = ""
		}
		g.P("\t", id, " = true", comma)
	}
	g.P("}")
	g.P()

	g.P("local COMPOUND_TYPE_CODECS: {[string]: {decode: string, encode: string}} = {")
	for _, pair := range pairs {
		name := strings.TrimSuffix(pair.wrapper.GoIdent.GoName, "Type")
		g.P("\t", name, " = { decode = \"decodeJson", name, "Type\", encode = \"encodeJson", name, "Type\" },")
	}
	g.P("}")
	g.P()

	g.P("local ENUM_TYPE_IDS: {[string]: boolean} = {")
	for i, id := range enumTypeIDs {
		comma := ","
		if i == len(enumTypeIDs)-1 {
			comma = ""
		}
		g.P("\t", id, " = true", comma)
	}
	g.P("}")
	g.P()
}

func emitDispatchFunctions(g *protogen.GeneratedFile) {
	// decodeFieldValueByTypeId
	g.P("function DataTypes.decodeFieldValueByTypeId(json: {[string]: any}): FieldValue")
	g.P(`	local typeId = json.typeId or ""`)
	g.P()
	g.P("\tif STRING_TYPE_IDS[typeId] then")
	g.P("\t\treturn DataTypes.decodeJsonStringType(json) :: any")
	g.P("\telseif NUMBER_TYPE_IDS[typeId] then")
	g.P("\t\treturn DataTypes.decodeJsonNumberType(json) :: any")
	g.P(`	elseif typeId == "Boolean" then`)
	g.P("\t\treturn DataTypes.decodeJsonBooleanType(json) :: any")
	g.P("\telseif ENUM_TYPE_IDS[typeId] then")
	g.P("\t\treturn DataTypes.decodeJsonEnumType(json) :: any")
	g.P(`	elseif typeId == "MultipleCheckbox" then`)
	g.P("\t\treturn DataTypes.decodeJsonMultipleCheckboxType(json) :: any")
	g.P(`	elseif typeId == "RadioGroup" then`)
	g.P("\t\treturn DataTypes.decodeJsonRadioGroupType(json) :: any")
	g.P(`	elseif typeId == "CustomCompound" or typeId == "CustomCompoundType" then`)
	g.P("\t\treturn DataTypes.decodeJsonCustomCompoundType(json) :: any")
	g.P("\telse")
	g.P("\t\tlocal codec = COMPOUND_TYPE_CODECS[typeId]")
	g.P("\t\tif codec then")
	g.P("\t\t\tlocal decoder = (DataTypes :: any)[codec.decode]")
	g.P("\t\t\treturn decoder(json)")
	g.P("\t\tend")
	g.P("\t\treturn json :: any")
	g.P("\tend")
	g.P("end")
	g.P()

	// encodeFieldValueByTypeId
	g.P("function DataTypes.encodeFieldValueByTypeId(value: FieldValue): {[string]: any}")
	g.P(`	local typeId = (value :: any).typeId or ""`)
	g.P()
	g.P("\tif STRING_TYPE_IDS[typeId] then")
	g.P("\t\treturn DataTypes.encodeJsonStringType(value :: any)")
	g.P("\telseif NUMBER_TYPE_IDS[typeId] then")
	g.P("\t\treturn DataTypes.encodeJsonNumberType(value :: any)")
	g.P(`	elseif typeId == "Boolean" then`)
	g.P("\t\treturn DataTypes.encodeJsonBooleanType(value :: any)")
	g.P("\telseif ENUM_TYPE_IDS[typeId] then")
	g.P("\t\treturn DataTypes.encodeJsonEnumType(value :: any)")
	g.P(`	elseif typeId == "MultipleCheckbox" then`)
	g.P("\t\treturn DataTypes.encodeJsonMultipleCheckboxType(value :: any)")
	g.P(`	elseif typeId == "RadioGroup" then`)
	g.P("\t\treturn DataTypes.encodeJsonRadioGroupType(value :: any)")
	g.P(`	elseif typeId == "CustomCompound" or typeId == "CustomCompoundType" then`)
	g.P("\t\treturn DataTypes.encodeJsonCustomCompoundType(value :: any)")
	g.P("\telse")
	g.P("\t\tlocal codec = COMPOUND_TYPE_CODECS[typeId]")
	g.P("\t\tif codec then")
	g.P("\t\t\tlocal encoder = (DataTypes :: any)[codec.encode]")
	g.P("\t\t\treturn encoder(value)")
	g.P("\t\tend")
	g.P("\t\treturn value :: any")
	g.P("\tend")
	g.P("end")
	g.P()
}
