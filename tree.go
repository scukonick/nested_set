package main

import (
	"database/sql"
	"errors"
	"log"
)

// ErrNodeDoesNotExist is returned by functions
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
	(left_key, right_key, value)
	VALUES ($1, $2, $3, $4)
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

func (t *Tree) MoveNode(node, newParent *Node) error {
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

	var newLeftKey int32 = 1
	if newParent != nil {
		newLeftKey = newParent.LeftKey + 1
	}
	log.Printf("New left key: %d", newLeftKey)

	width := node.RightKey - node.LeftKey + 1
	distance := newLeftKey - node.LeftKey
	tmpPos := node.LeftKey
	log.Printf("Width: %d, distance: %d", width, distance)

	if distance < 0 {
		distance -= width
		tmpPos += width
	}

	if distance == 0 {
		// no need to move anything
		return nil
	}

	// Creating space for our node and it's children
	log.Print("Creating space left_key")
	query := `
	UPDATE tree SET
	left_key = left_key + $1
	WHERE left_key >= $2`
	_, err = tx.Exec(query, width, newLeftKey)
	if err != nil {
		return err
	}
	showTxTree(tx)

	log.Print("Creating space right_key")
	query = `
	UPDATE tree SET
	right_key = right_key + $1
	WHERE right_key >= $2`
	_, err = tx.Exec(query, width, newLeftKey)
	if err != nil {
		return err
	}
	showTxTree(tx)

	// Moving our node into new space
	log.Print("Moving node..")
	query = `
	UPDATE tree SET
	left_key = left_key + $1,
	right_key = right_key + $1
  WHERE left_key >= $2
	AND right_key <= $3`
	_, err = tx.Exec(query, distance, tmpPos, node.RightKey)
	if err != nil {
		return err
	}
	showTxTree(tx)

	// Cleaning old space
	log.Print("Cleaning space - left key")
	query = `
	UPDATE tree SET
	left_key = left_key - $1
	WHERE left_key > $2`
	_, err = tx.Exec(query, width, node.RightKey)
	if err != nil {
		return err
	}
	showTxTree(tx)

	log.Print("Cleaning space - right_key")
	query = `
	UPDATE tree SET
	right_key = right_key - $1
	WHERE right_key > $2`
	_, err = tx.Exec(query, width, node.RightKey)
	return err
}

