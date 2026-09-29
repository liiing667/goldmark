package ast

import (
	gast "github.com/yuin/goldmark/v2/ast"
)

// A Container struct represents a container block directive of Markdown text.
//
// An example of a container block directive:
//
//	::: note optional title
//	contents
//	:::
type Container struct {
	gast.BaseBlock

	// Name is a name of this container(e.g. 'note').
	// This value is a subslice of the source, so it must not be modified.
	Name []byte
}

// Dump implements Node.Dump.
func (n *Container) Dump(_ []byte) *gast.NodeDump {
	return gast.NewNodeDump(n, map[string]any{
		"Name": string(n.Name),
	})
}

// KindContainer is a NodeKind of the Container node.
var KindContainer = gast.NewNodeKind("Container")

// Kind implements Node.Kind.
func (n *Container) Kind() gast.NodeKind {
	return KindContainer
}

// NewContainer returns a new Container node.
func NewContainer(name []byte) *Container {
	n := &Container{Name: name}
	n.Init(n)
	return n
}

// A ContainerTitle struct represents a title of a Container node.
// A ContainerTitle node is always a first child of a Container node and
// its source is parsed as inline elements.
type ContainerTitle struct {
	gast.BaseBlock
}

// Dump implements Node.Dump.
func (n *ContainerTitle) Dump(_ []byte) *gast.NodeDump {
	return gast.NewNodeDump(n, nil)
}

// KindContainerTitle is a NodeKind of the ContainerTitle node.
var KindContainerTitle = gast.NewNodeKind("ContainerTitle")

// Kind implements Node.Kind.
func (n *ContainerTitle) Kind() gast.NodeKind {
	return KindContainerTitle
}

// NewContainerTitle returns a new ContainerTitle node.
func NewContainerTitle() *ContainerTitle {
	n := &ContainerTitle{}
	n.Init(n)
	return n
}
