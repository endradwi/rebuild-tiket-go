package models

import (
	"context"
	"errors"
	"fmt"
	"tiket/lib"
	"time"

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

	_, err = FindEmail(user.User)
	
	// If FindEmail returns no error, it means the email WAS found, which is an error for registration.
	if err == nil {
		return fmt.Errorf("email already exists")
	}

	// If FindEmail returns an error, we need to check if it's the "no rows" error.
	// FindEmail wraps the error: fmt.Errorf("checking email existence: %w", err)
	// errors.Is handles wrapped errors automatically.
	if !errors.Is(err, pgx.ErrNoRows) {
		// If it's some OTHER error (db connection, etc.), return that error.
		return fmt.Errorf("checking email existence: %w", err)
	}

	// If we are here, err is pgx.ErrNoRows, which means the email is NOT in the database.
	// This is exactly what we want for a new registration.

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

func FindEmail(user lib.User) (lib.User, error) {
	pgConn := lib.InitDB()

	defer pgConn.Close(context.Background())

	// cek email ada atau tidak
	var dbUser lib.User
	err := pgConn.QueryRow(context.Background(), `SELECT id, email, password FROM "user" WHERE email = $1`, user.Email).Scan(&dbUser.Id, &dbUser.Email, &dbUser.Password)
	if err != nil {
		return user, fmt.Errorf("checking email existence: %w", err)
	}

	return dbUser, nil
}

func CreateResetPassword(resetPassword lib.ResetPassword) error {
	pgConn := lib.InitDB()

	defer pgConn.Close(context.Background())

	var profileId int
	err := pgConn.QueryRow(context.Background(), `SELECT id FROM profile WHERE user_id = $1`, resetPassword.ProfileId).Scan(&profileId)
	if err != nil {
		return fmt.Errorf("checking profile existence: %w", err)
	}

	data, err := pgConn.Exec(context.Background(), `INSERT INTO reset_password (profile_id, token_hash, expired_at) VALUES ($1, $2, $3)`, profileId, resetPassword.TokenHash, resetPassword.ExpiredAt)
	if err != nil {
		return fmt.Errorf("inserting reset password: %w", err)
	}

	fmt.Println(data)

	return nil
}

func FindResetPassword(tokenHash string) (lib.ResetPassword, error) {
	pgConn := lib.InitDB()

	defer pgConn.Close(context.Background())

	var resetPassword lib.ResetPassword
	err := pgConn.QueryRow(context.Background(), `SELECT id, profile_id, token_hash, expired_at, used_at, created_at FROM reset_password WHERE token_hash = $1`, tokenHash).Scan(&resetPassword.Id, &resetPassword.ProfileId, &resetPassword.TokenHash, &resetPassword.ExpiredAt, &resetPassword.UsedAt, &resetPassword.CreatedAt)

	if err != nil {
		return resetPassword, fmt.Errorf("checking reset password: %w", err)
	}

	return resetPassword, nil
}

func UpdatePassword(profileId int, password string) error {
	pgConn := lib.InitDB()

	defer pgConn.Close(context.Background())

	_, err := pgConn.Exec(context.Background(), `UPDATE "user" SET password = $1 WHERE id = (SELECT user_id FROM profile WHERE id = $2)`, password, profileId)
	if err != nil {
		return fmt.Errorf("updating password: %w", err)
	}

	_, err = pgConn.Exec(context.Background(), `UPDATE reset_password SET used_at = $1 WHERE profile_id = $2`, time.Now(), profileId)
	if err != nil {
		return fmt.Errorf("updating reset password: %w", err)
	}

	return nil
}

func DeleteResetPassword(tokenHash string) error {
	pgConn := lib.InitDB()

	defer pgConn.Close(context.Background())

	_, err := pgConn.Exec(context.Background(), `DELETE FROM reset_password WHERE token_hash = $1`, tokenHash)
	if err != nil {
		return fmt.Errorf("deleting reset password: %w", err)
	}

	return nil
}
