package models

import (
	"context"
	"encoding/json"
	"fmt"
	"tiket/lib"
	"time"

	"github.com/jackc/pgx/v5"
)

// CreateMovie inserts a new movie into the database with many-to-many relations
func CreateMovie(req lib.MovieCreateRequest) (lib.Movie, error) {
	pgConn := lib.InitDB()
	defer pgConn.Close(context.Background())

	tx, err := pgConn.Begin(context.Background())
	if err != nil {
		return lib.Movie{}, fmt.Errorf("beginning transaction: %w", err)
	}
	defer tx.Rollback(context.Background())

	var movie lib.Movie
	// We use the first ID from GenreIds/CasterIds for the legacy single-column fields if they exist
	var legacyGenreId, legacyCasterId *int
	if len(req.GenreIds) > 0 { legacyGenreId = &req.GenreIds[0] }
	if len(req.CasterIds) > 0 { legacyCasterId = &req.CasterIds[0] }

	err = tx.QueryRow(context.Background(), `
		INSERT INTO movie (image, title, released_at, recommendation, duration, synopsis, director_name, genre_id, caster_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, image, title, released_at, recommendation, duration, synopsis, director_name, genre_id, caster_id
	`, req.Image, req.Title, req.ReleasedAt, req.Recommendation, req.Duration, req.Synopsis, req.DirectorName, legacyGenreId, legacyCasterId).Scan(
		&movie.Id, &movie.Image, &movie.Title, &movie.ReleasedAt, &movie.Recommendation,
		&movie.Duration, &movie.Synopsis, &movie.DirectorName, &movie.GenreId, &movie.CasterId,
	)

	if err != nil {
		return movie, fmt.Errorf("inserting movie: %w", err)
	}

	// Insert into junction tables
	for _, gid := range req.GenreIds {
		_, err = tx.Exec(context.Background(), `INSERT INTO movie_genres (movie_id, genre_id) VALUES ($1, $2)`, movie.Id, gid)
		if err != nil { return movie, fmt.Errorf("inserting movie genre: %w", err) }
	}

	for _, cid := range req.CasterIds {
		_, err = tx.Exec(context.Background(), `INSERT INTO movie_casters (movie_id, caster_id) VALUES ($1, $2)`, movie.Id, cid)
		if err != nil { return movie, fmt.Errorf("inserting movie caster: %w", err) }
	}

	for _, cid := range req.CinemaIds {
		_, err = tx.Exec(context.Background(), `INSERT INTO movie_cinemas (movie_id, cinema_id) VALUES ($1, $2)`, movie.Id, cid)
		if err != nil { return movie, fmt.Errorf("inserting movie cinema: %w", err) }

		// Add showtimes for each cinema
		for _, st := range req.Showtimes {
			for _, timeVal := range st.Times {
				_, err = tx.Exec(context.Background(), `
					INSERT INTO movie_showtimes (movie_id, cinema_id, show_date, show_time, price)
					VALUES ($1, $2, $3, $4, $5)
				`, movie.Id, cid, st.Date, timeVal, st.Price)
				if err != nil { return movie, fmt.Errorf("inserting movie showtime: %w", err) }
			}
		}
	}

	err = tx.Commit(context.Background())
	if err != nil {
		return movie, fmt.Errorf("committing transaction: %w", err)
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
	
	whereClause := "WHERE title ILIKE $1"
	args := []interface{}{searchQuery, params.Limit, offset}
	if params.Month > 0 {
		whereClause += fmt.Sprintf(" AND EXTRACT(MONTH FROM released_at) = $%d", len(args)+1)
		args = append(args, params.Month)
	}
	if params.Year > 0 {
		whereClause += fmt.Sprintf(" AND EXTRACT(YEAR FROM released_at) = $%d", len(args)+1)
		args = append(args, params.Year)
	}

	query := fmt.Sprintf(`
		SELECT id, image, title, released_at, recommendation, duration, synopsis, director_name, genre_id, caster_id
		FROM movie
		%s
		ORDER BY id %s
		LIMIT $2 OFFSET $3
	`, whereClause, sortDir)
	rows, err := pgConn.Query(context.Background(), query, args...)

	if err != nil {
		return nil, pageInfo, fmt.Errorf("querying movies: %w", err)
	}
	defer rows.Close()

	var movies []lib.Movie
	for rows.Next() {
		var movie lib.Movie
		err := rows.Scan(
			&movie.Id, &movie.Image, &movie.Title, &movie.ReleasedAt, &movie.Recommendation,
			&movie.Duration, &movie.Synopsis, &movie.DirectorName, &movie.GenreId, &movie.CasterId,
		)
		if err != nil {
			return nil, pageInfo, fmt.Errorf("scanning movie: %w", err)
		}
		movie.Genres, movie.Casters, _ = GetMovieRelations(movie.Id)
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

func GetMovieRelations(movieId int) ([]string, []string, error) {
	pgConn := lib.InitDB()
	defer pgConn.Close(context.Background())

	var genres []string
	rows, _ := pgConn.Query(context.Background(), `
		SELECT g.name FROM movie_genres mg JOIN genre g ON mg.genre_id = g.id WHERE mg.movie_id = $1
	`, movieId)
	if rows != nil {
		defer rows.Close()
		for rows.Next() {
			var g string
			_ = rows.Scan(&g)
			genres = append(genres, g)
		}
	}

	var casters []string
	rows, _ = pgConn.Query(context.Background(), `
		SELECT c.name FROM movie_casters mc JOIN caster c ON mc.caster_id = c.id WHERE mc.movie_id = $1
	`, movieId)
	if rows != nil {
		defer rows.Close()
		for rows.Next() {
			var c string
			_ = rows.Scan(&c)
			casters = append(casters, c)
		}
	}

	return genres, casters, nil
}

// GetMovieById retrieves a specific movie by ID
func GetMovieById(id int) (lib.Movie, error) {
	pgConn := lib.InitDB()
	defer pgConn.Close(context.Background())

	var movie lib.Movie
	err := pgConn.QueryRow(context.Background(), `
		SELECT id, image, title, released_at, recommendation, duration, synopsis, director_name, genre_id, caster_id
		FROM movie
		WHERE id = $1
	`, id).Scan(
		&movie.Id, &movie.Image, &movie.Title, &movie.ReleasedAt, &movie.Recommendation,
		&movie.Duration, &movie.Synopsis, &movie.DirectorName, &movie.GenreId, &movie.CasterId,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return movie, fmt.Errorf("movie not found")
		}
		return movie, fmt.Errorf("querying movie: %w", err)
	}

	movie.Genres, movie.Casters, _ = GetMovieRelations(movie.Id)

	return movie, nil
}

// UpdateMovie updates a movie using COALESCE for optional fields and handles junction tables
func UpdateMovie(id int, req lib.MovieUpdateRequest) (lib.Movie, error) {
	pgConn := lib.InitDB()
	defer pgConn.Close(context.Background())

	tx, err := pgConn.Begin(context.Background())
	if err != nil {
		return lib.Movie{}, fmt.Errorf("beginning transaction: %w", err)
	}
	defer tx.Rollback(context.Background())

	var movie lib.Movie
	// Legacy single-column update if provided
	var legacyGenreId, legacyCasterId *int
	if len(req.GenreIds) > 0 { legacyGenreId = &req.GenreIds[0] }
	if len(req.CasterIds) > 0 { legacyCasterId = &req.CasterIds[0] }

	err = tx.QueryRow(context.Background(), `
		UPDATE movie 
		SET 
			image = COALESCE($1, image),
			title = COALESCE($2, title),
			released_at = COALESCE($3, released_at),
			recommendation = COALESCE($4, recommendation),
			duration = COALESCE($5, duration),
			synopsis = COALESCE($6, synopsis),
			director_name = COALESCE($7, director_name),
			genre_id = COALESCE($8, genre_id),
			caster_id = COALESCE($9, caster_id)
		WHERE id = $10
		RETURNING id, image, title, released_at, recommendation, duration, synopsis, director_name, genre_id, caster_id
	`, req.Image, req.Title, req.ReleasedAt, req.Recommendation, req.Duration, req.Synopsis, req.DirectorName, legacyGenreId, legacyCasterId, id).Scan(
		&movie.Id, &movie.Image, &movie.Title, &movie.ReleasedAt, &movie.Recommendation,
		&movie.Duration, &movie.Synopsis, &movie.DirectorName, &movie.GenreId, &movie.CasterId,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return movie, fmt.Errorf("movie not found")
		}
		return movie, fmt.Errorf("updating movie record: %w", err)
	}

	// Update junction tables (Clear and Re-insert)
	if len(req.GenreIds) > 0 {
		_, _ = tx.Exec(context.Background(), `DELETE FROM movie_genres WHERE movie_id = $1`, id)
		for _, gid := range req.GenreIds {
			_, err = tx.Exec(context.Background(), `INSERT INTO movie_genres (movie_id, genre_id) VALUES ($1, $2)`, id, gid)
			if err != nil { return movie, fmt.Errorf("updating movie genre: %w", err) }
		}
	}

	if len(req.CasterIds) > 0 {
		_, _ = tx.Exec(context.Background(), `DELETE FROM movie_casters WHERE movie_id = $1`, id)
		for _, cid := range req.CasterIds {
			_, err = tx.Exec(context.Background(), `INSERT INTO movie_casters (movie_id, caster_id) VALUES ($1, $2)`, id, cid)
			if err != nil { return movie, fmt.Errorf("updating movie caster: %w", err) }
		}
	}

	if len(req.CinemaIds) > 0 {
		_, _ = tx.Exec(context.Background(), `DELETE FROM movie_cinemas WHERE movie_id = $1`, id)
		for _, cid := range req.CinemaIds {
			_, err = tx.Exec(context.Background(), `INSERT INTO movie_cinemas (movie_id, cinema_id) VALUES ($1, $2)`, id, cid)
			if err != nil { return movie, fmt.Errorf("updating movie cinema: %w", err) }
		}
	}

	err = tx.Commit(context.Background())
	if err != nil {
		return movie, fmt.Errorf("committing transaction: %w", err)
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
