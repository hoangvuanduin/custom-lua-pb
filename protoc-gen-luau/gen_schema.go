package main

import (
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
	g.P("--!strict")
	g.P(`local DataTypes = require("`, requirePath, `")`)
	g.P()
	g.P("local ", moduleName, " = {}")
	g.P()
	g.P("return ", moduleName)
}
