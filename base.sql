DROP TABLE IF EXISTS payment CASCADE;
DROP TABLE IF EXISTS order_seats CASCADE;
DROP TABLE IF EXISTS orders CASCADE;
DROP TABLE IF EXISTS movie_showtimes CASCADE;
DROP TABLE IF EXISTS movie CASCADE;
DROP TABLE IF EXISTS seat CASCADE;
DROP TABLE IF EXISTS cinema CASCADE;
DROP TABLE IF EXISTS location CASCADE;
DROP TABLE IF EXISTS caster CASCADE;
DROP TABLE IF EXISTS genre CASCADE;
DROP TABLE IF EXISTS reset_password CASCADE;
DROP TABLE IF EXISTS profile CASCADE;
DROP TABLE IF EXISTS users CASCADE;
DROP TABLE IF EXISTS role CASCADE;

-- =========================
-- TABLE: role
-- =========================
CREATE TABLE role (
  id SERIAL PRIMARY KEY,
  name VARCHAR NOT NULL UNIQUE
);

-- =========================
-- TABLE: user
-- =========================
CREATE TABLE users (
  id SERIAL PRIMARY KEY,
  email VARCHAR NOT NULL UNIQUE,
  password VARCHAR NOT NULL
);


-- =========================
-- TABLE: profile
-- =========================
CREATE TABLE profile (
  id SERIAL PRIMARY KEY,
  first_name VARCHAR,
  last_name VARCHAR,
  phone_number VARCHAR,
image VARCHAR,
  point INT,
  tiket_status BOOLEAN,
  user_id INT,
  role_id INT,

  CONSTRAINT fk_profile_user
    FOREIGN KEY (user_id)
    REFERENCES users(id)
    ON DELETE CASCADE,

  CONSTRAINT fk_profile_role
    FOREIGN KEY (role_id)
    REFERENCES role(id)
);

-- =========================
-- TABLE: reset_password
-- =========================
CREATE TABLE reset_password (
  id SERIAL PRIMARY KEY,
  profile_id INT,
  token_hash VARCHAR NOT NULL,
  expired_at TIMESTAMP NOT NULL,
  used_at TIMESTAMP,
  created_at TIMESTAMP DEFAULT NOW(),

  CONSTRAINT fk_reset_profile
    FOREIGN KEY (profile_id)
    REFERENCES profile(id)
    ON DELETE CASCADE
);

-- =========================
-- TABLE: genre
-- =========================
CREATE TABLE genre (
  id SERIAL PRIMARY KEY,
  name VARCHAR NOT NULL
);

-- =========================
-- TABLE: caster
-- =========================
CREATE TABLE caster (
  id SERIAL PRIMARY KEY,
  name VARCHAR NOT NULL
);

-- =========================
-- TABLE: location
-- =========================
CREATE TABLE location (
  id SERIAL PRIMARY KEY,
  name VARCHAR NOT NULL
);

-- =========================
-- TABLE: cinema
-- =========================
CREATE TABLE cinema (
  id BIGSERIAL PRIMARY KEY,
  cinema_name VARCHAR NOT NULL,
  image VARCHAR,
  location_id INT,

  CONSTRAINT fk_cinema_location
    FOREIGN KEY (location_id)
    REFERENCES location(id)
);

-- =========================
-- TABLE: movie
-- =========================
CREATE TABLE movie (
  id SERIAL PRIMARY KEY,
  image VARCHAR,
  title VARCHAR NOT NULL,
  released_at TIMESTAMP,
  recommendation BOOLEAN,
  duration TIME,
  synopsis VARCHAR,
  director_name VARCHAR,
  genre_id INT,
  caster_id INT,

  CONSTRAINT fk_movie_genre
    FOREIGN KEY (genre_id)
    REFERENCES genre(id),

  CONSTRAINT fk_movie_caster
    FOREIGN KEY (caster_id)
    REFERENCES caster(id)
);

-- =========================
-- MANY-TO-MANY JUNCTIONS
-- =========================

CREATE TABLE movie_genres (
  movie_id INT NOT NULL,
  genre_id INT NOT NULL,
  PRIMARY KEY (movie_id, genre_id),
  FOREIGN KEY (movie_id) REFERENCES movie(id) ON DELETE CASCADE,
  FOREIGN KEY (genre_id) REFERENCES genre(id) ON DELETE CASCADE
);

CREATE TABLE movie_casters (
  movie_id INT NOT NULL,
  caster_id INT NOT NULL,
  PRIMARY KEY (movie_id, caster_id),
  FOREIGN KEY (movie_id) REFERENCES movie(id) ON DELETE CASCADE,
  FOREIGN KEY (caster_id) REFERENCES caster(id) ON DELETE CASCADE
);

CREATE TABLE movie_cinemas (
  movie_id INT NOT NULL,
  cinema_id INT NOT NULL,
  PRIMARY KEY (movie_id, cinema_id),
  FOREIGN KEY (movie_id) REFERENCES movie(id) ON DELETE CASCADE,
  FOREIGN KEY (cinema_id) REFERENCES cinema(id) ON DELETE CASCADE
);

