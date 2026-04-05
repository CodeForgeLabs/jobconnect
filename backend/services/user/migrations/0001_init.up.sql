CREATE TABLE IF NOT EXISTS profiles (
    id SERIAL PRIMARY KEY,
    user_id UUID UNIQUE NOT NULL,
    role VARCHAR NOT NULL,
    first_name VARCHAR,
    last_name VARCHAR,
    display_name VARCHAR,
    avatar_url VARCHAR,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS client_profiles (
    id SERIAL PRIMARY KEY,
    profile_id INT NOT NULL REFERENCES profiles(id) ON DELETE CASCADE,
    company_name VARCHAR,
    tax_id VARCHAR,
    verification_status VARCHAR,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS freelancer_profiles (
    id SERIAL PRIMARY KEY,
    profile_id INT NOT NULL REFERENCES profiles(id) ON DELETE CASCADE,
    headline VARCHAR,
    bio TEXT,
    skills JSONB,
    rating DECIMAL,
    verification_status VARCHAR,
    created_at TIMESTAMP DEFAULT NOW()
);
