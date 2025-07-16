package tree

import (
	"encoding/json"
	"fmt"
	"testing"
)

type TestNode struct {
	Id         int    `json:"id"`
	Name       string `json:"name"`
	Type       int    `json:"type"`
	ParentId   int    `json:"parentId"`
	ParentType int    `json:"parentType"`
}

func TestBuildTree(t *testing.T) {
	nodes := []*TestNode{
		{
			Id:         2,
			Name:       "机构2",
			Type:       0,
			ParentId:   0,
			ParentType: 0,
		},
		{
			Id:         4,
			Name:       "文件夹2",
			Type:       2,
			ParentId:   2,
			ParentType: 1,
		},
		{
			Id:         1,
			Name:       "机构1",
			Type:       0,
			ParentId:   0,
			ParentType: 0,
		},
		{
			Id:         2,
			Name:       "备课组",
			Type:       1,
			ParentId:   2,
			ParentType: 0,
		},
		{
			Id:         3,
			Name:       "文件夹1",
			Type:       2,
			ParentId:   2,
			ParentType: 1,
		},
		{
			Id:         15,
			Name:       "文件夹3",
			ParentId:   3,
			ParentType: 2,
		},
		{
			Id:         6,
			Name:       "文件夹6",
			ParentId:   3,
			ParentType: 2,
		},
	}
	tree := BuildTreeWithUniqueKey(
		nodes,
		func(node *TestNode) string {
			return fmt.Sprintf("%d_%d", node.Id, node.Type)
		},
		func(node *TestNode) string {
			return fmt.Sprintf("%d_%d", node.ParentId, node.ParentType)
		},
	)

	PrintTree(tree, 0)
	// &{2 机构2 0 0 0}
	//   &{2 备课组 1 2 0}
	//     &{4 文件夹2 2 2 1}
	//     &{3 文件夹1 2 2 1}
	//       &{15 文件夹3 0 3 2}
	//       &{6 文件夹6 0 3 2}
	// &{1 机构1 0 0 0}

	fmt.Println()

	SortTree(tree, func(a, b *Tree[*TestNode]) bool {
		return a.Node.Id < b.Node.Id
	})
	PrintTree(tree, 0)
	// &{1 机构1 0 0 0}
	// &{2 机构2 0 0 0}
	//   &{2 备课组 1 2 0}
	//     &{3 文件夹1 2 2 1}
	//       &{6 文件夹6 0 3 2}
	//       &{15 文件夹3 0 3 2}
	//     &{4 文件夹2 2 2 1}

	fmt.Println()

	SortTreeParallel(tree, func(a, b *Tree[*TestNode]) bool {
		return a.Node.Id < b.Node.Id
	})
	PrintTree(tree, 0)
	// &{1 机构1 0 0 0}
	// &{2 机构2 0 0 0}
	//   &{2 备课组 1 2 0}
	//     &{3 文件夹1 2 2 1}
	//       &{6 文件夹6 0 3 2}
	//       &{15 文件夹3 0 3 2}
	//     &{4 文件夹2 2 2 1}
	marshal, err := json.Marshal(tree)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(marshal))
}
