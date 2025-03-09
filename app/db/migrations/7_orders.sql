-- +goose Up
-- Create orders table
CREATE TABLE IF NOT EXISTS orders (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    user_profile_id INTEGER NOT NULL,
    status TEXT NOT NULL DEFAULT 'pending',
    delivery_date DATETIME NOT NULL,
    note TEXT,
    total_price REAL NOT NULL DEFAULT 0,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
    deleted_at DATETIME,
    FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY(user_profile_id) REFERENCES user_profiles(id) ON DELETE CASCADE
);

-- Create order_items table
CREATE TABLE IF NOT EXISTS order_items (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    order_id INTEGER NOT NULL,
    meal_option_id INTEGER NOT NULL,
    quantity INTEGER NOT NULL DEFAULT 1,
    price REAL NOT NULL DEFAULT 0,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
    deleted_at DATETIME,
    FOREIGN KEY(order_id) REFERENCES orders(id) ON DELETE CASCADE,
    FOREIGN KEY(meal_option_id) REFERENCES meal_options(id) ON DELETE CASCADE
);

-- Create delivery_infos table
CREATE TABLE IF NOT EXISTS delivery_infos (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    order_id INTEGER NOT NULL UNIQUE,
    driver_id INTEGER,
    scheduled_time DATETIME NOT NULL,
    actual_time DATETIME,
    delivery_status TEXT NOT NULL DEFAULT 'scheduled',
    delivery_notes TEXT,
    delivery_address TEXT NOT NULL,
	latitude  REAL NOT NULL DEFAULT 0,
	longitude REAL NOT NULL DEFAULT 0,
	custom_address BOOLEAN DEFAULT false,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
    deleted_at DATETIME,
    FOREIGN KEY(order_id) REFERENCES orders(id) ON DELETE CASCADE,
    FOREIGN KEY(driver_id) REFERENCES users(id) ON DELETE SET NULL
);

-- Create index for faster lookups
CREATE INDEX idx_orders_user_id ON orders(user_id);
CREATE INDEX idx_orders_status ON orders(status);
CREATE INDEX idx_order_items_order_id ON order_items(order_id);
CREATE INDEX idx_order_items_meal_option_id ON order_items(meal_option_id);
CREATE INDEX idx_delivery_infos_order_id ON delivery_infos(order_id);
CREATE INDEX idx_delivery_infos_driver_id ON delivery_infos(driver_id);
CREATE INDEX idx_delivery_infos_status ON delivery_infos(delivery_status);

-- +goose Down
-- Drop tables in reverse order to avoid foreign key constraints
DROP TABLE IF EXISTS delivery_infos;
DROP TABLE IF EXISTS order_items;
DROP TABLE IF EXISTS orders;