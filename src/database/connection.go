package database

import (
	"fmt"
	"log"
	"os"

	"github.com/epsierra/phinex-blog-api/src/models"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func NewDatabaseConnection() (*gorm.DB, error) {

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// Get database connection details from environment variables
	dbHost := os.Getenv("POSTGRES_HOST")
	dbPort := os.Getenv("POSTGRES_PORT")
	dbUser := os.Getenv("POSTGRES_USER")
	dbPassword := os.Getenv("POSTGRES_PASSWORD")
	dbName := os.Getenv("POSTGRES_DB")

	// Create a connection string
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", dbHost, dbPort, dbUser, dbPassword, dbName)

	// Connect to PostgreSQL database using GORM
	db, err := gorm.Open(postgres.Open(dsn))
	return db, err

}

func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(&models.User{}, &models.Blog{}, &models.Follow{}, &models.Comment{}, &models.Like{}, &models.Share{}, &models.Role{}, &models.UserRole{}, &models.UsersStats{})
}
