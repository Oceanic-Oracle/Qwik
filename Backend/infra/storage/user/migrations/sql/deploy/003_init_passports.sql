CREATE TABLE user_passports (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    passport_series VARCHAR(4) NOT NULL,
    passport_number VARCHAR(6) NOT NULL,
    passport_issued_by TEXT NOT NULL,
    passport_issue_date DATE NOT NULL,
    passport_division_code VARCHAR(7),
    birth_date DATE NOT NULL,
    birth_place TEXT,
    registration_address TEXT NOT NULL,
    inn VARCHAR(12) UNIQUE,
    snils VARCHAR(14),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(passport_series, passport_number)
); 