CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    email TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE users_favorites (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    hotel_id TEXT NOT NULL,
    target_price NUMERIC(10,2) NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    UNIQUE (user_id, hotel_id)
);