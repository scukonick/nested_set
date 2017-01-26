#!/bin/bash

"./wait-for-it.sh" "db:5432" "--" "echo" "postgres ready!"
sleep 5
goose -env integration up 
echo "Migrated"
go test -tags integration
goose -env integration down

