CREATE TABLE user_adress (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    country VARCHAR(100) DEFAULT 'Россия',
    city VARCHAR(100) NOT NULL,
    street VARCHAR(200) NOT NULL,
    house VARCHAR(20) NOT NULL,
    apartment VARCHAR(20),
    is_primary BOOLEAN DEFAULT false,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
); 