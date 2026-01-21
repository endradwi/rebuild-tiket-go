package models

import (
	"context"
	"fmt"
	"tiket/lib"
)

func Register(user lib.User) lib.User {
	pgConn := lib.InitDB()

	defer pgConn.Close(context.Background())

	//Create hash password
		hash, err := lib.HashPassword(user.Password)

		if err != nil {
			fmt.Println(err)
			return user
		}

	// Exec di gunakan untuk insert data ke database
	_, err = pgConn.Exec(context.Background(),
		"INSERT INTO users (username, password) VALUES ($1, $2)",
		user.Username, hash,
	)

	return user
}