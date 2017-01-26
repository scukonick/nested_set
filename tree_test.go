// +build integration

package main

import (
	"database/sql"
	"testing"
)

const pgUrl = "user=postgres dbname=postgres sslmode=disable host=db port=5432"

// checkTreeValidity runs some queries on tree table
// to check if tree is valid and nothing is wrong with it.
// It returns true and nil error if everything is ok,
// false in case if there is any problem with tree,
// non nil error in case if something is wrong with db requests
// This functions does not guarantee 100% validity but helps anyway
func checkTreeValidity(tree *Tree) (bool, error) {
	var i int32
	query := "SELECT COUNT(id) FROM tree WHERE left_key >= right_key"
	err := tree.DB.QueryRow(query).Scan(&i)
	if err != nil {
		return false, err
	}
	if i != 0 {
		return false, nil
	}

	query = `SELECT id FROM tree
  GROUP BY id HAVING (right_key - left_key) % 2 = 0;`
	rows, err := tree.DB.Query(query)
	if err != nil {
		return false, err
	}
	if rows.Next() {
		return false, nil
	}
	rows.Close()

	return true, nil
}

func TestTreeRenameNode(t *testing.T) {
	db, err := sql.Open("postgres", pgUrl)
	defer db.Close()

	if err != nil {
		t.Fatalf("Expected error nil, got: %v", err)
	}

	tree := NewTree(db)
	cats, err := tree.GetNodeByValue("cats")
	if err != nil {
		t.Fatalf("Expected error nil, got: %v", err)
	}

	err = tree.RenameNode(cats, "tigers")
	if err != nil {
		t.Fatalf("Expected error nil, got: %v", err)
	}

	tigers, err := tree.GetNodeByValue("tigers")
	if err != nil {
		t.Fatalf("Expected error nil, got: %v", err)
	}

	if tigers.ID != cats.ID {
		t.Error("Failed to rename node")
	}

	valid, err := checkTreeValidity(tree)
	if err != nil {
		t.Fatalf("Expected error nil, got: %v", err)
	}
	if !valid {
		t.Error("Expected valid true, got false")
	}
}

func TestGetParent(t *testing.T) {
	db, err := sql.Open("postgres", pgUrl)
	defer db.Close()
	if err != nil {
		t.Fatalf("Expected error nil, got: %v", err)
	}
	tree := NewTree(db)

	insects, err := tree.GetNodeByValue("insects")
	if err != nil {
		t.Fatalf("Expected error nil, got: %v", err)
	}

	flies, err := tree.GetNodeByValue("flies")
	if err != nil {
		t.Fatalf("Expected error nil, got: %v", err)
	}

	fliesParent, err := tree.GetParent(flies)
	if err != nil {
		t.Fatalf("Expected error nil, got: %v", err)
	}

	if fliesParent.ID != insects.ID {
		t.Fatalf("Expected flies parent id: %d, got: %d", insects.ID, fliesParent.ID)
	}
}

func TestAddNode(t *testing.T) {
	db, err := sql.Open("postgres", pgUrl)
	defer db.Close()
	if err != nil {
		t.Fatalf("Expected error nil, got: %v", err)
	}
	tree := NewTree(db)

	mammals, err := tree.GetNodeByValue("mammals")
	if err != nil {
		t.Fatalf("Expected error nil, got: %v", err)
	}

	horses, err := tree.InsertChild(mammals, "horses")
	if err != nil {
		t.Fatalf("Expected error nil, got: %v", err)
	}

	// checking horses parent
	horsesParent, err := tree.GetParent(horses)
	if err != nil {
		t.Fatalf("Expected error nil, got: %v", err)
	}
	if horsesParent.ID != mammals.ID {
		t.Fatalf("Expected new parent id: %d, got: %d", horses.ID, horsesParent.ID)
	}

	// Checking tree validity
	valid, err := checkTreeValidity(tree)
	if err != nil {
		t.Fatalf("Expected error nil, got: %v", err)
	}
	if !valid {
		t.Error("Expected valid true, got false")
	}
}

func TestDeleteNode(t *testing.T) {
	db, err := sql.Open("postgres", pgUrl)
	defer db.Close()
	if err != nil {
		t.Fatalf("Expected error nil, got: %v", err)
	}
	tree := NewTree(db)

	cats, err := tree.GetNodeByValue("cats")
	if err != nil {
		t.Fatalf("Expected error nil, got: %v", err)
	}

	err = tree.DeleteNode(cats)
	if err != nil {
		t.Fatalf("Expected error nil, got: %v", err)
	}

	_, err = tree.GetNodeByValue("cats")
	if err != ErrNodeDoesNotExist {
		t.Fatalf("Expected error ErrNodeDoesNotExist, got: %v", err)
	}
}
