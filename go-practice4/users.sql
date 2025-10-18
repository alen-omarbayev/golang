CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    email TEXT UNIQUE NOT NULL,
    balance NUMERIC(10,2) DEFAULT 0
);

INSERT INTO users (name, email, balance) VALUES
('Alen Omarbayev', 'alen@example.com', 1000),
('Cristiano Ronaldo', 'Ronaldo@example.com', 800);