package models

import (
	"context"
	"encoding/json"
	"fmt"
	"tiket/lib"
	"time"

	"github.com/jackc/pgx/v5"
)

// CreateMovie inserts a new movie into the database
func CreateMovie(req lib.MovieCreateRequest) (lib.Movie, error) {
	pgConn := lib.InitDB()
	defer pgConn.Close(context.Background())

	var movie lib.Movie
	err := pgConn.QueryRow(context.Background(), `
		INSERT INTO movie (image, title, released_at, recommendation, duration, synopsis, genre_id, caster_id, cinema_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, image, title, released_at, recommendation, duration, synopsis, genre_id, caster_id, cinema_id
	`, req.Image, req.Title, req.ReleasedAt, req.Recommendation, req.Duration, req.Synopsis, req.GenreId, req.CasterId, req.CinemaId).Scan(
		&movie.Id, &movie.Image, &movie.Title, &movie.ReleasedAt, &movie.Recommendation,
		&movie.Duration, &movie.Synopsis, &movie.GenreId, &movie.CasterId, &movie.CinemaId,
	)

	if err != nil {
		return movie, fmt.Errorf("creating movie: %w", err)
	}

	return movie, nil
}

// GetAllMovies retrieves movies with pagination and optional search filter
func GetAllMovies(params lib.MovieQueryParams) ([]lib.Movie, lib.PageInfo, error) {
	// Set defaults
	if params.Page <= 0 {
		params.Page = 1
	}
	if params.Limit <= 0 {
		params.Limit = 10
	}
	if params.Sort != "desc" {
		params.Sort = "asc"
	}

	// Try reading from Redis first
	cacheKey := fmt.Sprintf("movies:limit:%d:page:%d:search:%s:sort:%s", params.Limit, params.Page, params.Search, params.Sort)
	rdb := lib.Redis()

	if rdb != nil {
		cachedVal, err := rdb.Get(context.Background(), cacheKey).Result()
		if err == nil {
			var cachedData struct {
				Movies   []lib.Movie  `json:"movies"`
				PageInfo lib.PageInfo `json:"page_info"`
			}
			if json.Unmarshal([]byte(cachedVal), &cachedData) == nil {
				return cachedData.Movies, cachedData.PageInfo, nil
			}
		}
	}

	pgConn := lib.InitDB()
	defer pgConn.Close(context.Background())

	offset := (params.Page - 1) * params.Limit
	searchQuery := "%" + params.Search + "%"

	// 1. Get total data count for pagination metadata
	var totalData int
	countErr := pgConn.QueryRow(context.Background(), `
		SELECT COUNT(id) FROM movie 
		WHERE title ILIKE $1
	`, searchQuery).Scan(&totalData)
	
	if countErr != nil {
		return nil, lib.PageInfo{}, fmt.Errorf("counting movies: %w", countErr)
	}

	totalPage := (totalData + params.Limit - 1) / params.Limit // Ceil division
	
	nextPage := params.Page + 1
	if nextPage > totalPage {
		nextPage = 0
	}

	prevPage := params.Page - 1
	if prevPage < 1 {
		prevPage = 0
	}

	pageInfo := lib.PageInfo{
		CurentPage: params.Page,
		NextPage:   nextPage,
		PrevPage:   prevPage,
		TotalPage:  totalPage,
		TotalData:  totalData,
	}

	// 2. Query movies
	sortDir := "ASC"
	if params.Sort == "desc" {
		sortDir = "DESC"
	}
	query := fmt.Sprintf(`
		SELECT id, image, title, released_at, recommendation, duration, synopsis, genre_id, caster_id, cinema_id
		FROM movie
		WHERE title ILIKE $1
		ORDER BY id %s
		LIMIT $2 OFFSET $3
	`, sortDir)
	rows, err := pgConn.Query(context.Background(), query, searchQuery, params.Limit, offset)

	if err != nil {
		return nil, pageInfo, fmt.Errorf("querying movies: %w", err)
	}
	defer rows.Close()

	var movies []lib.Movie
	for rows.Next() {
		var movie lib.Movie
		err := rows.Scan(
			&movie.Id, &movie.Image, &movie.Title, &movie.ReleasedAt, &movie.Recommendation,
			&movie.Duration, &movie.Synopsis, &movie.GenreId, &movie.CasterId, &movie.CinemaId,
		)
		if err != nil {
			return nil, pageInfo, fmt.Errorf("scanning movie: %w", err)
		}
		movies = append(movies, movie)
	}

	if movies == nil {
		movies = []lib.Movie{}
	}

	// Save to Redis before returning
	if rdb != nil {
		cacheData := struct {
			Movies   []lib.Movie  `json:"movies"`
			PageInfo lib.PageInfo `json:"page_info"`
		}{
			Movies:   movies,
			PageInfo: pageInfo,
		}

		if cacheBytes, cacheErr := json.Marshal(cacheData); cacheErr == nil {
			rdb.Set(context.Background(), cacheKey, cacheBytes, 5*time.Minute)
		}
	}

	return movies, pageInfo, nil
}

