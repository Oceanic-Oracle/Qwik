CREATE EXTENSION IF NOT EXISTS postgis;
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE rack (
    id SERIAL PRIMARY KEY,
    aisle TEXT NOT NULL
);

CREATE TABLE shelf (
    id SERIAL PRIMARY KEY,
    rack_id INTEGER NOT NULL REFERENCES rack(id) ON DELETE CASCADE,
    level INTEGER NOT NULL,
    priority INTEGER NOT NULL,
    max_capacity NUMERIC NOT NULL,
    used_capacity NUMERIC NOT NULL DEFAULT 0,
    CONSTRAINT check_capacity CHECK (used_capacity <= max_capacity)
);

CREATE TABLE product (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    preview_url TEXT,
    name TEXT NOT NULL,
    description TEXT,
    price INTEGER,
    width NUMERIC,
    height NUMERIC,
    depth NUMERIC,
    weight NUMERIC,
    volume NUMERIC,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    visibility BOOLEAN DEFAULT true
);

CREATE TABLE file (
    id SERIAL PRIMARY KEY,
    product_id UUID NOT NULL REFERENCES product(id) ON DELETE CASCADE,
    url TEXT NOT NULL
);

CREATE TABLE review (
    id SERIAL PRIMARY KEY,
    login TEXT,
    grade INTEGER NOT NULL CHECK (grade BETWEEN 1 AND 5),
    description TEXT,
    product_id UUID NOT NULL REFERENCES product(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE UNIQUE INDEX unique_review_with_login 
ON review (login, product_id) 
WHERE login IS NOT NULL;

CREATE TABLE tag (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL UNIQUE
);

CREATE TABLE product_tag (
    id SERIAL PRIMARY KEY,
    product_id UUID NOT NULL REFERENCES product(id) ON DELETE CASCADE,
    tag_id INTEGER NOT NULL REFERENCES tag(id) ON DELETE CASCADE,
    UNIQUE (product_id, tag_id)
);

CREATE TABLE shelf_product (
    id SERIAL PRIMARY KEY,
    shelf_id INTEGER NOT NULL REFERENCES shelf(id) ON DELETE CASCADE,
    product_id UUID NOT NULL REFERENCES product(id) ON DELETE CASCADE,
    quantity INTEGER NOT NULL CHECK (quantity > 0),
    allocated_capacity NUMERIC NOT NULL,
    UNIQUE (shelf_id, product_id)
);