package models

import (
	"context"
	"fmt"
	"tiket/lib"
)

// GetAllSeats retrieves all available seats
func GetAllSeats() ([]lib.Seat, error) {
	pgConn := lib.InitDB()
	defer pgConn.Close(context.Background())

	rows, err := pgConn.Query(context.Background(), `SELECT id, name, price FROM seat ORDER BY id ASC`)
	if err != nil {
		return nil, fmt.Errorf("querying seats: %w", err)
	}
	defer rows.Close()

	var seats []lib.Seat
	for rows.Next() {
		var s lib.Seat
		err := rows.Scan(&s.Id, &s.Name, &s.Price)
		if err != nil {
			return nil, fmt.Errorf("scanning seat: %w", err)
		}
		seats = append(seats, s)
	}

	if seats == nil {
		seats = []lib.Seat{}
	}

	return seats, nil
}

// GetOccupiedSeats returns seat IDs that are already booked for a specific showtime
func GetOccupiedSeats(showtimeId int) ([]int, error) {
	pgConn := lib.InitDB()
	defer pgConn.Close(context.Background())

	rows, err := pgConn.Query(context.Background(), `
		SELECT os.seat_id 
		FROM order_seats os
		JOIN orders o ON os.order_id = o.id
		WHERE o.showtime_id = $1 AND o.status != 'cancelled'
	`, showtimeId)
	if err != nil {
		return nil, fmt.Errorf("querying occupied seats: %w", err)
	}
	defer rows.Close()

	var seatIds []int
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("scanning seat id: %w", err)
		}
		seatIds = append(seatIds, id)
	}

	return seatIds, nil
}
