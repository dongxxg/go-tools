package main

import "fmt"

type node struct {
	Parent   *node
	Type     string
	Value    interface{}
	Children []*node
}

func newNode(parent *node, typ string, value interface{}) *node {
	return &node{
		Parent:   parent,
		Type:     typ,
		Value:    value,
		Children: []*node{},
	}
}

func (n *node) AddChild(child *node) {
	n.Children = append(n.Children, child)
}

var NodeTemplate = &node{
	Type: "survey",
	Children: []*node{
		&node{
			Type: "catalog",
			Children: []*node{
				&node{
					Type: "question",
					Children: []*node{
						&node{
							Type:     "option",
							Children: make([]*node, 0),
						},
					},
				},
			},
		},
	},
}

func main() {
	n := NodeTemplate
	fmt.Println(n.Type)
	for len(n.Children) != 0 {
		n = n.Children[0]
		fmt.Println(n.Type)
	}
}
