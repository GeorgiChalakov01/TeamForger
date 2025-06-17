BEGIN;

CREATE EXTENSION IF NOT EXISTS vector;

CREATE TABLE users (
	id SERIAL PRIMARY KEY,
	name TEXT,
	email TEXT NOT NULL UNIQUE,
	passwordHash TEXT NOT NULL,
	sessionToken TEXT NOT NULL,
	csrfToken TEXT NOT NULL,
	isAdmin BOOLEAN NOT NULL DEFAULT FALSE,
	cv TEXT
);

CREATE TABLE cv_chunks (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    chunk TEXT NOT NULL,
    embedding vector(768) NOT NULL
);

CREATE INDEX ON cv_chunks USING hnsw (embedding vector_cosine_ops);

COMMIT;
