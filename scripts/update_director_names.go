package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found")
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL is not set")
	}

	ctx := context.Background()
	conn, err := pgx.Connect(ctx, dbURL)
	if err != nil {
		log.Fatal("Unable to connect to database:", err)
	}
	defer conn.Close(ctx)

	// 1. Update director_names
	updates := map[string]string{
		"Spider-Man: Homecoming": "Jon Watts",
		"Inception":              "Christopher Nolan",
		"The Dark Knight":        "Christopher Nolan",
		"Avengers: Endgame":      "Anthony & Joe Russo",
	}

	for title, director := range updates {
		_, err := conn.Exec(ctx, "UPDATE movie SET director_name = $1 WHERE title = $2", director, title)
		if err != nil {
			log.Printf("Failed to update movie %s: %v", title, err)
		} else {
			fmt.Printf("Updated director for %s: %s\n", title, director)
		}
	}

	// 2. Populate junction tables if empty
	// We'll just run these as idempotent inserts where possible or check existence
	
	fmt.Println("Database updates completed!")
}
