package main

import (
	"path"
	"strings"
	"unicode"

	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"

	"protoc-gen-luau/mappingoptions"
)

// mappingField represents a single source or target field in a mapping rule.
type mappingField struct {
	structField string   // field name in the Input/Output struct (derived from actual path)
	rawPath     string   // original annotation value e.g. "master.some_field"
	pathParts   []string // split by ".", each part converted to camelCase
	isSource    bool
}

// mappingRule represents one proto message that defines a mapping rule.
type mappingRule struct {
	messageName string
	sources     []mappingField
	targets     []mappingField
}

// parseMappingRules extracts mapping rules from all messages in the file.
func parseMappingRules(file *protogen.File) []mappingRule {
	var rules []mappingRule
	for _, msg := range file.Messages {
		rule := mappingRule{
			messageName: msg.GoIdent.GoName,
		}
		for _, field := range msg.Fields {
			opts := field.Desc.Options()
			if opts == nil {
				continue
			}
			if proto.HasExtension(opts, mappingoptions.E_Source) {
				rawPath := proto.GetExtension(opts, mappingoptions.E_Source).(string)
				mf := parseMappingField(rawPath, true)
				rule.sources = append(rule.sources, mf)
			}
			if proto.HasExtension(opts, mappingoptions.E_Target) {
				rawPath := proto.GetExtension(opts, mappingoptions.E_Target).(string)
				mf := parseMappingField(rawPath, false)
				rule.targets = append(rule.targets, mf)
			}
		}
		if len(rule.sources) > 0 || len(rule.targets) > 0 {
			rules = append(rules, rule)
		}
	}
	return rules
}

// parseMappingField converts a raw annotation path into a mappingField.
// For sources: "master.some_field" → structField "masterSomeField"
// For sources (compound): "master.wire_instructions.bank_name" → structField "masterWireInstructionsBankName"
// For targets: "target.sf_some_field" → structField "sfSomeField" (strips "target." prefix)
func parseMappingField(rawPath string, isSource bool) mappingField {
	parts := strings.Split(rawPath, ".")
	camelParts := make([]string, len(parts))
	for i, p := range parts {
		camelParts[i] = snakeToCamel(p)
	}

	var structField string
	if isSource {
		// Join all camelCase parts: "master" + "SomeField" → "masterSomeField"
		structField = camelParts[0]
		for _, p := range camelParts[1:] {
			// Capitalize first char of each subsequent part
			runes := []rune(p)
			if len(runes) > 0 {
				runes[0] = unicode.ToUpper(runes[0])
			}
			structField += string(runes)
		}
	} else {
		// For targets, strip "target." prefix — use the field name directly
		if len(camelParts) >= 2 {
			structField = camelParts[1]
			for _, p := range camelParts[2:] {
				runes := []rune(p)
				if len(runes) > 0 {
					runes[0] = unicode.ToUpper(runes[0])
				}
				structField += string(runes)
			}
		} else {
			structField = strings.Join(camelParts, "")
		}
	}

	return mappingField{
		structField: structField,
		rawPath:     rawPath,
		pathParts:   camelParts,
		isSource:    isSource,
	}
}

// transformFuncName returns the transform function name: first char lowercased.
func transformFuncName(messageName string) string {
	if messageName == "" {
		return ""
	}
	runes := []rune(messageName)
	runes[0] = unicode.ToLower(runes[0])
	return string(runes)
}

// buildExtraction builds the Luau expression for reading a source field.
// 2-segment: input.table and input.table.field or nil
// 3-segment: input.table and input.table.parent and input.table.parent.valueSubFields and input.table.parent.valueSubFields.child or nil
func buildExtraction(pathParts []string) string {
	if len(pathParts) == 2 {
		table := pathParts[0]
		field := pathParts[1]
		return "input." + table + " and input." + table + "." + field + " or nil"
	}
	if len(pathParts) == 3 {
		table := pathParts[0]
		parent := pathParts[1]
		child := pathParts[2]
		prefix := "input." + table
		return prefix + " and " +
			prefix + "." + parent + " and " +
			prefix + "." + parent + ".valueSubFields and " +
			prefix + "." + parent + ".valueSubFields." + child + " or nil"
	}
	var parts []string
	acc := "input"
	for _, p := range pathParts {
		acc += "." + p
		parts = append(parts, acc)
	}
	return strings.Join(parts, " and ") + " or nil"
}

