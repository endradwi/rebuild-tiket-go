package models

import (
	"context"
	"fmt"
	"tiket/lib"
)

// CreateShowtime inserts a new showtime entry
func CreateShowtime(req lib.MovieShowtime) (lib.MovieShowtime, error) {
	pgConn := lib.InitDB()
	defer pgConn.Close(context.Background())

	var showtime lib.MovieShowtime
	err := pgConn.QueryRow(context.Background(), `
		INSERT INTO movie_showtimes (movie_id, cinema_id, show_date, show_time, price)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, movie_id, cinema_id, show_date, show_time, price
	`, req.MovieId, req.CinemaId, req.ShowDate, req.ShowTime, req.Price).Scan(
		&showtime.Id, &showtime.MovieId, &showtime.CinemaId, &showtime.ShowDate, &showtime.ShowTime, &showtime.Price,
	)

	if err != nil {
		return showtime, fmt.Errorf("creating showtime: %w", err)
	}

	return showtime, nil
}

// GetShowtimesByMovie retrieves all showtimes for a specific movie
// Optionally filtered by location
func GetShowtimesByMovie(movieId int, locationId int) ([]lib.MovieShowtime, error) {
	pgConn := lib.InitDB()
	defer pgConn.Close(context.Background())

	query := `
		SELECT s.id, s.movie_id, s.cinema_id, s.show_date, s.show_time, s.price
		FROM movie_showtimes s
		JOIN cinema c ON s.cinema_id = c.id
		WHERE s.movie_id = $1
	`
	args := []any{movieId}

	if locationId > 0 {
		query += " AND c.location_id = $2"
		args = append(args, locationId)
	}

	rows, err := pgConn.Query(context.Background(), query, args...)
	if err != nil {
		return nil, fmt.Errorf("querying showtimes: %w", err)
	}
	defer rows.Close()

	var showtimes []lib.MovieShowtime
	for rows.Next() {
		var s lib.MovieShowtime
		err := rows.Scan(&s.Id, &s.MovieId, &s.CinemaId, &s.ShowDate, &s.ShowTime, &s.Price)
		if err != nil {
			return nil, fmt.Errorf("scanning showtime: %w", err)
		}
		showtimes = append(showtimes, s)
	}

	if showtimes == nil {
		showtimes = []lib.MovieShowtime{}
	}

	return showtimes, nil
}
