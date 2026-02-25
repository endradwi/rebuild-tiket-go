package models

import (
	"context"
	"fmt"
	"tiket/lib"

	"github.com/jackc/pgx/v5"
)

// GetUserProfile retrieves the user's profile and email
func GetUserProfile(userId int) (lib.UserProfile, error) {
	pgConn := lib.InitDB()
	defer pgConn.Close(context.Background())

	var profile lib.UserProfile
	err := pgConn.QueryRow(context.Background(), `
		SELECT u.id, u.email, p.first_name, p.last_name, p.phone_number, p.image, p.point
		FROM "user" u
		LEFT JOIN profile p ON u.id = p.user_id
		WHERE u.id = $1
	`, userId).Scan(
		&profile.Id, &profile.Email, &profile.FirstName, &profile.LastName, 
		&profile.PhoneNumber, &profile.Image, &profile.Point,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return profile, fmt.Errorf("user not found")
		}
		return profile, fmt.Errorf("querying user profile: %w", err)
	}

	return profile, nil
}

// UpdateUserProfile updates the user's profile fields
func UpdateUserProfile(userId int, req lib.ProfileUpdateRequest) (lib.UserProfile, error) {
	pgConn := lib.InitDB()
	defer pgConn.Close(context.Background())

	// Build dynamic query for PATCH using COALESCE
	_, err := pgConn.Exec(context.Background(), `
		UPDATE profile 
		SET 
			first_name = COALESCE($1, first_name),
			last_name = COALESCE($2, last_name),
			phone_number = COALESCE($3, phone_number),
			image = COALESCE($4, image)
		WHERE user_id = $5
	`, req.FirstName, req.LastName, req.PhoneNumber, req.Image, userId)

	if err != nil {
		return lib.UserProfile{}, fmt.Errorf("updating profile: %w", err)
	}

	// Fetch returning profile
	return GetUserProfile(userId)
}

// GetAllUsers retrieves all users
func GetAllUsers() ([]lib.UserProfile, error) {
	pgConn := lib.InitDB()
	defer pgConn.Close(context.Background())

	rows, err := pgConn.Query(context.Background(), `
		SELECT u.id, u.email, p.first_name, p.last_name, p.phone_number, p.image, p.point
		FROM "user" u
		LEFT JOIN profile p ON u.id = p.user_id
		ORDER BY u.id ASC
	`)
	if err != nil {
		return nil, fmt.Errorf("querying users: %w", err)
	}
	defer rows.Close()

	var users []lib.UserProfile
	for rows.Next() {
		var profile lib.UserProfile
		err := rows.Scan(
			&profile.Id, &profile.Email, &profile.FirstName, &profile.LastName, 
			&profile.PhoneNumber, &profile.Image, &profile.Point,
		)
		if err != nil {
			return nil, fmt.Errorf("scanning user: %w", err)
		}
		users = append(users, profile)
	}

	if users == nil {
		users = []lib.UserProfile{}
	}

	return users, nil
}

// DeleteUser deletes a user by ID
func DeleteUser(userId int) error {
	pgConn := lib.InitDB()
	defer pgConn.Close(context.Background())
	// Because of ON DELETE CASCADE, deleting user deletes profile too
	_, err := pgConn.Exec(context.Background(), `DELETE FROM "user" WHERE id = $1`, userId)
	if err != nil {
		return fmt.Errorf("deleting user: %w", err)
	}
	return nil
}
