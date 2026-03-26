package main

import (
	"fmt"

	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// topoSortMessages sorts messages in dependency order using Kahn's algorithm.
// A message depends on another if it has a field whose type is a message
// defined in the same file. Returns error on circular dependency.
func topoSortMessages(msgs []*protogen.Message, fileDesc protoreflect.FileDescriptor) ([]*protogen.Message, error) {
	filePath := fileDesc.Path()

	nameIdx := make(map[string]int, len(msgs))
	for i, m := range msgs {
		nameIdx[m.GoIdent.GoName] = i
	}

	inDegree := make([]int, len(msgs))
	dependents := make([][]int, len(msgs))

	for i, m := range msgs {
		for _, field := range m.Fields {
			if field.Desc.Kind() != protoreflect.MessageKind {
				continue
			}
			if field.Message.Desc.ParentFile().Path() != filePath {
				continue
			}
			depName := field.Message.GoIdent.GoName
			if j, ok := nameIdx[depName]; ok && j != i {
				dependents[j] = append(dependents[j], i)
				inDegree[i]++
			}
		}
	}

	var queue []int
	for i, deg := range inDegree {
		if deg == 0 {
			queue = append(queue, i)
		}
	}

	var sorted []*protogen.Message
	for len(queue) > 0 {
		curr := queue[0]
		queue = queue[1:]
		sorted = append(sorted, msgs[curr])
		for _, dep := range dependents[curr] {
			inDegree[dep]--
			if inDegree[dep] == 0 {
				queue = append(queue, dep)
			}
		}
	}

	if len(sorted) != len(msgs) {
		return nil, fmt.Errorf("circular dependency detected among messages in %s", filePath)
	}
	return sorted, nil
}
