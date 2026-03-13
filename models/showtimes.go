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
	
	if err == nil {
		// Fetch names for the display after successful creation
		err = pgConn.QueryRow(context.Background(), `
			SELECT m.title, c.cinema_name, l.name as location_name
			FROM movie_showtimes s
			JOIN movie m ON s.movie_id = m.id
			JOIN cinema c ON s.cinema_id = c.id
			JOIN location l ON c.location_id = l.id
			WHERE s.id = $1
		`, showtime.Id).Scan(&showtime.MovieTitle, &showtime.CinemaName, &showtime.LocationName)
	}

	if err != nil {
		return showtime, fmt.Errorf("creating showtime: %w", err)
	}

	return showtime, nil
}
