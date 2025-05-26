CREATE TABLE __MigrationsHistory (
    id SERIAL PRIMARY KEY,
    fileName TEXT NOT NULL UNIQUE,
    hash TEXT NOT NULL UNIQUE
);