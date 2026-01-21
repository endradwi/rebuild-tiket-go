-- =========================
-- TABLE: role
-- =========================
CREATE TABLE "role" (
  id SERIAL PRIMARY KEY,
  name VARCHAR NOT NULL UNIQUE
);

-- =========================
-- TABLE: user
-- =========================
CREATE TABLE "user" (
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
  phone_number INT,
  point INT,
  tiket_status BOOLEAN,
  user_id INT,
  role_id INT,

  CONSTRAINT fk_profile_user
    FOREIGN KEY (user_id)
    REFERENCES "user"(id)
    ON DELETE CASCADE,

  CONSTRAINT fk_profile_role
    FOREIGN KEY (role_id)
    REFERENCES "role"(id)
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
-- TABLE: cinema
-- =========================
CREATE TABLE cinema (
  id BIGSERIAL PRIMARY KEY,
  cinema_name VARCHAR NOT NULL,
  images BIGINT,
  location VARCHAR,
  watching_time TIMESTAMP,
  watching_date BIGINT
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
  genre_id INT,
  caster_id INT,
  cinema_id INT,

  CONSTRAINT fk_movie_genre
    FOREIGN KEY (genre_id)
    REFERENCES genre(id),

  CONSTRAINT fk_movie_caster
    FOREIGN KEY (caster_id)
    REFERENCES caster(id),

  CONSTRAINT fk_movie_cinema
    FOREIGN KEY (cinema_id)
    REFERENCES cinema(id)
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
CREATE TABLE "order" (
  id BIGSERIAL PRIMARY KEY,
  profile_id INT,
  movie_id INT,
  seat_id INT,

  CONSTRAINT fk_order_profile
    FOREIGN KEY (profile_id)
    REFERENCES profile(id),

  CONSTRAINT fk_order_movie
    FOREIGN KEY (movie_id)
    REFERENCES movie(id),

  CONSTRAINT fk_order_seat
    FOREIGN KEY (seat_id)
    REFERENCES seat(id)
);

-- =========================
-- TABLE: payment
-- =========================
CREATE TABLE payment (
  id SERIAL PRIMARY KEY,
  total_payment INT,
  payment_method VARCHAR,
  expired_at TIMESTAMP,
  qr_code VARCHAR,
  payment_status VARCHAR,
  order_id INT,
  create_at TIMESTAMP,
  updated_at TIMESTAMP,

  CONSTRAINT fk_payment_order
    FOREIGN KEY (order_id)
    REFERENCES "order"(id)
);


INSERT INTO "role" (name) VALUES ('USER'), ('ADMIN');
