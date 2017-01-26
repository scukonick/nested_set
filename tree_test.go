// +build integration

package main

import (
	"database/sql"
	"testing"
)

const pgUrl = "user=postgres dbname=postgres sslmode=disable host=db port=5432"

func TestTreeRenameNode(t *testing.T) {
	db, err := sql.Open("postgres", pgUrl)
	if err != nil {
		t.Fatalf("Failed to connect to db: %v", err)
	}
	tree := NewTree(db)
	t.Logf("Tree: %v", err)
}
