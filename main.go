package main

import (
	"fmt"
	_ "github.com/lib/pq"
	"log"
)

func main() {
	fmt.Println("Hello, World!")
	store, err := NewPostgresStorage()
	if err != nil {
		log.Fatal(err)
	}

	err = store.Init()
	if err != nil {
		log.Fatal(err)
	}

	email := NewEmailConnection()
	if err != nil {
		log.Fatal(err)
	}

	lib := NewLibServer(":8000", store, *email)

	lib.Run()
}
