package models

import (
	"context"
	"fmt"
	"tiket/lib"
)

// GetDashboardStats retrieves sales statistics for the admin dashboard
func GetDashboardStats() (lib.DashboardStats, error) {
	pgConn := lib.InitDB()
	defer pgConn.Close(context.Background())

	var stats lib.DashboardStats

	// 1. Sales Chart (Last 6 months)
	rows, err := pgConn.Query(context.Background(), `
		SELECT TO_CHAR(created_at, 'Mon') as label, SUM(total_price) as value
		FROM orders
		WHERE status = 'paid' AND created_at >= NOW() - INTERVAL '6 months'
		GROUP BY TO_CHAR(created_at, 'Mon'), DATE_TRUNC('month', created_at)
		ORDER BY DATE_TRUNC('month', created_at)
	`)
	if err != nil {
		return stats, fmt.Errorf("fetching sales chart: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var s lib.SalesStat
		if err := rows.Scan(&s.Label, &s.Value); err != nil {
			return stats, err
		}
		stats.SalesChart = append(stats.SalesChart, s)
	}

	// 2. Ticket Sales by Location (for the bottom chart)
	rows, err = pgConn.Query(context.Background(), `
		SELECT l.name as label, SUM(o.total_price) as value
		FROM orders o
		JOIN movie_showtimes s ON o.showtime_id = s.id
		JOIN cinema c ON s.cinema_id = c.id
		JOIN location l ON c.location_id = l.id
		WHERE o.status = 'paid'
		GROUP BY l.name
		ORDER BY value DESC
		LIMIT 5
	`)
	if err != nil {
		return stats, fmt.Errorf("fetching ticket sales by location: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var s lib.SalesStat
		if err := rows.Scan(&s.Label, &s.Value); err != nil {
			return stats, err
		}
		stats.TicketSales = append(stats.TicketSales, s)
	}

	// 3. Average Earnings
	err = pgConn.QueryRow(context.Background(), `
		SELECT COALESCE(AVG(total_price), 0)::INT
		FROM orders
		WHERE status = 'paid'
	`).Scan(&stats.AverageEarnings)
	if err != nil {
		return stats, fmt.Errorf("fetching average earnings: %w", err)
	}

	if stats.SalesChart == nil {
		stats.SalesChart = []lib.SalesStat{}
	}
	if stats.TicketSales == nil {
		stats.TicketSales = []lib.SalesStat{}
	}

	return stats, nil
}
