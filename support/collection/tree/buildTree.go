package tree

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"strings"
	"sync"
)

// Tree 树节点结构体
type Tree[T any] struct {
	Node     T
	Children []*Tree[T]
}

// BuildTreeWithUniqueKey 树形结构构建函数
// notes: 节点列表
// getUniqueKey: 生成当前节点唯一key
// getUniqueParentKey: 生成当前的父节点唯一key
func BuildTreeWithUniqueKey[T any, K comparable](
	nodes []T,
	getUniqueKey func(T) K, // 生成节点唯一key
	getUniqueParentKey func(T) K, // 生成父节点唯一key
) []*Tree[T] {

	// 创建节点映射
	nodeMap := make(map[K]*Tree[T], len(nodes))
	for _, node := range nodes {
		key := getUniqueKey(node)
		nodeMap[key] = &Tree[T]{Node: node}
	}

	// 构建树结构
	var roots []*Tree[T]
	for _, node := range nodes {
		currentKey := getUniqueKey(node)
		parentKey := getUniqueParentKey(node)

		// 防止自引用
		if parentKey == currentKey {
			continue
		}

		// 挂载到父节点
		if parent, exists := nodeMap[parentKey]; exists {
			parent.Children = append(parent.Children, nodeMap[currentKey])
		} else {
			roots = append(roots, nodeMap[currentKey])
		}
	}

	return roots
}

// PrintTree 递归打印树结构, 用于快速验证
func PrintTree[T any](nodes []*Tree[T], indent int) {
	for _, node := range nodes {
		fmt.Printf("%s%v\n", strings.Repeat("  ", indent), node.Node)
		PrintTree(node.Children, indent+1)
	}
}

// SortTree 树节点排序函数（深度优先递归）
func SortTree[T any](nodes []*Tree[T], compare func(a, b *Tree[T]) bool) {
	// 对当前层级排序
	sort.SliceStable(nodes, func(i, j int) bool {
		return compare(nodes[i], nodes[j])
	})

	// 递归排序子节点
	for _, node := range nodes {
		if len(node.Children) > 0 {
			SortTree(node.Children, compare)
		}
	}
}

// SortTreeParallel 树节点排序函数（并行）
func SortTreeParallel[T any](nodes []*Tree[T], compare func(a, b *Tree[T]) bool) {
	var wg sync.WaitGroup
	sort.SliceStable(nodes, func(i, j int) bool {
		return compare(nodes[i], nodes[j])
	})

	for _, node := range nodes {
		wg.Add(1)
		go func(n *Tree[T]) {
			defer wg.Done()
			SortTreeParallel(n.Children, compare)
		}(node)
	}
	wg.Wait()
}

// MarshalJSON 使用反射实现自定义 Tree的 JSON 序列化, 这里是为了返回给前端的是一个嵌套结构, 即:children作为node的一个字段, 如:
//
//	{
//	    id: 1,
//	    ...
//	    "children": [
//	        ...
//	    ],
//	}
//
// 如果不自定义MarshalJSON, 由于Tree是一个泛型结构, 没有办法直接实现这种嵌套结构(至少我没找到如何实现), 返回的是:
//
//	{
//	    "node": {
//	        "id": 1,
//	        ...
//	    },
//	    "children": [
//	        ...
//	  ]
//	},
//
// 这种node和children分开的情况, 对于前端来说不是一个标准的树形结构的格式, 还是需要自行遍历树去解析, 所以需要自定义MarshalJSON进行一步转换
func (t *Tree[T]) MarshalJSON() ([]byte, error) {
	// 使用反射动态合并字段
	nodeType := reflect.TypeOf(t.Node).Elem()
	nodeValue := reflect.ValueOf(t.Node).Elem()

	// 创建动态字段映射
	fields := make(map[string]interface{})
	for i := 0; i < nodeType.NumField(); i++ {
		jsonTag := nodeType.Field(i).Tag.Get("json")
		if jsonTag != "" && jsonTag != "-" {
			fields[strings.Split(jsonTag, ",")[0]] = nodeValue.Field(i).Interface()
		}
	}

	// 添加子节点字段
	if len(t.Children) > 0 {
		fields["children"] = t.Children
	}

	return json.Marshal(fields)
}
