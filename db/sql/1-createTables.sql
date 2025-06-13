BEGIN;

-- SET client_encoding = 'LATIN1';

CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    name TEXT,
    email TEXT NOT NULL UNIQUE,
    passwordHash TEXT NOT NULL,
    isAdmin BOOLEAN NOT NULL DEFAULT FALSE
);

COMMIT;
