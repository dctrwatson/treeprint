// Package treeprint provides a simple ASCII tree composing tool.
package treeprint

import (
	"bytes"
	"fmt"
	"reflect"
)

type Value interface{}
type MetaValue interface{}

// Tree represents a tree structure with leaf-nodes and branch-nodes.
type Tree interface {
	// AddNode adds a new node to a branch.
	AddNode(v Value) Tree
	// AddMetaNode adds a new node with meta value provided to a branch.
	AddMetaNode(meta MetaValue, v Value) Tree
	// AddBranch adds a new branch node (a level deeper).
	AddBranch(v Value) Tree
	// AddMetaBranch adds a new branch node (a level deeper) with meta value provided.
	AddMetaBranch(meta MetaValue, v Value) Tree
	// Branch converts a leaf-node to a branch-node,
	// applying this on a branch-node does no effect.
	Branch() Tree
	// FindByMeta finds a node whose meta value matches the provided one by reflect.DeepEqual,
	// returns nil if not found.
	FindByMeta(meta MetaValue) Tree
	// FindByValue finds a node whose value matches the provided one by reflect.DeepEqual,
	// returns nil if not found.
	FindByValue(value Value) Tree
	//  returns the last node of a tree
	FindLastNode() Tree
	// String renders the tree or subtree as a string.
	String() string
	// Bytes renders the tree or subtree as byteslice.
	Bytes() []byte

	GetValue() Value
	SetValue(value Value)
	GetMetaValue() MetaValue
	SetMetaValue(meta MetaValue)

	Walk(TreeWalkFn) error
}

type TreeWalkFn func(v *Vertex, level int) error

type Vertex struct {
	*node
	Level int
}

type node struct {
	Root  *node
	Meta  MetaValue
	Value Value
	Nodes []*node
}

func (n *node) FindLastNode() Tree {
	ns := n.Nodes
	n = ns[len(ns)-1]
	return n
}

func (n *node) AddNode(v Value) Tree {
	n.Nodes = append(n.Nodes, &node{
		Root:  n,
		Value: v,
	})
	if n.Root != nil {
		return n.Root
	}
	return n
}

func (n *node) AddMetaNode(meta MetaValue, v Value) Tree {
	n.Nodes = append(n.Nodes, &node{
		Root:  n,
		Meta:  meta,
		Value: v,
	})
	if n.Root != nil {
		return n.Root
	}
	return n
}

func (n *node) AddBranch(v Value) Tree {
	branch := &node{
		Value: v,
	}
	n.Nodes = append(n.Nodes, branch)
	return branch
}

func (n *node) AddMetaBranch(meta MetaValue, v Value) Tree {
	branch := &node{
		Meta:  meta,
		Value: v,
	}
	n.Nodes = append(n.Nodes, branch)
	return branch
}

func (n *node) Branch() Tree {
	n.Root = nil
	return n
}

func (n *node) FindByMeta(meta MetaValue) Tree {
	for _, node := range n.Nodes {
		if reflect.DeepEqual(node.Meta, meta) {
			return node
		}
		if v := node.FindByMeta(meta); v != nil {
			return v
		}
	}
	return nil
}

func (n *node) FindByValue(value Value) Tree {
	for _, node := range n.Nodes {
		if reflect.DeepEqual(node.Value, value) {
			return node
		}
		if v := node.FindByMeta(value); v != nil {
			return v
		}
	}
	return nil
}

func (n *node) Walk(walkFn TreeWalkFn) error {
	vertices := []*Vertex{&Vertex{n, 0}}

	for len(vertices) > 0 {
		n := len(vertices)
		v := vertices[n-1]
		vertices = vertices[:n-1]

		if err := walkFn(v, v.Level); err != nil {
			return err
		}

		for i := len(v.node.Nodes) - 1; i >= 0; i-- {
			vertices = append(vertices, &Vertex{v.node.Nodes[i], v.Level + 1})
		}
	}

	return nil
}

func (n *node) Bytes() []byte {
	buf := new(bytes.Buffer)
	levelSize := map[int]int{
		1: len(n.Nodes),
	}

	if n.Root == nil {
		if n.Meta != nil {
			buf.WriteString(fmt.Sprintf("[%v]  %v", n.Meta, n.Value))
		} else {
			buf.WriteString(fmt.Sprintf("%v", n.Value))
		}
		buf.WriteByte('\n')
	}

	n.Walk(func(v *Vertex, level int) error {
		// Already did the 0-th node
		if level == 0 {
			return nil
		}
		// Decrement counter for current level
		levelSize[level]--
		// Save counter for next level
		if len(v.Nodes) > 0 {
			levelSize[level+1] = len(v.Nodes)
		}

		// If there are no more nodes at this level, use end edge
		var edge EdgeType
		if levelSize[level] == 0 {
			edge = EdgeTypeEnd
		} else {
			edge = EdgeTypeMid
		}

		// For every level, indent
		for i := 1; i < level; i++ {
			if levelSize[i] > 0 {
				// If level has nodes, continue its link
				fmt.Fprintf(buf, "%s%c%c ", EdgeTypeLink, '\u00A0', '\u00A0')
			} else {
				// If not, just print empty padding
				fmt.Fprint(buf, "    ")
			}
		}

		if v.Meta != nil {
			fmt.Fprintf(buf, "%s [%v]  %v\n", edge, v.Meta, v.Value)
		} else {
			fmt.Fprintf(buf, "%s %v\n", edge, v.Value)
		}
		return nil
	})

	return buf.Bytes()
}

func (n *node) String() string {
	return string(n.Bytes())
}

func (n *node) SetValue(value Value) {
	n.Value = value
}

func (n *node) GetValue() Value {
	return n.Value
}

func (n *node) SetMetaValue(meta MetaValue) {
	n.Meta = meta
}

func (n *node) GetMetaValue() MetaValue {
	return n.Meta
}

type EdgeType string

var (
	EdgeTypeLink EdgeType = "│"
	EdgeTypeMid  EdgeType = "├──"
	EdgeTypeEnd  EdgeType = "└──"
)

func New() Tree {
	return &node{Value: "."}
}