// buildTargetAccess builds the Luau expression for writing a target field.
func buildTargetAccess(pathParts []string) string {
	return "output." + strings.Join(pathParts, ".")
}

// generateMapping is the main entry point for the mapping generator.
func generateMapping(plugin *protogen.Plugin, file *protogen.File) {
	protoPath := file.Desc.Path()
	baseName := strings.TrimSuffix(path.Base(protoPath), ".proto")

	rules := parseMappingRules(file)
	if len(rules) == 0 {
		return
	}

	transformsModule := strings.Replace(baseName, "_mappings", "_transforms", 1)
	transformsRequire := "@lib/" + transformsModule

	// === Types file (always generated — every rule gets Input + Output types) ===
	gt := plugin.NewGeneratedFile(baseName+"_types.luau", "")
	gt.P("-- GENERATED by protoc-gen-luau \u2014 DO NOT EDIT")
	gt.P("--!strict")
	gt.P(`local DataTypes = require("@lib/data_types")`)
	gt.P()
	gt.P("local Types = {}")
	gt.P()

	for _, rule := range rules {
		// Input type
		gt.P("export type ", rule.messageName, "Input = {")
		for _, s := range rule.sources {
			gt.P("\t", s.structField, ": DataTypes.StringType?,")
		}
		gt.P("}")
		gt.P()

		// Output type
		gt.P("export type ", rule.messageName, "Output = {")
		for _, t := range rule.targets {
			gt.P("\t", t.structField, ": DataTypes.StringType?,")
		}
		gt.P("}")
		gt.P()
	}

	gt.P("return Types")

	// === Generated file ===
	g := plugin.NewGeneratedFile(baseName+"_generated.luau", "")
	g.P("-- GENERATED by protoc-gen-luau \u2014 DO NOT EDIT")
	g.P("--!strict")
	g.P(`local DataTypes = require("@lib/data_types")`)
	g.P(`local SourceTables = require("@lib/source_tables")`)
	g.P(`local TargetTables = require("@lib/target_tables")`)
	g.P(`local Transforms = require("`, transformsRequire, `")`)
	g.P()
	g.P("local Generated = {}")
	g.P()

	// === Glue functions ===
	for _, rule := range rules {
		emitGlueFunction(g, rule)
	}

	// === runAll ===
	g.P("function Generated.runAll(input: SourceTables.SourceTablesType, output: TargetTables.TargetTablesType)")
	for _, rule := range rules {
		g.P("\tGenerated.run", rule.messageName, "(input, output)")
	}
	g.P("end")
	g.P()
	g.P("return Generated")
}

// emitGlueFunction emits a single glue function for a mapping rule.
func emitGlueFunction(g *protogen.GeneratedFile, rule mappingRule) {
	funcName := transformFuncName(rule.messageName)

	g.P("function Generated.run", rule.messageName, "(input: SourceTables.SourceTablesType, output: TargetTables.TargetTablesType)")

	// Build the input struct
	g.P("\tlocal transformInput = {")
	for _, s := range rule.sources {
		g.P("\t\t", s.structField, " = ", buildExtraction(s.pathParts), ",")
	}
	g.P("\t}")
	g.P()

	// Call transform
	g.P("\tlocal result = Transforms.", funcName, "(transformInput)")

	// Unpack output struct to target fields
	for _, t := range rule.targets {
		targetAccess := buildTargetAccess(t.pathParts)
		g.P("\tif result.", t.structField, " ~= nil then")
		g.P("\t\t", targetAccess, " = result.", t.structField)
		g.P("\tend")
	}

	g.P("end")
	g.P()
}
