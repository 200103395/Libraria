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
	fmt.Println("Initializing DB connection")
	store, err := database.NewPostgresStorage()
	if err != nil {
		log.Fatal(err)
	}

	err = store.Init()
	if err != nil {
		log.Fatal(err)
	}

	//fmt.Println(store.DropTable("book"))

	fmt.Println("Initializing Mail connection")
	email := mail.NewEmailConnection()
	if err != nil {
		log.Fatal(err)
	}

	lib := controllers.NewLibServer(":8000", store, *email)

	//go store.ClearRequests()
	//s, _ := bcrypt.GenerateFromPassword([]byte("Password"), bcrypt.DefaultCost)
	//fmt.Println(string(s))
	//$2a$10$iXXY02lkYivkjSs9Hr/MRe/wOw.E0WC0EdLSP9qjQaK6Az9GbCgSm
	fmt.Println("Starting application")
	lib.Run()

	fmt.Println("Application is running")

}
