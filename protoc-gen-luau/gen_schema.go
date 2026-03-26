package main

import (
	"fmt"
	"path"
	"strings"

	"google.golang.org/protobuf/compiler/protogen"
)

func generateSchema(plugin *protogen.Plugin, file *protogen.File) {
	protoPath := file.Desc.Path()
	baseName := strings.TrimSuffix(path.Base(protoPath), ".proto")
	moduleName := deriveModuleName(protoPath)
	requirePath := getDataTypesRequirePath(file)

	g := plugin.NewGeneratedFile(baseName+".luau", "")
	cm := classifyMessages(file)

	// Collect all non-compound messages for generic handling
	var genericMsgs []*protogen.Message
	genericMsgs = append(genericMsgs, cm.simpleTypes...)
	if cm.customCompound != nil {
		genericMsgs = append(genericMsgs, cm.customCompound)
	}
	if cm.oneofMsg != nil {
		genericMsgs = append(genericMsgs, cm.oneofMsg)
	}

	// Topological sort for correct dependency ordering
	sorted, err := topoSortMessages(genericMsgs, file.Desc)
	if err != nil {
		g.P(fmt.Sprintf("-- ERROR: %v", err))
		return
	}

	// === Header ===
	g.P("--!strict")
	g.P(`local DataTypes = require("`, requirePath, `")`)
	g.P()
	g.P("local ", moduleName, " = {}")
	g.P()

	// === Compound pair type definitions ===
	for _, pair := range cm.compoundPairs {
		sfName := pair.subFields.GoIdent.GoName
		emitCompoundSubFieldsDef(g, pair.subFields, file)
		g.P()
		emitCompoundWrapperDef(g, pair.wrapper, sfName)
		g.P()
	}

	// === Generic type definitions (topo-sorted) ===
	for _, msg := range sorted {
		emitSchemaTypeDef(g, msg, file)
	}

	// === Compound pair constructors ===
	for _, pair := range cm.compoundPairs {
		sfName := pair.subFields.GoIdent.GoName
		emitCompoundSubFieldsConstructor(g, pair.subFields, moduleName, file)
		emitCompoundWrapperConstructor(g, pair.wrapper, sfName, moduleName)
	}

	// === Generic constructors (topo-sorted) ===
	for _, msg := range sorted {
		emitSchemaConstructor(g, msg, moduleName, file)
	}

	// === Per-module optional helpers for DataTypes-imported types ===
	// Collect from both compound pair SubFields and generic messages
	var allMsgsForHelpers []*protogen.Message
	for _, pair := range cm.compoundPairs {
		allMsgsForHelpers = append(allMsgsForHelpers, pair.subFields)
	}
	allMsgsForHelpers = append(allMsgsForHelpers, sorted...)
	emitSchemaOptionalHelpers(g, allMsgsForHelpers, file)

	// === Compound pair codecs ===
	for _, pair := range cm.compoundPairs {
		sfName := pair.subFields.GoIdent.GoName
		emitCompoundSubFieldsEncodeWith(g, pair.subFields, moduleName, tableFileHelperName)
		emitCompoundSubFieldsDecodeWith(g, pair.subFields, moduleName, tableFileHelperName)
		emitCompoundWrapperEncode(g, pair.wrapper, sfName, moduleName)
		emitCompoundWrapperDecode(g, pair.wrapper, sfName, moduleName)
	}

	// === Generic codecs (topo-sorted) ===
	for _, msg := range sorted {
		emitSchemaEncode(g, msg, moduleName, file)
		emitSchemaDecode(g, msg, moduleName, file)
	}

	// === Footer ===
	g.P("return ", moduleName)
}