// GetMovieById retrieves a specific movie by ID
func GetMovieById(id int) (lib.Movie, error) {
	pgConn := lib.InitDB()
	defer pgConn.Close(context.Background())

	var movie lib.Movie
	err := pgConn.QueryRow(context.Background(), `
		SELECT id, image, title, released_at, recommendation, duration, synopsis, genre_id, caster_id, cinema_id
		FROM movie
		WHERE id = $1
	`, id).Scan(
		&movie.Id, &movie.Image, &movie.Title, &movie.ReleasedAt, &movie.Recommendation,
		&movie.Duration, &movie.Synopsis, &movie.GenreId, &movie.CasterId, &movie.CinemaId,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return movie, fmt.Errorf("movie not found")
		}
		return movie, fmt.Errorf("querying movie: %w", err)
	}

	return movie, nil
}

// UpdateMovie updates a movie using COALESCE for optional fields (PATCH behavior)
func UpdateMovie(id int, req lib.MovieUpdateRequest) (lib.Movie, error) {
	pgConn := lib.InitDB()
	defer pgConn.Close(context.Background())

	var movie lib.Movie
	err := pgConn.QueryRow(context.Background(), `
		UPDATE movie 
		SET 
			image = COALESCE($1, image),
			title = COALESCE($2, title),
			released_at = COALESCE($3, released_at),
			recommendation = COALESCE($4, recommendation),
			duration = COALESCE($5, duration),
			synopsis = COALESCE($6, synopsis),
			genre_id = COALESCE($7, genre_id),
			caster_id = COALESCE($8, caster_id),
			cinema_id = COALESCE($9, cinema_id)
		WHERE id = $10
		RETURNING id, image, title, released_at, recommendation, duration, synopsis, genre_id, caster_id, cinema_id
	`, req.Image, req.Title, req.ReleasedAt, req.Recommendation, req.Duration, req.Synopsis, req.GenreId, req.CasterId, req.CinemaId, id).Scan(
		&movie.Id, &movie.Image, &movie.Title, &movie.ReleasedAt, &movie.Recommendation,
		&movie.Duration, &movie.Synopsis, &movie.GenreId, &movie.CasterId, &movie.CinemaId,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return movie, fmt.Errorf("movie not found")
		}
		return movie, fmt.Errorf("updating movie: %w", err)
	}

	return movie, nil
}

// DeleteMovie deletes a movie by ID
func DeleteMovie(id int) error {
	pgConn := lib.InitDB()
	defer pgConn.Close(context.Background())

	result, err := pgConn.Exec(context.Background(), `DELETE FROM movie WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("deleting movie: %w", err)
	}
	
	if result.RowsAffected() == 0 {
		return fmt.Errorf("movie not found")
	}

	return nil
}
