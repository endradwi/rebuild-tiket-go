package models

import (
	"context"
	"fmt"
	"tiket/lib"
	"time"
)

func Register(req lib.RegisterRequest) error {
	pgConn := lib.InitDB()
	defer pgConn.Close(context.Background())

	tx, err := pgConn.Begin(context.Background())
	if err != nil {
		return fmt.Errorf("starting transaction: %w", err)
	}
	defer tx.Rollback(context.Background())

	// Create hash password
	hash, err := lib.HashPassword(req.Password)
	if err != nil {
		return fmt.Errorf("hashing password: %w", err)
	}

	// Default to USER role
	var roleId int
	err = tx.QueryRow(context.Background(), `SELECT id FROM role WHERE name = $1`, "USER").Scan(&roleId)
	if err != nil {
		return fmt.Errorf("getting role id: %w", err)
	}

	// Check if email already exists
	var exists bool
	err = tx.QueryRow(context.Background(), `SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)`, req.Email).Scan(&exists)
	if err != nil {
		return fmt.Errorf("checking email existence: %w", err)
	}
	if exists {
		return fmt.Errorf("email already exists")
	}

	// Insert into users table
	var userId int
	err = tx.QueryRow(context.Background(), `INSERT INTO users (email, password) VALUES ($1, $2) RETURNING id`, req.Email, hash).Scan(&userId)
	if err != nil {
		return fmt.Errorf("inserting user: %w", err)
	}

	// Insert into profile table
	_, err = tx.Exec(context.Background(), `INSERT INTO profile (user_id, role_id) VALUES ($1, $2)`, userId, roleId)
	if err != nil {
		return fmt.Errorf("inserting profile: %w", err)
	}

	if err := tx.Commit(context.Background()); err != nil {
		return fmt.Errorf("committing transaction: %w", err)
	}

	return nil
}

func FindEmail(email string) (lib.User, error) {
	pgConn := lib.InitDB()

	defer pgConn.Close(context.Background())

	// cek email ada atau tidak
	var dbUser lib.User
	err := pgConn.QueryRow(context.Background(), `
		SELECT u.id, u.email, u.password, p.role_id, r.name as role_name 
		FROM users u
		JOIN profile p ON u.id = p.user_id
		JOIN role r ON p.role_id = r.id
		WHERE u.email = $1
	`, email).Scan(&dbUser.Id, &dbUser.Email, &dbUser.Password, &dbUser.RoleId, &dbUser.RoleName)
	if err != nil {
		return lib.User{}, fmt.Errorf("checking email existence: %w", err)
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

	tx, err := pgConn.Begin(context.Background())
	if err != nil {
		return fmt.Errorf("starting transaction: %w", err)
	}
	defer tx.Rollback(context.Background())

	_, err = tx.Exec(context.Background(), `UPDATE users SET password = $1 WHERE id = (SELECT user_id FROM profile WHERE id = $2)`, password, profileId)
	if err != nil {
		return fmt.Errorf("updating password: %w", err)
	}

	_, err = tx.Exec(context.Background(), `UPDATE reset_password SET used_at = $1 WHERE profile_id = $2`, time.Now(), profileId)
	if err != nil {
		return fmt.Errorf("updating reset password: %w", err)
	}

	if err := tx.Commit(context.Background()); err != nil {
		return fmt.Errorf("committing transaction: %w", err)
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
