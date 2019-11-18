CREATE DATABASE companies;
\connect companies;
CREATE TABLE people
(
    id   bigserial NOT NULL,
    name text      NULL,
    CONSTRAINT people_pk PRIMARY KEY (id)

);

CREATE TABLE companies
(
    id     bigserial NOT NULL,
    name   text      NULL,
    ceo_id integer   NOT NULL REFERENCES people (id),
    CONSTRAINT companies_pk PRIMARY KEY (id)
);
