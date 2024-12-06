package main

import (
	"fmt"
	"log"
	"os"

	"github.com/epsierra/phinex-blog-api/src/app"
	"github.com/epsierra/phinex-blog-api/src/database"
	"github.com/joho/godotenv"
)

func main() {

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	PORT := os.Getenv("PORT")

	// Initialize database conntention
	db, err := database.NewDatabaseConnection()
	if err != nil {
		log.Fatal("Error connection to database")
	}

	// Migrate Databases

	err = database.AutoMigrate(db)
	if err != nil {
		log.Fatal("Error migrating models")
	}

	// Create a new Fiber app
	app := app.AppSetup(db)

	// Start the server
	err = app.Listen(fmt.Sprintf(":%v", PORT))
	if err != nil {
		log.Fatal("Error starting server:", err)
	}
}
