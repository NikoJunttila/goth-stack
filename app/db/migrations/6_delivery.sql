-- +goose Up
-- Create meal_centers table
CREATE TABLE IF NOT EXISTS meal_centers (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    address TEXT,
    latitude REAL,
    longitude REAL,
    phone_number TEXT,
    is_active BOOLEAN DEFAULT true,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
    deleted_at DATETIME
);

-- Create days_meals table
CREATE TABLE IF NOT EXISTS days_meals (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    meal_center_id INTEGER NOT NULL,
    name TEXT NOT NULL,
    description TEXT,
    meal_date DATETIME NOT NULL,
    is_active BOOLEAN DEFAULT true,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
    deleted_at DATETIME,
    FOREIGN KEY(meal_center_id) REFERENCES meal_centers(id) ON DELETE CASCADE
);

-- Create dietary_restrictions table
CREATE TABLE IF NOT EXISTS dietary_restrictions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    description TEXT,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
    deleted_at DATETIME
);

-- Create meal_options table
CREATE TABLE IF NOT EXISTS meal_options (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    days_meals_id INTEGER NOT NULL,
    name TEXT NOT NULL,
    description TEXT,
    price REAL NOT NULL DEFAULT 0,
    image TEXT,
    nutritional_info TEXT,
    is_available BOOLEAN DEFAULT true,
    max_daily_quantity INTEGER,
    current_daily_quantity INTEGER DEFAULT 0,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
    deleted_at DATETIME,
    FOREIGN KEY(days_meals_id) REFERENCES days_meals(id) ON DELETE CASCADE
);

-- Create user_profiles table
CREATE TABLE IF NOT EXISTS user_profiles (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL UNIQUE,
    address TEXT,
    latitude REAL,
    longitude REAL,
    phone_number TEXT,
    delivery_notes TEXT,
    dietary_notes TEXT,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
    deleted_at DATETIME,
    FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Create many-to-many relationship tables
CREATE TABLE IF NOT EXISTS meal_dietary_restrictions (
    meal_option_id INTEGER NOT NULL,
    dietary_restriction_id INTEGER NOT NULL,
    created_at DATETIME NOT NULL,
    PRIMARY KEY(meal_option_id, dietary_restriction_id),
    FOREIGN KEY(meal_option_id) REFERENCES meal_options(id) ON DELETE CASCADE,
    FOREIGN KEY(dietary_restriction_id) REFERENCES dietary_restrictions(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS user_dietary_restrictions (
    user_profile_id INTEGER NOT NULL,
    dietary_restriction_id INTEGER NOT NULL,
    created_at DATETIME NOT NULL,
    PRIMARY KEY(user_profile_id, dietary_restriction_id),
    FOREIGN KEY(user_profile_id) REFERENCES user_profiles(id) ON DELETE CASCADE,
    FOREIGN KEY(dietary_restriction_id) REFERENCES dietary_restrictions(id) ON DELETE CASCADE
);

-- +goose Down
-- Drop tables in reverse order to avoid foreign key constraints
DROP TABLE IF EXISTS user_dietary_restrictions;
DROP TABLE IF EXISTS meal_dietary_restrictions;
DROP TABLE IF EXISTS user_profiles;
DROP TABLE IF EXISTS meal_options;
DROP TABLE IF EXISTS dietary_restrictions;
DROP TABLE IF EXISTS days_meals;
DROP TABLE IF EXISTS meal_centers;