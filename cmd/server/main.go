package main

import (
	"fmt"

	database "github.com/anas-salha/2do/internal/database"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	fmt.Println("2do - An overengineered TODO app")

	db, err := database.GetDatabase()
	if err != nil {
		fmt.Println(err)
		return
	}

	// Verify the connection
	err = db.Ping()
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("Successfully connected to the database!")

	err = database.RunMigrations(db)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("Database migrations applied successfully.")
}
