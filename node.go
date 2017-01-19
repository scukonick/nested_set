package main

// Node represents a node of nested set tree
type Node struct {
	ID       int32
	LeftKey  int32
	RightKey int32
	Level    int32
	Value    string
}

// NeWNode returns pointer to newly initialized Node
func NeWNode(id, leftKey, rightKey, level int32, value string) *Node {
	n := &Node{
		ID:       id,
		LeftKey:  leftKey,
		RightKey: rightKey,
		Level:    level,
		Value:    value,
	}
	return n
}
