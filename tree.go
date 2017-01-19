package main

import (
	"database/sql"
	"errors"
	"log"
)

// ErrNodeDoesNotExist is returned by functions
// GetNodeByValue if suckh node does not exist.
var ErrNodeDoesNotExist = errors.New("sql: node does not exist")

// Tree represents the whole tree structure in the database
// and provides methods to work with it
type Tree struct {
	Root *Node
	DB   *sql.DB
}

// NewTree returns pointer to newly initialized Tree
func NewTree(db *sql.DB) *Tree {
	t := &Tree{
		DB: db,
	}
	return t
}

// IsPopulated tries to find root node in the database.
// If t.Root is not nil it does not do anything and returns true.
// Else, if node with left_key = 1 does not exist, it returns false
// If it exists, it sets t.Root to point to it
func (t *Tree) IsPopulated() (bool, error) {
	if t.Root != nil {
		return true, nil
	}

	n := &Node{}

	query := `SELECT t.id, t.left_key, t.right_key, t.level, t.value
  FROM tree as t
  WHERE t.left_key = 1`

	err := t.DB.QueryRow(query).Scan(&n.ID, &n.LeftKey, &n.RightKey, &n.Level, &n.Value)
	if err != nil && err == sql.ErrNoRows {
		return false, nil
	} else if err != nil {
		return false, err
	}
	t.Root = n
	return true, nil
}

// Plant inserts root node of the tree to the database
// and sets t.Root to point to the root node.
// It does nothing and returns nil error if t.Root is not nil.
// It returns error in case if something is wrong.
func (t *Tree) Plant(value string) error {
	query := `INSERT INTO tree
  (left_key, right_key, level, value)
  VALUES (1,2, 0, $1)
  RETURNING id`
	var id int32
	err := t.DB.QueryRow(query, value).Scan(&id)
	if err != nil {
		return err
	}
	t.Root = &Node{
		ID:       id,
		LeftKey:  1,
		RightKey: 2,
		Level:    0,
		Value:    value,
	}

	return nil
}

// GetAllNodes traverses all the tree and returns all its' nodes
func (t *Tree) GetAllNodes() ([]*Node, error) {
	query := `
  SELECT t.id, t.left_key, t.right_key, t.level, t.value
  FROM tree as t
  ORDER BY left_key
  `
	rows, err := t.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]*Node, 0)
	for rows.Next() {
		n := &Node{}
		err = rows.Scan(&n.ID, &n.LeftKey, &n.RightKey, &n.Level, &n.Value)
		if err != nil {
			return nil, err
		}
		result = append(result, n)
	}
	return result, nil
}

// GetNodeByValue search node by value in the tree and returns
// a pointer to it. If there is more than one node with the same value
// it returns the first found (ordered by left_key).
// If something goes wrong it returns non-nil error
func (t *Tree) GetNodeByValue(value string) (*Node, error) {
	query := `
  SELECT t.id, t.left_key, t.right_key, t.level, t.value
  FROM tree as t
  WHERE t.value = $1
  ORDER BY left_key LIMIT 1`
	n := &Node{}

	err := t.DB.QueryRow(query, value).Scan(&n.ID, &n.LeftKey, &n.RightKey, &n.Level, &n.Value)
	if err != nil && err == sql.ErrNoRows {
		return nil, ErrNodeDoesNotExist
	} else if err != nil {
		return nil, err
	}

	return n, nil
}

// InsertChild creates new node with value and
// inserts it as child of the parent node.
// If something goes wrong it returns non-nil error.
// Please not that values of left_key, right_key in the parent node
// would be outdated after this operation.
func (t *Tree) InsertChild(parent *Node, value string) (*Node, error) {
	// Update other childs of partner
	tx, err := t.DB.Begin()
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
			return
		}
		err = tx.Commit()
	}()

	query := `
	UPDATE tree SET
	left_key = left_key + 2,
	right_key = right_key + 2
	WHERE left_key > $1`

	_, err = tx.Exec(query, parent.RightKey)
	if err != nil {
		return nil, err
	}

	// Update parent's branch
	query = `
	UPDATE tree SET
	right_key = right_key + 2
	WHERE right_key >= $1 AND left_key < $1
	`
	_, err = tx.Exec(query, parent.RightKey)
	if err != nil {
		return nil, err
	}

	// Insert new node
	n := &Node{
		Value:    value,
		LeftKey:  parent.RightKey,
		RightKey: parent.RightKey + 1,
		Level:    parent.Level + 1,
	}
	query = `
	INSERT INTO tree
	(left_key, right_key, level, value)
	VALUES ($1, $2, $3, $4)
	RETURNING id
	`
	err = tx.QueryRow(query, n.LeftKey, n.RightKey, n.Level, n.Value).Scan(&n.ID)

	return n, err
}

func (t *Tree) DeleteNode(n *Node) (err error) {
	tx, err := t.DB.Begin()
	if err != nil {
		return
	}
	defer func() {
		if err != nil {
			tx.Rollback()
			return
		}
		err = tx.Commit()
	}()

	// Removing node and all it's children
	query := `
	DELETE FROM tree WHERE
	left_key >= $1 AND
	right_key <= $2`
	_, err = tx.Exec(query, n.LeftKey, n.RightKey)
	if err != nil {
		return
	}
	t.showTree()

	// Update parent branch
	query = `
	UPDATE tree SET
	right_key = right_key - ($2 - $1 + 1)
	WHERE right_key > $2
	AND left_key < $1`
	_, err = tx.Exec(query, n.LeftKey, n.RightKey)
	if err != nil {
		return
	}
	t.showTree()

	// Update next nodes
	query = `
	UPDATE tree SET
	left_key = left_key - ($2 - $1 + 1),
	right_key = right_key - ($2 - $1 + 1)
	WHERE left_key > $2
	`
	_, err = tx.Exec(query, n.LeftKey, n.RightKey)
	return
}

func (t *Tree) showTree() error {
	nodes, err := t.GetAllNodes()
	if err != nil {
		return err
	}
	for _, node := range nodes {
		log.Printf("Node: %+v", node)
	}
	return nil
}
