-- 0001_init.up.sql
-- Create base reference tables
CREATE TABLE IF NOT EXISTS ranks (
  rank_id SERIAL PRIMARY KEY,
  rank_code TEXT,
  rank_title TEXT
);

CREATE TABLE IF NOT EXISTS regions (
  region_id SERIAL PRIMARY KEY,
  name TEXT NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS formations (
  formation_id SERIAL PRIMARY KEY,
  region_id INT REFERENCES regions(region_id) ON DELETE SET NULL,
  name TEXT NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS postings_units (
  posting_id SERIAL PRIMARY KEY,
  name TEXT NOT NULL UNIQUE,
  description TEXT
);

-- Persons
CREATE TABLE IF NOT EXISTS person (
  person_id SERIAL PRIMARY KEY,
  regulation_number TEXT UNIQUE,
  first_name TEXT NOT NULL,
  middle_name TEXT,
  last_name TEXT NOT NULL,
  sex CHAR(1) CHECK (sex IN ('M','F')) ,
  rank_id INT REFERENCES ranks(rank_id),
  formation_id INT REFERENCES formations(formation_id),
  posting_id INT REFERENCES postings_units(posting_id),
  region_id INT REFERENCES regions(region_id),
  phone TEXT,
  email TEXT,
  is_active BOOLEAN DEFAULT TRUE,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Accounts (one per person optional)
CREATE TABLE IF NOT EXISTS account (
  account_id SERIAL PRIMARY KEY,
  person_id INT UNIQUE REFERENCES person(person_id) ON DELETE CASCADE,
  email TEXT NOT NULL UNIQUE,
  password_hash TEXT NOT NULL,
  role TEXT NOT NULL,
  is_active BOOLEAN DEFAULT TRUE,
  activation_token TEXT,
  pass_reset_token TEXT,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Courses
CREATE TABLE IF NOT EXISTS courses (
  course_id SERIAL PRIMARY KEY,
  code TEXT NOT NULL UNIQUE,
  title TEXT NOT NULL,
  description TEXT,
  category TEXT CHECK (category IN ('mandatory','elective','instructor')) DEFAULT 'mandatory',
  credit_hours NUMERIC,
  is_course_active BOOLEAN DEFAULT TRUE,
  rating NUMERIC(3,2),
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Course sessions (instances of a course)
CREATE TABLE IF NOT EXISTS course_sessions (
  course_session_id SERIAL PRIMARY KEY,
  course_id INT REFERENCES courses(course_id) ON DELETE CASCADE,
  start_date DATE NOT NULL,
  end_date DATE,
  location TEXT,
  region_id INT REFERENCES regions(region_id),
  formation_id INT REFERENCES formations(formation_id),
  capacity INT,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Facilitators
CREATE TABLE IF NOT EXISTS facilitators (
  facilitator_id SERIAL PRIMARY KEY,
  first_name TEXT,
  last_name TEXT,
  rank_id INT REFERENCES ranks(rank_id),
  posting_id INT REFERENCES postings_units(posting_id),
  phone_number TEXT,
  email TEXT,
  is_police BOOLEAN DEFAULT FALSE,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Participant: junction table person <-> course_session
CREATE TABLE IF NOT EXISTS participant (
  participant_id SERIAL PRIMARY KEY,
  course_session_id INT REFERENCES course_sessions(course_session_id) ON DELETE CASCADE,
  person_id INT REFERENCES person(person_id) ON DELETE CASCADE,
  status TEXT CHECK (status IN ('completed','pending','failed')) DEFAULT 'pending',
  credit_awarded NUMERIC DEFAULT 0,
  note TEXT,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  UNIQUE (course_session_id, person_id)
);

-- Helpful indexes
CREATE INDEX IF NOT EXISTS idx_person_last_name ON person(last_name);
CREATE INDEX IF NOT EXISTS idx_courses_code ON courses(code);
CREATE INDEX IF NOT EXISTS idx_cs_course ON course_sessions(course_id);
CREATE INDEX IF NOT EXISTS idx_participant_course ON participant(course_session_id);
