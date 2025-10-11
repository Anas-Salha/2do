package main

import (
	"fmt"
	"log"

	"github.com/anas-salha/2do/internal/buildinfo"
	"github.com/anas-salha/2do/internal/config"
	"github.com/anas-salha/2do/internal/database"
	"github.com/anas-salha/2do/internal/http"
	"github.com/anas-salha/2do/internal/todo"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	fmt.Printf(
		"2do - An overengineered TODO app (version %s, commit %s, built %s)\n",
		buildinfo.Version,
		buildinfo.Commit,
		buildinfo.Date,
	)

	cfg := config.Load()
	db, err := database.Open(cfg)
	if err != nil {
		log.Fatal(err)
	}

	err = database.RunMigrations(db)
	if err != nil {
		log.Fatalf("Running migrations failed: %v", err)
	}

	todoRepo := todo.NewRepo(db)
	todoService := todo.NewService(todoRepo)
	todoHandler := todo.NewHandler(todoService)

	r := http.NewRouter(todoHandler, cfg.AllowedOrigins)

	r.Run("0.0.0.0:" + cfg.Port)
}
