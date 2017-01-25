package main

import (
	"database/sql"
	"log"

	_ "github.com/lib/pq"
)

// http://www.getinfo.ru/article610.html

func main() {
	db, err := sql.Open("postgres", "postgres://nest_tree:qwerty@localhost/nest_tree?sslmode=disable")
	if err != nil {
		log.Fatalf("Failed to connect to db: %v", err)
	}

	t := NewTree(db)
	pop, err := t.IsPopulated()
	if err != nil {
		log.Fatalf("Error during checking tree: %v", err)
	}
	log.Printf("P: %v", pop)

	nodes, err := t.GetAllNodes()
	if err != nil {
		log.Fatalf("Failed to get tree nodes: %v", err)
	}
	for _, node := range nodes {
		log.Printf("Node: %+v", node)
	}

	value := "bees"
	movingNode, err := t.GetNodeByValue(value)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("To move: %+v", movingNode)

	newParent, err := t.GetNodeByValue("dogs")
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("New Parent: %+v", newParent)
	err = t.MyMove(newParent, movingNode)
	if err != nil {
		log.Fatal(err)
	}
	t.showTree()

	/*log.Printf("Inserting horses")
	_, err = t.InsertChild(mammals, "horses")
	if err != nil {
		log.Fatal(err)
	}
	t.showTree()

	mammals, err = t.GetNodeByValue(value)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Inserting sheeps")
	cats, err := t.InsertChild(mammals, "sheeps")
	if err != nil {
		log.Fatal(err)
	}
	t.showTree()

	log.Printf("Deleting sheeps")
	err = t.DeleteNode(cats)
	if err != nil {
		log.Fatal(err)
	}
	t.showTree()*/
}
