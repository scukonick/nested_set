package nestedset

import (
	"database/sql"
	"errors"
	"log"
)

// ErrNodeDoesNotExist is returned by
// GetNodeByValue if such node does not exist.
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

	query := `SELECT t.id, t.left_key, t.right_key, t.value
  FROM tree as t
  WHERE t.left_key = 1`

	err := t.DB.QueryRow(query).Scan(&n.ID, &n.LeftKey, &n.RightKey, &n.Value)
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
  (left_key, right_key, value)
  VALUES (1,2, $1)
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
  SELECT t.id, t.left_key, t.right_key, t.value
  FROM tree as t
  ORDER BY left_key
  `
	rows, err := t.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []*Node
	for rows.Next() {
		n := &Node{}
		err = rows.Scan(&n.ID, &n.LeftKey, &n.RightKey, &n.Value)
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
  SELECT t.id, t.left_key, t.right_key, t.value
  FROM tree as t
  WHERE t.value = $1
  ORDER BY left_key LIMIT 1`
	n := &Node{}

	err := t.DB.QueryRow(query, value).Scan(&n.ID, &n.LeftKey, &n.RightKey, &n.Value)
	if err == sql.ErrNoRows {
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
	(left_key, right_key, value)
	VALUES ($1, $2, $3)
	RETURNING id
	`
	err = tx.QueryRow(query, n.LeftKey, n.RightKey, n.Value).Scan(&n.ID)

	return n, err
}

// DeleteNode deletes node n and all it's children
// from the tree. It returns non-nil error
// in case if something goes wrong
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

	// Update parent branch
	leftLim := n.RightKey - n.LeftKey + 1

	query = `
	UPDATE tree SET
	right_key = right_key - $1
	WHERE right_key > $3
	AND left_key < $2`
	_, err = tx.Exec(query, leftLim, n.LeftKey, n.RightKey)
	if err != nil {
		return
	}

	// Update next nodes
	query = `
	UPDATE tree SET
	left_key = left_key - $1,
	right_key = right_key - $1
	WHERE left_key > $2
	`
	_, err = tx.Exec(query, leftLim, n.RightKey)
	return
}

// GetParent returns parent of node. If there is no parent node
// it returns ErrNodeDoesNotExist error. It also returns this error
// for root node, so user of this function should check if input node
// is root node himself
func (t *Tree) GetParent(node *Node) (*Node, error) {
	query := `
	SELECT id, left_key, right_key, value
	FROM tree WHERE
	left_key < $1 AND right_key > $2
	ORDER BY left_key DESC LIMIT 1`
	parent := &Node{}

	err := t.DB.QueryRow(query, node.LeftKey, node.RightKey).Scan(
		&parent.ID, &parent.LeftKey, &parent.RightKey, &parent.Value)
	if err != nil && err == sql.ErrNoRows {
		return parent, ErrNodeDoesNotExist
	}
	return parent, err

}

// IsDescendantOf checks if node child is really descendant of node parent
func IsDescendantOf(child, parent *Node) bool {
	if child.LeftKey > parent.LeftKey && child.RightKey < parent.RightKey {
		return true
	}
	return false
}

// MoveNode moves node to new parent newParent.
// It refuses to move node to it's descendant or move root node
// and returns error in that case.
// It does nothing and returns nil error in case of trying
// to move node to itself.
func (t *Tree) MoveNode(newParent, node *Node) error {
	width := node.RightKey - node.LeftKey + 1
	distance := node.RightKey - newParent.RightKey

	// Doing checks if operation is possible
	if distance == 0 {
		return nil
	}

	if IsDescendantOf(newParent, node) {
		return errors.New("Could not move node to it's own descendant")
	}

	_, err := t.GetParent(node)
	if err != nil && err == ErrNodeDoesNotExist {
		return errors.New("Not possible to move orphan node (or root node)")
	}

	// Let's start transaction!
	tx, err := t.DB.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
			return
		}
		err = tx.Commit()
	}()

	log.Printf("Removing our branch")
	query := `
	UPDATE tree SET
	left_key = -left_key, right_key = -right_key
	WHERE left_key >= $1 AND right_key <= $2`
	_, err = tx.Exec(query, node.LeftKey, node.RightKey)
	if err != nil {
		return err
	}

	log.Print("Decrease right key for nodes after removal and parents")
	query = `
	UPDATE tree SET
	right_key = right_key - $2
	WHERE
	right_key > $1`
	_, err = tx.Exec(query, node.RightKey, width)
	if err != nil {
		return err
	}

	log.Printf("Decrease left key for nodes after removal")
	query = `
	UPDATE tree SET
	left_key = left_key - $2
	WHERE left_key > $1`
	_, err = tx.Exec(query, node.LeftKey, width)
	if err != nil {
		return err
	}

	newParentUpdated := newParent.RightKey - width
	if distance > 0 {
		newParentUpdated = newParent.RightKey
	}
	log.Printf("Increasing right key after the place of insertion and new parents")
	query = `
	UPDATE tree SET
	right_key = right_key + $2
	WHERE right_key >= $1`
	_, err = tx.Exec(query, newParentUpdated, width)
	if err != nil {
		return err
	}

	log.Printf("Increasing left key after the place of insertion")
	query = `
	UPDATE tree SET
	left_key = left_key + $2
	WHERE left_key > $1`
	_, err = tx.Exec(query, newParentUpdated, width)
	if err != nil {
		return err
	}

	var d int32
	if distance > 0 {
		newParentRK := newParent.RightKey + width
		d = node.RightKey - newParentRK + 1
	} else {
		d = distance + 1
	}
	log.Print("Actually moving our branch")
	query = `
	UPDATE tree SET
	right_key = -right_key - $1,
	left_key = -left_key - $1
	WHERE left_key <= 0`
	_, err = tx.Exec(query, d)
	if err != nil {
		return err
	}

	return nil
}

// RenameNode updates value of node. It returns non-nil error
// in case if something goes wrong
func (t *Tree) RenameNode(node *Node, newName string) error {
	query := `
	UPDATE tree SET
	value = $1 WHERE id = $2`
	_, err := t.DB.Exec(query, newName, node.ID)
	return err
}
