package main

import (
	"io"
	"log"
	"os"

	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/pluginpb"
)

func main() {
	// Debug mode: capture raw CodeGeneratorRequest for replay testing.
	// Usage: DUMP_REQUEST=testdata/request.pb protoc --luau_out=. ...
	if dumpPath := os.Getenv("DUMP_REQUEST"); dumpPath != "" {
		data, err := io.ReadAll(os.Stdin)
		if err != nil {
			log.Fatalf("reading stdin: %v", err)
		}
		if err := os.WriteFile(dumpPath, data, 0644); err != nil {
			log.Fatalf("writing %s: %v", dumpPath, err)
		}
		// Write empty response so protoc succeeds.
		resp := &pluginpb.CodeGeneratorResponse{}
		out, err := proto.Marshal(resp)
		if err != nil {
			log.Fatalf("marshaling response: %v", err)
		}
		os.Stdout.Write(out)
		return
	}

	protogen.Options{}.Run(func(plugin *protogen.Plugin) error {
		return generate(plugin)
	})
}
