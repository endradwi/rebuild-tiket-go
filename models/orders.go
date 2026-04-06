package models

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"tiket/lib"
	"time"

	"github.com/jackc/pgx/v5"
	xendit "github.com/xendit/xendit-go/v6"
	"github.com/xendit/xendit-go/v6/invoice"
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

	// 1. Lock the showtime row to serialize bookings for this showtime
	var dummy int
	err = tx.QueryRow(context.Background(), "SELECT id FROM movie_showtimes WHERE id = $1 FOR UPDATE", req.ShowtimeId).Scan(&dummy)
	if err != nil {
		return lib.Order{}, fmt.Errorf("locking showtime: %w", err)
	}

	// 2. Check seat availability
	var occupiedCount int
	err = tx.QueryRow(context.Background(), `
		SELECT COUNT(*) 
		FROM order_seats os
		JOIN orders o ON os.order_id = o.id
		WHERE o.showtime_id = $1 AND os.seat_id = ANY($2) AND o.status != 'cancelled'
	`, req.ShowtimeId, req.SeatIds).Scan(&occupiedCount)
	if err != nil {
		return lib.Order{}, fmt.Errorf("checking seat availability: %w", err)
	}
	if occupiedCount > 0 {
		return lib.Order{}, fmt.Errorf("one or more seats are already booked")
	}

	// 3. Get total price from seats
	var totalPrice int
	err = tx.QueryRow(context.Background(), `
		SELECT SUM(price) FROM seat WHERE id = ANY($1)
	`, req.SeatIds).Scan(&totalPrice)
	if err != nil {
		return lib.Order{}, fmt.Errorf("calculating total price: %w", err)
	}

	orderNumber := GenerateOrderNumber()

	// 4. Insert into orders table
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

	// 5. Insert into order_seats table
	for _, seatId := range req.SeatIds {
		_, err = tx.Exec(context.Background(), `
			INSERT INTO order_seats (order_id, seat_id)
			VALUES ($1, $2)
		`, order.Id, seatId)
		if err != nil {
			return lib.Order{}, fmt.Errorf("inserting order seat: %w", err)
		}
	}

	// 6. Commit transaction
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
		       m.title as movie_title, c.cinema_name, l.name as location_name, s.show_date, s.show_time::TEXT
		FROM orders o
		JOIN movie_showtimes s ON o.showtime_id = s.id
		JOIN movie m ON s.movie_id = m.id
		JOIN cinema c ON s.cinema_id = c.id
		JOIN location l ON c.location_id = l.id
		WHERE o.id = $1
	`, orderId).Scan(
		&order.Id, &order.OrderNumber, &order.ProfileId, &order.ShowtimeId, &order.FullName, &order.Email, &order.PhoneNumber, &order.TotalPrice, &order.Status, &order.CreatedAt,
		&order.MovieTitle, &order.CinemaName, &order.LocationName, &order.ShowDate, &order.ShowTime,
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

	// 1. Update order info and status (Change status to 'pending' waiting for payment)
	var totalPrice int
	var orderNumber string
	err = tx.QueryRow(context.Background(), `
		UPDATE orders 
		SET full_name = $1, email = $2, phone_number = $3, status = 'pending'
		WHERE id = $4
		RETURNING total_price, order_number
	`, req.FullName, req.Email, req.PhoneNumber, orderId).Scan(&totalPrice, &orderNumber)
	if err != nil {
		return lib.Payment{}, fmt.Errorf("updating order: %w", err)
	}

	// 2. Xendit Invoice Creation
	xenditKey := os.Getenv("XENDIT_SECRET_KEY")
	if xenditKey == "" {
		return lib.Payment{}, fmt.Errorf("XENDIT_SECRET_KEY is not set")
	}

	client := xendit.NewClient(xenditKey)
	
	description := "Payment for Ticket " + orderNumber
	// Create Invoice Request
	invoiceData := invoice.CreateInvoiceRequest{
		ExternalId: orderNumber,
		Amount:     float64(totalPrice),
		PayerEmail: &req.Email,
		Description: &description,
	}

	resp, _, xenditErr := client.InvoiceApi.CreateInvoice(context.Background()).
		CreateInvoiceRequest(invoiceData).
		Execute()

	if xenditErr != nil {
		return lib.Payment{}, fmt.Errorf("creating xendit invoice: %v", xenditErr.Error())
	}

	// 3. Insert into payment table
	var payment lib.Payment
	expiredAt := time.Now().Add(24 * time.Hour) // Payment expires in 24h
	err = tx.QueryRow(context.Background(), `
		INSERT INTO payment (order_id, total_payment, payment_method, payment_status, expired_at, invoice_url, external_id)
		VALUES ($1, $2, $3, 'pending', $4, $5, $6)
		RETURNING id, order_id, total_payment, payment_method, payment_status, expired_at, invoice_url, external_id
	`, orderId, totalPrice, req.PaymentMethod, expiredAt, resp.InvoiceUrl, resp.ExternalId).Scan(
		&payment.Id, &payment.OrderId, &payment.TotalPayment, &payment.PaymentMethod, &payment.PaymentStatus, &payment.ExpiredAt, &payment.InvoiceUrl, &payment.ExternalId,
	)
	if err != nil {
		return lib.Payment{}, fmt.Errorf("inserting payment: %w", err)
	}

	// 4. Commit
	if err := tx.Commit(context.Background()); err != nil {
		return lib.Payment{}, fmt.Errorf("committing transaction: %w", err)
	}

	return payment, nil
}

// GetUserOrders retrieves all orders for a specific user profile
func GetUserOrders(profileId int) ([]lib.Order, error) {
	pgConn := lib.InitDB()
	defer pgConn.Close(context.Background())

	rows, err := pgConn.Query(context.Background(), `
		SELECT o.id, o.order_number, o.showtime_id, o.total_price, o.status, o.created_at,
		       m.title as movie_title, c.cinema_name, c.image as cinema_image, l.name as location_name
		FROM orders o
		JOIN movie_showtimes s ON o.showtime_id = s.id
		JOIN movie m ON s.movie_id = m.id
		JOIN cinema c ON s.cinema_id = c.id
		JOIN location l ON c.location_id = l.id
		WHERE o.profile_id = $1
		ORDER BY o.created_at DESC
	`, profileId)
	if err != nil {
		return nil, fmt.Errorf("querying user orders: %w", err)
	}
	defer rows.Close()

	var orders []lib.Order
	for rows.Next() {
		var o lib.Order
		// Note: we use lib.Order struct but only some fields are populated for the list view
		err := rows.Scan(
			&o.Id, &o.OrderNumber, &o.ShowtimeId, &o.TotalPrice, &o.Status, &o.CreatedAt,
			&o.MovieTitle, &o.CinemaName, &o.CinemaImage, &o.LocationName,
		)
		if err != nil {
			return nil, fmt.Errorf("scanning user order: %w", err)
		}
		orders = append(orders, o)
	}

	if orders == nil {
		orders = []lib.Order{}
	}

	return orders, nil
}

// UpdatePaymentStatusByExternalId updates order and payment status based on Xendit callback
func UpdatePaymentStatusByExternalId(externalId string, status string) error {
	pgConn := lib.InitDB()
	defer pgConn.Close(context.Background())

	tx, err := pgConn.Begin(context.Background())
	if err != nil {
		return err
	}
	defer tx.Rollback(context.Background())

	// Mapping Xendit status to our local status
	// Xendit statuses: PAID, EXPIRED, SETTLED
	var orderStatus string
	var paymentStatus string

	switch status {
	case "PAID", "SETTLED":
		orderStatus = "paid"
		paymentStatus = "success"
	case "EXPIRED":
		orderStatus = "cancelled"
		paymentStatus = "failed"
	default:
		return nil // Ignore unknown statuses
	}

	// 1. Update Payment Status
	_, err = tx.Exec(context.Background(), `
		UPDATE payment SET payment_status = $1, updated_at = NOW()
		WHERE external_id = $2
	`, paymentStatus, externalId)
	if err != nil {
		return fmt.Errorf("updating payment status: %w", err)
	}

	// 2. Update Order Status
	_, err = tx.Exec(context.Background(), `
		UPDATE orders SET status = $1, updated_at = NOW()
		WHERE order_number = $2
	`, orderStatus, externalId)
	if err != nil {
		return fmt.Errorf("updating order status: %w", err)
	}

	return tx.Commit(context.Background())
}
