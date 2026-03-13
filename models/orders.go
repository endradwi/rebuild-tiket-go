package models

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"tiket/lib"
	"time"

	"github.com/jackc/pgx/v5"
)

// GenerateRandomString generates a random hex string of given length
func GenerateOrderNumber() string {
	b := make([]byte, 4)
	rand.Read(b)
	return "TKT-" + hex.EncodeToString(b)
}

// CreateOrder handles the transactional creation of an order and its seats
func CreateOrder(userId int, req lib.OrderCreateRequest) (lib.Order, error) {
	pgConn := lib.InitDB()
	defer pgConn.Close(context.Background())

	tx, err := pgConn.Begin(context.Background())
	if err != nil {
		return lib.Order{}, fmt.Errorf("starting transaction: %w", err)
	}
	defer tx.Rollback(context.Background())

	// 1. Get total price from seats
	var totalPrice int
	err = tx.QueryRow(context.Background(), `
		SELECT SUM(price) FROM seat WHERE id = ANY($1)
	`, req.SeatIds).Scan(&totalPrice)
	if err != nil {
		return lib.Order{}, fmt.Errorf("calculating total price: %w", err)
	}

	orderNumber := GenerateOrderNumber()

	// 2. Insert into orders table
	var order lib.Order
	err = tx.QueryRow(context.Background(), `
		INSERT INTO orders (order_number, profile_id, showtime_id, total_price, status)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, order_number, profile_id, showtime_id, total_price, status, created_at
	`, orderNumber, userId, req.ShowtimeId, totalPrice, "pending").Scan(
		&order.Id, &order.OrderNumber, &order.ProfileId, &order.ShowtimeId, &order.TotalPrice, &order.Status, &order.CreatedAt,
	)
	if err != nil {
		return lib.Order{}, fmt.Errorf("inserting order: %w", err)
	}

	// 3. Insert into order_seats table
	for _, seatId := range req.SeatIds {
		_, err = tx.Exec(context.Background(), `
			INSERT INTO order_seats (order_id, seat_id)
			VALUES ($1, $2)
		`, order.Id, seatId)
		if err != nil {
			return lib.Order{}, fmt.Errorf("inserting order seat: %w", err)
		}
	}

	// 4. Commit transaction
	if err := tx.Commit(context.Background()); err != nil {
		return lib.Order{}, fmt.Errorf("committing transaction: %w", err)
	}

	return order, nil
}

// GetOrderById retrieves a complete order with its associations
func GetOrderById(orderId int) (lib.Order, error) {
	pgConn := lib.InitDB()
	defer pgConn.Close(context.Background())

	var order lib.Order
	err := pgConn.QueryRow(context.Background(), `
		SELECT o.id, o.order_number, o.profile_id, o.showtime_id, o.full_name, o.email, o.phone_number, o.total_price, o.status, o.created_at,
		       m.title as movie_title, c.cinema_name, s.show_date, s.show_time::TEXT
		FROM orders o
		JOIN movie_showtimes s ON o.showtime_id = s.id
		JOIN movie m ON s.movie_id = m.id
		JOIN cinema c ON s.cinema_id = c.id
		WHERE o.id = $1
	`, orderId).Scan(
		&order.Id, &order.OrderNumber, &order.ProfileId, &order.ShowtimeId, &order.FullName, &order.Email, &order.PhoneNumber, &order.TotalPrice, &order.Status, &order.CreatedAt,
		&order.MovieTitle, &order.CinemaName, &order.ShowDate, &order.ShowTime,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			fmt.Printf("Order not found: ID %d\n", orderId)
			return order, fmt.Errorf("order not found")
		}
		fmt.Printf("Error querying order ID %d: %v\n", orderId, err)
		return order, fmt.Errorf("querying order: %w", err)
	}

	// Fetch seats
	rows, err := pgConn.Query(context.Background(), `
		SELECT s.id, s.name, s.price
		FROM order_seats os
		JOIN seat s ON os.seat_id = s.id
		WHERE os.order_id = $1
	`, orderId)
	if err != nil {
		return order, fmt.Errorf("querying order seats: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var seat lib.Seat
		if err := rows.Scan(&seat.Id, &seat.Name, &seat.Price); err != nil {
			return order, fmt.Errorf("scanning seat: %w", err)
		}
		order.Seats = append(order.Seats, seat)
	}

	return order, nil
}

// ProcessPayment updates order with personal info and creates a payment record
func CreatePayment(orderId int, req lib.PaymentRequest) (lib.Payment, error) {
	pgConn := lib.InitDB()
	defer pgConn.Close(context.Background())

	tx, err := pgConn.Begin(context.Background())
	if err != nil {
		return lib.Payment{}, fmt.Errorf("starting transaction: %w", err)
	}
	defer tx.Rollback(context.Background())

	// 1. Update order info and status
	var totalPrice int
	err = tx.QueryRow(context.Background(), `
		UPDATE orders 
		SET full_name = $1, email = $2, phone_number = $3, status = 'paid'
		WHERE id = $4
		RETURNING total_price
	`, req.FullName, req.Email, req.PhoneNumber, orderId).Scan(&totalPrice)
	if err != nil {
		return lib.Payment{}, fmt.Errorf("updating order: %w", err)
	}

	// 2. Insert into payment table
	var payment lib.Payment
	expiredAt := time.Now().Add(24 * time.Hour) // Payment expires in 24h
	err = tx.QueryRow(context.Background(), `
		INSERT INTO payment (order_id, total_payment, payment_method, payment_status, expired_at)
		VALUES ($1, $2, $3, 'success', $4)
		RETURNING id, order_id, total_payment, payment_method, payment_status, expired_at
	`, orderId, totalPrice, req.PaymentMethod, expiredAt).Scan(
		&payment.Id, &payment.OrderId, &payment.TotalPayment, &payment.PaymentMethod, &payment.PaymentStatus, &payment.ExpiredAt,
	)
	if err != nil {
		return lib.Payment{}, fmt.Errorf("inserting payment: %w", err)
	}

	// 3. Commit
	if err := tx.Commit(context.Background()); err != nil {
		return lib.Payment{}, fmt.Errorf("committing transaction: %w", err)
	}

	return payment, nil
}
