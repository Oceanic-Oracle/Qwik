CREATE EXTENSION IF NOT EXISTS postgis;

CREATE TABLE warehouse (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    location GEOGRAPHY(POINT, 4326),
    address TEXT
);

CREATE TABLE rack (
    id SERIAL PRIMARY KEY,
    warehouse_id INTEGER NOT NULL REFERENCES warehouse(id) ON DELETE CASCADE,
    aisle TEXT
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
    id SERIAL PRIMARY KEY,
    preview_url TEXT,
    name TEXT NOT NULL,
    description TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    visibility BOOLEAN DEFAULT true
);

CREATE TABLE file (
    id SERIAL PRIMARY KEY,
    product_id INTEGER NOT NULL REFERENCES product(id) ON DELETE CASCADE,
    url TEXT NOT NULL
);

CREATE TABLE review (
    id SERIAL PRIMARY KEY,
    login TEXT NOT NULL,
    grade INTEGER NOT NULL CHECK (grade BETWEEN 1 AND 5),
    description TEXT,
    product_id INTEGER NOT NULL REFERENCES product(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE (login, product_id)
);

CREATE TABLE tag (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL UNIQUE
);

CREATE TABLE product_tag (
    id SERIAL PRIMARY KEY,
    product_id INTEGER NOT NULL REFERENCES product(id) ON DELETE CASCADE,
    tag_id INTEGER NOT NULL REFERENCES tag(id) ON DELETE CASCADE,
    UNIQUE (product_id, tag_id)
);

CREATE TABLE shelf_product (
    id SERIAL PRIMARY KEY,
    shelf_id INTEGER NOT NULL REFERENCES shelf(id) ON DELETE CASCADE,
    product_id INTEGER NOT NULL REFERENCES product(id) ON DELETE CASCADE,
    allocated_capacity NUMERIC NOT NULL CHECK (allocated_capacity > 0),
    UNIQUE (shelf_id, product_id)
);