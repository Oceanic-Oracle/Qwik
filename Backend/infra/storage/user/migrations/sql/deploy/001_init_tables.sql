CREATE TABLE roles (
    id SERIAL PRIMARY KEY,
    role TEXT NOT NULL UNIQUE
);

CREATE TABLE users (
    id UUID,
    login TEXT NOT NULL UNIQUE,
    password TEXT NOT NULL,
    surname TEXT,
    name TEXT,
    patronymic TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    shardEmail INTEGER
);

CREATE TABLE users_email (
    id_users UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email TEXT NOT NULL UNIQUE
);

CREATE TABLE usersroles (
    id SERIAL PRIMARY KEY,
    user_id UUID NOT NULL,
    role_id INTEGER NOT NULL,
    FOREIGN KEY (role_id) REFERENCES roles(id)
);