func (t *Tree) MoveNodeA(node, newParent *Node) error {
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

	width := node.RightKey - node.LeftKey + 1

	// 'Remove' our branch
	log.Printf("Removing our branch")
	query := `
	UPDATE tree SET
	left_key = 0-$1, right_key = 0-$2
	WHERE left_key >= $1 AND right_key <= $2`
	_, err = tx.Exec(query, node.LeftKey, node.RightKey)
	if err != nil {
		return err
	}
	showTxTree(tx)

	// decrease left and/or right keys of currently 'lower' items (and parents)
	log.Printf("Decreasing keys of currently lower items (left_key)")
	query = `
	UPDATE tree SET
	left_key = left_key - $2
	WHERE left_key > $1;`
	_, err = tx.Exec(query, node.RightKey, width)
	if err != nil {
		return err
	}
	showTxTree(tx)
	log.Printf("Decreasing keys of currently lower items (right_key)")
	query = `
	UPDATE tree SET
	right_key = right_key - $2
	WHERE right_key > $1;`
	_, err = tx.Exec(query, node.RightKey, width)
	if err != nil {
		return err
	}
	showTxTree(tx)

	// increase left and/or right keys of future 'lower' items (and parents)
	log.Printf("Increasing keys of future lower items (left_key)")
	query = `
	UPDATE tree SET
	left_key = left_key + $2
	WHERE left_key >=
	(CASE WHEN $3::INTEGER > $1 THEN $3::INTEGER - $2 ELSE $1 END)`
	// No idea why postgres is asking to convert $3 to INTEGER
	_, err = tx.Exec(query, node.RightKey, width, newParent.RightKey)
	if err != nil {
		return err
	}
	showTxTree(tx)
	log.Printf("Increasing keys of future lower items (right_key)")
	query = `
	UPDATE tree SET
	right_key = right_key + $2
	WHERE right_key >=
	(CASE WHEN $3::INTEGER > $1 THEN $3::INTEGER - $2 ELSE $1 END)`
	_, err = tx.Exec(query, node.RightKey, width, newParent.RightKey)
	if err != nil {
		return err
	}
	showTxTree(tx)

	// move subtree) and update it's parent item id
	log.Printf("move subtree and update it's parent item id")
	query = `
	UPDATE tree
	SET
	    left_key = 0-(left_key)+
			(CASE WHEN $4 > $2
				THEN $4::INTEGER - $2::INTEGER - 1
				ELSE $4::INTEGER - $2::INTEGER - 1 + $3
				END),
	    right_key = 0-(right_key)+
			(CASE WHEN $4 > $2
				THEN $4::INTEGER - $2::INTEGER - 1
				ELSE $4::INTEGER - $2::INTEGER - 1 + $3
				END)
	WHERE left_key <= 0-($1::INTEGER) AND right_key >= 0-($2::INTEGER)`
	_, err = tx.Exec(query, node.LeftKey, node.RightKey, width, newParent.RightKey)
	if err != nil {
		return err
	}
	showTxTree(tx)
	return nil

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

func showTxTree(tx *sql.Tx) {
	query := `
  SELECT t.id, t.left_key, t.right_key, t.value
  FROM tree as t
  ORDER BY left_key
  `
	rows, _ := tx.Query(query)
	defer rows.Close()

	for rows.Next() {
		n := &Node{}
		rows.Scan(&n.ID, &n.LeftKey, &n.RightKey, &n.Value)
		log.Printf("Node: %d\t%d\t%s", n.LeftKey, n.RightKey, n.Value)
	}
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

func (t *Tree) MyMove(newParent, node *Node) error {
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

	width := node.RightKey - node.LeftKey + 1
	distance := node.RightKey - newParent.RightKey - 1

	if distance == 0 {
		// not moving anything
		return nil
	}

	oldParent, err := t.GetParent(node)
	if err != nil {
		// TODO  - check if our node is root node.
		return err
	}
	log.Printf("Old parent: %+v", oldParent)

	// Decrease right key for all the nodes after removal
	log.Print("Decrease right key for all the next nodes after removal")
	query := `
	UPDATE tree SET
	right_key = right_key - $3
	WHERE
	right_key > $2`
	_, err = tx.Exec(query, node.LeftKey, node.RightKey, width)
	if err != nil {
		return err
	}
	showTxTree(tx)

	// Decrease right key for old parent branch
	log.Printf("Decrease right key for old parent branch")
	query = `
	UPDATE tree SET
	right_key = right_key - $3
	WHERE left_key <= $1 AND
	right_key >= $2`
	_, err = tx.Exec(query, oldParent.LeftKey, oldParent.RightKey, width)
	if err != nil {
		return err
	}
	showTxTree(tx)

	// Actually moving our branch
	log.Print("Actually moving our branch")
	query = `
	UPDATE tree SET
	right_key = right_key - $3,
	left_key = left_key - $3
	WHERE left_key >= $1 AND
	right_key <= $2`
	_, err = tx.Exec(query, node.LeftKey, node.RightKey, distance)
	if err != nil {
		return err
	}
	showTxTree(tx)

	if distance > 0 {
		// If moving left increase both keys
		// for all the next nodes after the place of insertion
		log.Print("Increase both keys for all the next nodes after the place of insertion")
		query = `
		UPDATE tree SET
		right_key = right_key + $3,
		left_key = left_key + $3
		WHERE left_key > $1 AND
		right_key > $2 AND
		NOT (left_key >= $4 AND right_key <= $5)` // ignoring our branch
		_, err = tx.Exec(query, newParent.LeftKey, newParent.RightKey, width,
			node.LeftKey-distance, node.RightKey-distance)
	} else {
		// if moving right decrease both keys
		// for all the next nodes after the place of removal
		log.Printf("Decreasing both keys for all the next nodes after the place of removal")
		query = `
		UPDATE tree SET
		right_key = right_key - $3,
		left_key = left_key - $3
		WHERE left_key > $1 AND
		right_key > $2 AND
		NOT (left_key >= $4 AND right_key <= $5)`
		newRightKey := node.RightKey - distance
		newLeftKey := node.LeftKey - distance
		_, err = tx.Exec(query, node.LeftKey, node.RightKey, width,
			newLeftKey, newRightKey)
	}
	if err != nil {
		return err
	}
	showTxTree(tx)

	err = errors.New("Doing rollback anyway")
	return err
}
