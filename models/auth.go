package models

import (
	"context"
	"errors"
	"fmt"
	"tiket/lib"

	"github.com/jackc/pgx/v5"
)

func Register(user lib.UserRole) error {
	pgConn := lib.InitDB()

	defer pgConn.Close(context.Background())
	

	//Create hash password
		hash, err := lib.HashPassword(user.User.Password)

		if err != nil {
			return  fmt.Errorf("hashing password: %w", err)
		}

	// Ambil role id
	var roleId int
	if user.RoleId != 0 {
		roleId = user.RoleId
	} else {
		
		// Cari dari DB
		err = pgConn.QueryRow(context.Background(), `SELECT id FROM "role" WHERE name = $1`, "USER").Scan(&roleId)
		if err != nil {
			return  fmt.Errorf("getting role id: %w", err)
		}
	}

	// Cari email jika sudah di daftarkan karena email UNIQUE
	var existingEmail string
	err = pgConn.QueryRow(context.Background(), `SELECT email FROM "user" WHERE email = $1`, user.User.Email).Scan(&existingEmail)
	if err == nil {
		return  fmt.Errorf("email already exists")
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return  fmt.Errorf("checking email existence: %w", err)
	}

	// Insert ke dalam table
	var userId int

	err = pgConn.QueryRow(context.Background(), `INSERT INTO "user" (email, password) VALUES ($1, $2) RETURNING id`, user.User.Email, hash).Scan(&userId)

	if err != nil {
		return  fmt.Errorf("inserting user: %w", err)
	}

	// Exec di gunakan untuk melakukan insert tanpa returning
	_,err = pgConn.Exec(context.Background(), `INSERT INTO profile (user_id, role_id) VALUES ($1, $2)`, userId, roleId)
	if err != nil {
		return  fmt.Errorf("inserting profile: %w", err)
	}

	return  nil
}

// func Login(user lib.User) lib.User {
// 	pgConn := lib.InitDB()

// 	defer pgConn.Close(context.Background())

// 	//Create hash password
	
	
// }
