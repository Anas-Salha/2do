package main

import (
	"fmt"
	"log"

	"github.com/anas-salha/2do/internal/database"
	"github.com/anas-salha/2do/internal/router"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	fmt.Println("2do - An overengineered TODO app")

	db, err := database.GetDatabase()
	if err != nil {
		log.Fatal(err)
	}

	// Verify the connection
	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Successfully connected to the database!")

	err = database.RunMigrations(db)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Database migrations applied successfully.")

	r := router.NewRouter(db)
	r.Run("0.0.0.0:8080")
}