-- =========================
-- TABLE: movie_showtimes
-- =========================
CREATE TABLE movie_showtimes (
  id SERIAL PRIMARY KEY,
  movie_id INT NOT NULL,
  cinema_id INT NOT NULL,
  show_date DATE NOT NULL,
  show_time TIME NOT NULL,
  price INT NOT NULL,

  CONSTRAINT fk_showtime_movie
    FOREIGN KEY (movie_id)
    REFERENCES movie(id)
    ON DELETE CASCADE,

  CONSTRAINT fk_showtime_cinema
    FOREIGN KEY (cinema_id)
    REFERENCES cinema(id)
    ON DELETE CASCADE
);

-- =========================
-- TABLE: seat
-- =========================
CREATE TABLE seat (
  id SERIAL PRIMARY KEY,
  name VARCHAR NOT NULL,
  price INT NOT NULL
);

-- =========================
-- TABLE: order
-- =========================
CREATE TABLE orders (
  id BIGSERIAL PRIMARY KEY,
  order_number VARCHAR UNIQUE NOT NULL,
  profile_id INT,
  showtime_id INT,
  full_name VARCHAR,
  email VARCHAR,
  phone_number VARCHAR,
  total_price INT,
  status VARCHAR DEFAULT 'pending', -- pending, paid, cancelled
  created_at TIMESTAMP DEFAULT NOW(),
  updated_at TIMESTAMP DEFAULT NOW(),

  CONSTRAINT fk_order_profile
    FOREIGN KEY (profile_id)
    REFERENCES profile(id),

  CONSTRAINT fk_order_showtime
    FOREIGN KEY (showtime_id)
    REFERENCES movie_showtimes(id)
);

-- =========================
-- TABLE: order_seats
-- =========================
CREATE TABLE order_seats (
  id BIGSERIAL PRIMARY KEY,
  order_id BIGINT NOT NULL,
  seat_id INT NOT NULL,

  CONSTRAINT fk_order_seats_order
    FOREIGN KEY (order_id)
    REFERENCES orders(id)
    ON DELETE CASCADE,

  CONSTRAINT fk_order_seats_seat
    FOREIGN KEY (seat_id)
    REFERENCES seat(id)
);

-- =========================
-- TABLE: payment
-- =========================
CREATE TABLE payment (
  id SERIAL PRIMARY KEY,
  order_id BIGINT UNIQUE NOT NULL,
  total_payment INT,
  payment_method VARCHAR,
  payment_status VARCHAR DEFAULT 'pending', -- pending, success, failed
  expired_at TIMESTAMP,
  qr_code VARCHAR,
  created_at TIMESTAMP DEFAULT NOW(),
  updated_at TIMESTAMP DEFAULT NOW(),

  CONSTRAINT fk_payment_order
    FOREIGN KEY (order_id)
    REFERENCES orders(id)
    ON DELETE CASCADE
);


INSERT INTO role (name) VALUES ('USER'), ('ADMIN');

-- -------------------------
-- default user (Seed)
-- -------------------------
INSERT INTO users (id, email, password) 
VALUES (1, 'test@example.com', '$argon2id$v=19$m=65536,t=3,p=4$vP/WfW9lK9vY7K/k9I1fBw$Z7Q9Z6a9Z6a9Z6a9Z6a9Z6a9Z6a9Z6a9Z6a9Z6a9Z6a9Z6a'), -- password: password123 (hashed)
       (2, 'admin@example.com', '$argon2id$v=19$m=65536,t=3,p=4$vP/WfW9lK9vY7K/k9I1fBw$Z7Q9Z6a9Z6a9Z6a9Z6a9Z6a9Z6a9Z6a9Z6a9Z6a9Z6a9Z6a') -- password: password123 (hashed)
ON CONFLICT (id) DO NOTHING;

INSERT INTO profile (id, user_id, role_id, first_name, last_name)
VALUES (1, 1, 1, 'Test', 'User'), (2, 2, 2, 'Admin', 'User')
ON CONFLICT (id) DO NOTHING;


-- =============================================================
-- SEED DATA
-- Safe to re-run: only inserts if table is empty
-- =============================================================

-- -------------------------
-- genre (10 rows)
-- -------------------------
INSERT INTO genre (name)
SELECT name FROM (VALUES
  ('Action'),
  ('Adventure'),
  ('Animation'),
  ('Comedy'),
  ('Drama'),
  ('Horror'),
  ('Romance'),
  ('Sci-Fi'),
  ('Thriller'),
  ('Fantasy')
) AS v(name)
WHERE NOT EXISTS (SELECT 1 FROM genre LIMIT 1);

-- -------------------------
-- caster (10 rows)
-- -------------------------
INSERT INTO caster (name)
SELECT name FROM (VALUES
  ('Leonardo DiCaprio'),
  ('Scarlett Johansson'),
  ('Tom Hanks'),
  ('Natalie Portman'),
  ('Robert Downey Jr.'),
  ('Meryl Streep'),
  ('Chris Evans'),
  ('Cate Blanchett'),
  ('Brad Pitt'),
  ('Margot Robbie')
) AS v(name)
WHERE NOT EXISTS (SELECT 1 FROM caster LIMIT 1);

