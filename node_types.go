package hypertune

import "slices"

type BoolNode struct {
	node *Node
}

func NewBoolNode(node *Node) *BoolNode {
	return &BoolNode{node: node}
}

func (n *BoolNode) Get(fallback bool) bool {
	result, err := n.node.Evaluate()
	if err != nil {
		return fallback
	}

	boolVal, ok := result.(bool)
	if !ok {
		n.node.LogUnexpectedValueError(result)
		return fallback
	}

	return boolVal
}

type IntNode struct {
	node *Node
}

func NewIntNode(node *Node) *IntNode {
	return &IntNode{node: node}
}

func (n *IntNode) Get(fallback int) int {
	result, err := n.node.Evaluate()
	if err != nil {
		return fallback
	}

	switch v := result.(type) {
	case int:
		return v
	case int32:
		return int(v)
	case float32:
		return int(v)
	case int64:
		return int(v)
	case float64:
		return int(v)
	default:
		n.node.LogUnexpectedValueError(result)
		return fallback
	}
}

type FloatNode struct {
	node *Node
}

func NewFloatNode(node *Node) *FloatNode {
	return &FloatNode{node: node}
}

func (n *FloatNode) Get(fallback float64) float64 {
	result, err := n.node.Evaluate()
	if err != nil {
		return fallback
	}

	switch v := result.(type) {
	case float64:
		return v
	case int:
		return float64(v)
	case int32:
		return float64(v)
	case float32:
		return float64(v)
	case int64:
		return float64(v)
	default:
		n.node.LogUnexpectedValueError(result)
		return fallback
	}
}

type StringNode struct {
	node *Node
}

func NewStringNode(node *Node) *StringNode {
	return &StringNode{node: node}
}

func (n *StringNode) Get(fallback string) string {
	result, err := n.node.Evaluate()
	if err != nil {
		return fallback
	}

	strVal, ok := result.(string)
	if !ok {
		n.node.LogUnexpectedValueError(strVal)
		return fallback
	}

	return strVal
}

type VoidNode struct {
	node *Node
}

func NewVoidNode(node *Node) *VoidNode {
	return &VoidNode{node: node}
}

func (n *VoidNode) Get() {
	result, err := n.node.Evaluate()
	if err != nil {
		return
	}

	boolVal, ok := result.(bool)
	if ok && boolVal {
		return
	}

	n.node.LogUnexpectedValueError(boolVal)
}

type EnumNode[T ~string] struct {
	node          *Node
	allowedValues []T
}

func NewEnumNode[T ~string](allowedValues []T, node *Node) *EnumNode[T] {
	return &EnumNode[T]{node: node, allowedValues: allowedValues}
}

func (n *EnumNode[T]) Get(fallback T) T {
	result, err := n.node.Evaluate()
	if err != nil {
		return fallback
	}

	strVal, ok := result.(string)
	if !ok || strVal == "" || !slices.Contains(n.allowedValues, (T)(strVal)) {
		n.node.LogUnexpectedValueError(strVal)
		return fallback
	}

	return (T)(strVal)
}
