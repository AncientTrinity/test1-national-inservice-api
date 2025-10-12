-- Filename: migrations/0001_init.up.sql
CREATE TABLE IF NOT EXISTS rank (
  rank_id SERIAL PRIMARY KEY,
  rank_code TEXT,
  rank_title TEXT
);

CREATE TABLE IF NOT EXISTS region (
  region_id SERIAL PRIMARY KEY,
  name TEXT NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS formations (
  formation_id SERIAL PRIMARY KEY,
  region_id INT REFERENCES region(region_id),
  name TEXT NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS postings_units (
  posting_id SERIAL PRIMARY KEY,
  name TEXT NOT NULL UNIQUE,
  description TEXT
);

CREATE TABLE IF NOT EXISTS person (
  person_id SERIAL PRIMARY KEY,
  regulation_number TEXT UNIQUE,
  first_name TEXT NOT NULL,
  middle_name TEXT,
  last_name TEXT NOT NULL,
  sex CHAR(1) CHECK (sex IN ('M','F')),
  rank_id INT REFERENCES rank(rank_id),
  formation_id INT REFERENCES formations(formation_id),
  posting_id INT REFERENCES postings_units(posting_id),
  region_id INT REFERENCES region(region_id),
  phone TEXT,
  email TEXT,
  is_active BOOLEAN DEFAULT TRUE,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