-- -------------------------
-- location (10 rows)
-- -------------------------
INSERT INTO location (name)
SELECT name FROM (VALUES
  ('Jakarta Pusat'),
  ('Jakarta Selatan'),
  ('Jakarta Barat'),
  ('Jakarta Utara'),
  ('Jakarta Timur'),
  ('Depok'),
  ('Bekasi'),
  ('Tangerang'),
  ('Bogor'),
  ('Bandung')
) AS v(name)
WHERE NOT EXISTS (SELECT 1 FROM location LIMIT 1);

-- -------------------------
-- cinema (20 rows)
-- -------------------------
INSERT INTO cinema (cinema_name, image, location_id)
SELECT cinema_name, image, location_id FROM (VALUES
  ('ebv.id',                'https://picsum.photos/seed/cinema_1/800/450', 1),
  ('hiflix',                'https://picsum.photos/seed/cinema_2/800/450', 2),
  ('CineOne21',             'https://picsum.photos/seed/cinema_3/800/450', 3),
  ('XXI Plaza Senayan',     'https://picsum.photos/seed/cinema_4/800/450', 2),
  ('CGV Grand Indonesia',   'https://picsum.photos/seed/cinema_5/800/450', 1),
  ('Cinepolis Lippo Mall',  'https://picsum.photos/seed/cinema_6/800/450', 6),
  ('CGV Transmart Cibubur', 'https://picsum.photos/seed/cinema_7/800/450', 7),
  ('XXI Summarecon Mal',    'https://picsum.photos/seed/cinema_8/800/450', 7),
  ('XXI Senayan City',      'https://picsum.photos/seed/cinema_9/800/450', 2),
  ('CGV Pacific Place',     'https://picsum.photos/seed/cinema_10/800/450', 2)
) AS v(cinema_name, image, location_id)
WHERE NOT EXISTS (SELECT 1 FROM cinema LIMIT 1);

-- -------------------------
-- movie (100 rows)
-- -------------------------
INSERT INTO movie (image, title, released_at, recommendation, duration, synopsis, genre_id, caster_id)
SELECT image, title, released_at, recommendation, duration, synopsis, genre_id, caster_id
FROM (VALUES
  ('https://picsum.photos/seed/movie_1/400/600', 'Spider-Man: Homecoming', '2017-06-28'::TIMESTAMP, TRUE,  '02:13:00'::TIME, 'Peter Parker tries to balance his life as an ordinary high school student in Queens with his superhero alter-ego Spider-Man.', 1, 5),
  ('https://picsum.photos/seed/movie_2/400/600', 'Inception',             '2010-07-16'::TIMESTAMP, TRUE,  '02:28:00'::TIME, 'A thief who enters the dreams of others.',            8, 1),
  ('https://picsum.photos/seed/movie_3/400/600', 'The Dark Knight',       '2008-07-18'::TIMESTAMP, TRUE,  '02:32:00'::TIME, 'Batman faces the Joker in Gotham City.',               1, 3),
  ('https://picsum.photos/seed/movie_4/400/600', 'Avengers: Endgame',     '2019-04-26'::TIMESTAMP, TRUE,  '03:02:00'::TIME, 'The Avengers assemble for a final showdown.',          1, 5)
) AS v(image, title, released_at, recommendation, duration, synopsis, genre_id, caster_id)
WHERE NOT EXISTS (SELECT 1 FROM movie LIMIT 1);

-- -------------------------
-- movie_showtimes (Seed)
-- -------------------------
INSERT INTO movie_showtimes (movie_id, cinema_id, show_date, show_time, price)
SELECT movie_id, cinema_id, show_date, show_time, price FROM (VALUES
  (1, 1, '2026-07-21'::DATE, '08:30:00'::TIME, 50000),
  (1, 1, '2026-07-21'::DATE, '10:30:00'::TIME, 50000),
  (1, 2, '2026-07-21'::DATE, '08:30:00'::TIME, 55000),
  (1, 3, '2026-07-21'::DATE, '13:00:00'::TIME, 45000),
  (2, 1, '2026-07-21'::DATE, '14:00:00'::TIME, 50000),
  (3, 2, '2026-07-21'::DATE, '16:00:00'::TIME, 60000)
) AS v(movie_id, cinema_id, show_date, show_time, price)
WHERE NOT EXISTS (SELECT 1 FROM movie_showtimes LIMIT 1);

-- -------------------------
-- seat (Seed)
-- -------------------------
INSERT INTO seat (name, price)
SELECT name, price FROM (VALUES
  ('A1', 50000), ('A2', 50000), ('A3', 50000), ('A4', 50000),
  ('B1', 50000), ('B2', 50000), ('B3', 50000), ('B4', 50000),
  ('C1', 50000), ('C2', 50000), ('C3', 50000), ('C4', 50000),
  ('D1', 50000), ('D2', 50000), ('D3', 50000), ('D4', 50000)
) AS v(name, price)
WHERE NOT EXISTS (SELECT 1 FROM seat LIMIT 1);
