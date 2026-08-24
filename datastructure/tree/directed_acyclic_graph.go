package main

import (
	"fmt"
)

// DAG 结构体表示有向无环图
type DAG struct {
	nodes map[string]*Node
}

// Node 结构体表示图中的节点
type Node struct {
	name     string
	children []*Node
}

// 添加节点到图中
func (dag *DAG) addNode(name string) {
	if dag.nodes == nil {
		dag.nodes = make(map[string]*Node)
	}

	if _, exists := dag.nodes[name]; !exists {
		dag.nodes[name] = &Node{
			name:     name,
			children: make([]*Node, 0),
		}
	}
}

// 添加边连接两个节点
func (dag *DAG) addEdge(from, to string) {
	fromNode := dag.nodes[from]
	toNode := dag.nodes[to]

	if fromNode == nil || toNode == nil {
		return
	}

	fromNode.children = append(fromNode.children, toNode)
}

// 拓扑排序算法
func (dag *DAG) topologicalSort() []string {
	visited := make(map[string]bool)
	order := make([]string, 0)

	var visit func(node *Node)
	visit = func(node *Node) {
		if visited[node.name] {
			return
		}

		visited[node.name] = true

		for _, child := range node.children {
			visit(child)
		}

		order = append(order, node.name)
	}

	for _, node := range dag.nodes {
		visit(node)
	}

	// 将结果反转，得到拓扑排序的顺序
	for i, j := 0, len(order)-1; i < j; i, j = i+1, j-1 {
		order[i], order[j] = order[j], order[i]
	}

	return order
}

func main() {
	dag := &DAG{}

	// 添加节点
	dag.addNode("A")
	dag.addNode("B")
	dag.addNode("C")
	dag.addNode("D")
	dag.addNode("E")

	// 添加边连接节点
	dag.addEdge("A", "B")
	dag.addEdge("A", "C")
	dag.addEdge("B", "D")
	dag.addEdge("C", "D")
	dag.addEdge("D", "E")

	// 执行拓扑排序
	order := dag.topologicalSort()

	// 打印拓扑排序结果
	fmt.Println("拓扑排序结果:")
	for _, node := range order {
		fmt.Println(node)
	}
}
