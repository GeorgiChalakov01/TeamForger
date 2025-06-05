--
-- PostgreSQL port of the MySQL "developerDB" database.
--

BEGIN;

SET client_encoding = 'LATIN1';

CREATE TABLE developers (
    id integer NOT NULL,
    name text NOT NULL,
    email text NOT NULL,
    role text NOT NULL,
    status text NOT NULL,
    team text NOT NULL
);



COPY developers (id, name, email, role, status, team) FROM stdin(Delimiter ',');
2,Harry Potter,harry@chosenone.com,Backend,Full Time,Team G
1,Albus Dumbledore,hogwarts.headmaster@gryffindor.com,Fullstack,Not available,Team G
3,Tom Riddle,thedarklord@horcrux.com,Frontend,Contractor,Team S
4,Luna Lovegood,blibbering_humdinger@dumbledorearmy.com,Fullstack,Full Time,Team R
\.

ALTER TABLE ONLY developers
    ADD CONSTRAINT developer_pkey PRIMARY KEY (id);


COMMIT;

ANALYZE developers;
