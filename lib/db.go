package lib

import (
	"context"
	"os"

	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
)


func InitDB() *pgx.Conn {
	if err := godotenv.Load(); err != nil {
		godotenv.Load("../.env")
	}


	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		panic("DATABASE_URL is not set")
	}

	conn, err := pgx.Connect(context.Background(), connStr)
	if err != nil {
		panic(err)
	}

	return conn
}
