package main

import (
	"fmt"
	"path"
	"sort"
	"strings"
	"unicode"

	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"

	"protoc-gen-luau/mappingoptions"
)

// mappingField represents a single source or target field in a mapping rule.
type mappingField struct {
	rawPath   string   // original annotation value e.g. "master.some_field"
	pathParts []string // split by ".", each part converted to camelCase
	isSource  bool
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

func parseMappingField(rawPath string, isSource bool) mappingField {
	parts := strings.Split(rawPath, ".")
	camelParts := make([]string, len(parts))
	for i, p := range parts {
		camelParts[i] = snakeToCamel(p)
	}
	return mappingField{
		rawPath:   rawPath,
		pathParts: camelParts,
		isSource:  isSource,
	}
}

func transformFuncName(messageName string) string {
	if messageName == "" {
		return ""
	}
	runes := []rune(messageName)
	runes[0] = unicode.ToLower(runes[0])
	return string(runes)
}

// --- Input type tree ---
// Sources are grouped into a nested tree by namespace.
// 2-segment: master.field → { master: { field: leaf } }
// 3-segment: master.compound.sub → { master: { compound: { valueSubFields: { sub: leaf } } } }

type typeNode struct {
	children map[string]*typeNode // non-nil if branch
	isLeaf   bool                 // true if this is a terminal field
}

func newBranch() *typeNode {
	return &typeNode{children: make(map[string]*typeNode)}
}

func newLeaf() *typeNode {
	return &typeNode{isLeaf: true}
}

// buildSourceTree builds a nested tree from source fields.
func buildSourceTree(sources []mappingField) *typeNode {
	root := newBranch()
	for _, s := range sources {
		parts := s.pathParts // e.g. ["master", "someField"] or ["master", "wireInstructions", "asaBankname"]
		if len(parts) == 2 {
			// namespace.field
			ns := parts[0]
			field := parts[1]
			if root.children[ns] == nil {
				root.children[ns] = newBranch()
			}
			root.children[ns].children[field] = newLeaf()
		} else if len(parts) == 3 {
			// namespace.compound.subField → namespace.compound.valueSubFields.subField
			ns := parts[0]
			compound := parts[1]
			subField := parts[2]
			if root.children[ns] == nil {
				root.children[ns] = newBranch()
			}
			nsNode := root.children[ns]
			if nsNode.children[compound] == nil {
				nsNode.children[compound] = newBranch()
			}
			compNode := nsNode.children[compound]
			if compNode.children["valueSubFields"] == nil {
				compNode.children["valueSubFields"] = newBranch()
			}
			compNode.children["valueSubFields"].children[subField] = newLeaf()
		}
	}
	return root
}

// emitTypeTree emits the Luau type definition for a nested tree.
func emitTypeTree(g *protogen.GeneratedFile, node *typeNode, indent string) {
	// Sort keys for deterministic output
	keys := make([]string, 0, len(node.children))
	for k := range node.children {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		child := node.children[k]
		if child.isLeaf {
			g.P(indent, k, ": DataTypes.StringType?,")
		} else {
			g.P(indent, k, ": {")
			emitTypeTree(g, child, indent+"\t")
			g.P(indent, "}?,")
		}
	}
}

// --- Glue extraction ---
// Builds nested Luau table for the transformInput, extracting from the real source tables.

func emitExtractionTree(g *protogen.GeneratedFile, node *typeNode, accessPrefix string, indent string) {
	keys := make([]string, 0, len(node.children))
	for k := range node.children {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		child := node.children[k]
		if child.isLeaf {
			g.P(indent, k, " = ", accessPrefix, " and ", accessPrefix, ".", k, " or nil,")
		} else {
			// Check if parent exists before building sub-table
			g.P(indent, k, " = if ", accessPrefix, " and ", accessPrefix, ".", k, " then {")
			emitExtractionTree(g, child, accessPrefix+"."+k, indent+"\t")
			g.P(indent, "} else nil,")
		}
	}
}

// --- Target field name (strips "target." prefix) ---

func targetFieldName(pathParts []string) string {
	// ["target", "sfSomeField"] → "sfSomeField"
	if len(pathParts) >= 2 {
		return strings.Join(pathParts[1:], ".")
	}
	return strings.Join(pathParts, ".")
}

func buildTargetAccess(pathParts []string) string {
	return "output." + strings.Join(pathParts, ".")
}

// --- Main generator ---

func generateMapping(plugin *protogen.Plugin, file *protogen.File) {
	protoPath := file.Desc.Path()
	baseName := strings.TrimSuffix(path.Base(protoPath), ".proto")

	rules := parseMappingRules(file)
	if len(rules) == 0 {
		return
	}

	transformsModule := strings.Replace(baseName, "_mappings", "_transforms", 1)
	transformsRequire := "@lib/" + transformsModule

	// === Types file ===
	gt := plugin.NewGeneratedFile(baseName+"_types.luau", "")
	gt.P("-- GENERATED by protoc-gen-luau \u2014 DO NOT EDIT")
	gt.P("--!strict")
	gt.P(`local DataTypes = require("@lib/data_types")`)
	gt.P()
	gt.P("local Types = {}")
	gt.P()

	for _, rule := range rules {
		// Input type — nested by namespace
		tree := buildSourceTree(rule.sources)
		gt.P("export type ", rule.messageName, "Input = {")
		emitTypeTree(gt, tree, "\t")
		gt.P("}")
		gt.P()

		// Output type — flat (target fields)
		gt.P("export type ", rule.messageName, "Output = {")
		for _, t := range rule.targets {
			gt.P("\t", targetFieldName(t.pathParts), ": DataTypes.StringType?,")
		}
		gt.P("}")
		gt.P()
	}

	gt.P("return Types")

	// === Generated file ===
	g := plugin.NewGeneratedFile(baseName+"_generated.luau", "")
	g.P("-- GENERATED by protoc-gen-luau \u2014 DO NOT EDIT")
	g.P("--!strict")
	g.P(fmt.Sprintf("local DataTypes = require(%q)", getDataTypesRequirePath(file)))
	g.P(`local SourceTables = require("@lib/source_tables")`)
	g.P(`local TargetTables = require("@lib/target_tables")`)
	g.P(fmt.Sprintf("local Transforms = require(%q)", transformsRequire))
	g.P()
	g.P("local Generated = {}")
	g.P()

	for _, rule := range rules {
		emitGlueFunction(g, rule)
	}

	g.P("function Generated.runAll(input: SourceTables.SourceTablesType, output: TargetTables.TargetTablesType)")
	for _, rule := range rules {
		g.P("\tGenerated.run", rule.messageName, "(input, output)")
	}
	g.P("end")
	g.P()
	g.P("return Generated")
}

func emitGlueFunction(g *protogen.GeneratedFile, rule mappingRule) {
	funcName := transformFuncName(rule.messageName)
	tree := buildSourceTree(rule.sources)

	g.P("function Generated.run", rule.messageName, "(input: SourceTables.SourceTablesType, output: TargetTables.TargetTablesType)")

	// Build nested input struct
	g.P("\tlocal transformInput = {")
	emitExtractionTree(g, tree, "input", "\t\t")
	g.P("\t}")
	g.P()

	// Call transform
	g.P("\tlocal result = Transforms.", funcName, "(transformInput)")

	// Unpack output
	for _, t := range rule.targets {
		tfn := targetFieldName(t.pathParts)
		targetAccess := buildTargetAccess(t.pathParts)
		g.P("\tif result.", tfn, " ~= nil then")
		g.P("\t\t", targetAccess, " = result.", tfn)
		g.P("\tend")
	}

	g.P("end")
	g.P()
}
