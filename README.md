## How to
### Usage
Example usage:
```go
// main.go
package main

import (
	"database/sql"
	"log"

	_ "github.com/lib/pq"
	"github.com/scukonick/nested_set"
)

func main() {
	dbURL := "user=nest_tree dbname=nest_tree sslmode=disable host=127.0.0.1 port=5432"

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Could not open DB: %v", err)
	}

	tree := nestedset.NewTree(db)

	// check if tree is empty
	populated, err := tree.IsPopulated()
	if err != nil {
		log.Fatalf("Unexpected err: %v", err)
	}
	if populated {
		log.Fatal("Expected unpopulated tree")
	}

	// creating new root node
	err = tree.Plant("animals")
	if err != nil {
		log.Fatalf("Unexpected err: %v", err)
	}

	// populating tree
	mammals, err := tree.InsertChild(tree.Root, "mammals")
	if err != nil {
		log.Fatalf("Unexpected err: %v", err)
	}
	animals, err := tree.GetNodeByValue("animals")
	if err != nil {
		log.Fatalf("Unexpected err: %v", err)
	}
	insects, err := tree.InsertChild(animals, "insects")

	if err != nil {
		log.Fatalf("Unexpected err: %v", err)
	}
	// rename node
	err = tree.RenameNode(insects, "fish")
	if err != nil {
		log.Fatalf("Unexpected err: %v", err)
	}

	// move node
	fish, err := tree.GetNodeByValue("fish")
	err = tree.MoveNode(mammals, fish)
	if err != nil {
		log.Fatalf("Unexpected err: %v", err)
	}

	// removing node
	fish, err = tree.GetNodeByValue("fish")
	err = tree.DeleteNode(fish)
	if err != nil {
		log.Fatalf("Unexpected err: %v", err)
	}
}

```
Result in the database would be like this:
```
nest_tree=> SELECT * FROM tree ORDER BY left_key ;
 id | left_key | right_key |  value  
----+----------+-----------+---------
  1 |        1 |         4 | animals
  2 |        2 |         3 | mammals
```

### Tests
To test package please run:

```
go get github.com/scukonick/nested_set
cd $GOPATH/src/github.com/scukonick/nested_set
docker-compose build
docker-compose up --force-recreate
```
