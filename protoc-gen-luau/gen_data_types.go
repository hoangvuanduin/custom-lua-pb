package main

import (
	"sort"
	"strings"

	"google.golang.org/protobuf/compiler/protogen"
)

type compoundPair struct {
	subFields *protogen.Message
	wrapper   *protogen.Message
}

type classifiedMessages struct {
	simpleTypes    []*protogen.Message
	compoundPairs  []compoundPair
	customCompound *protogen.Message
	oneofMsg       *protogen.Message
}

func classifyMessages(file *protogen.File) *classifiedMessages {
	cm := &classifiedMessages{}
	subFieldsMap := make(map[string]*protogen.Message)
	wrapperMap := make(map[string]*protogen.Message)

	for _, msg := range file.Messages {
		name := msg.GoIdent.GoName

		hasRealOneof := false
		for _, oneof := range msg.Oneofs {
			if !oneof.Desc.IsSynthetic() {
				hasRealOneof = true
				break
			}
		}
		if hasRealOneof {
			cm.oneofMsg = msg
			continue
		}

		hasMap := false
		for _, f := range msg.Fields {
			if f.Desc.IsMap() {
				hasMap = true
				break
			}
		}
		if hasMap {
			cm.customCompound = msg
			continue
		}

		if strings.HasSuffix(name, "SubFields") {
			prefix := strings.TrimSuffix(name, "SubFields")
			subFieldsMap[prefix] = msg
			continue
		}

		if strings.HasSuffix(name, "Fields") {
			prefix := strings.TrimSuffix(name, "Fields")
			subFieldsMap[prefix] = msg
			continue
		}

		hasValueSubFields := false
		for _, f := range msg.Fields {
			if string(f.Desc.Name()) == "value_sub_fields" {
				hasValueSubFields = true
				break
			}
		}
		if hasValueSubFields {
			prefix := strings.TrimSuffix(name, "Type")
			wrapperMap[prefix] = msg
			continue
		}

		cm.simpleTypes = append(cm.simpleTypes, msg)
	}

	for prefix, sf := range subFieldsMap {
		if w, ok := wrapperMap[prefix]; ok {
			cm.compoundPairs = append(cm.compoundPairs, compoundPair{
				subFields: sf,
				wrapper:   w,
			})
		}
	}
	sort.Slice(cm.compoundPairs, func(i, j int) bool {
		return cm.compoundPairs[i].wrapper.GoIdent.GoName < cm.compoundPairs[j].wrapper.GoIdent.GoName
	})

	return cm
}

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
