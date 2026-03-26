package main

import (
	"fmt"

	"google.golang.org/protobuf/compiler/protogen"
)

func generate(plugin *protogen.Plugin) error {
	for _, f := range plugin.Files {
		if !f.Generate {
			continue
		}

		genType := getGeneratorType(f)
		if genType == "" {
			continue
		}

		gen, err := lookupGenerator(genType)
		if err != nil {
			return fmt.Errorf("%s: %w", f.Desc.Path(), err)
		}
		gen.Generate(plugin, f)
	}
	return nil
}
