package main

import (
	"Libraria/controllers"
	"Libraria/database"
	"Libraria/mail"
	"fmt"
	_ "github.com/lib/pq"
	"log"
)

func main() {
	fmt.Println("Hello, World!")
	store, err := database.NewPostgresStorage()
	if err != nil {
		log.Fatal(err)
	}

	err = store.Init()
	if err != nil {
		log.Fatal(err)
	}

	//fmt.Println(store.DropTable("book"))

	email := mail.NewEmailConnection()
	if err != nil {
		log.Fatal(err)
	}

	lib := controllers.NewLibServer(":8000", store, *email)

	//go store.ClearRequests()

	lib.Run()

}
