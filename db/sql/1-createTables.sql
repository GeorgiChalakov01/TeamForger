BEGIN;

-- SET client_encoding = 'LATIN1';

CREATE TABLE users (
	id integer NOT NULL,
	name text,
	email text NOT NULL,
	passwordHash text NOT NULL,
	isAdmin bool NOT NULL DEFAULT False
);

ALTER TABLE ONLY users
	ADD CONSTRAINT users_pkey PRIMARY KEY (id);

COMMIT;
