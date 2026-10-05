CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    role VARCHAR(50) NOT NULL DEFAULT 'client',
    is_confirmed BOOLEAN NOT NULL DEFAULT FALSE,
    confirmation_code VARCHAR(10),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS products (
    id VARCHAR(20) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    category VARCHAR(100) NOT NULL,
    price_ht NUMERIC(10, 2) NOT NULL,
    price_ttc NUMERIC(10, 2) NOT NULL,
    stock INT NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS carts (
    id VARCHAR(20) PRIMARY KEY,
    user_id INT REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS cart_items (
    id SERIAL PRIMARY KEY,
    cart_id VARCHAR(20) REFERENCES carts(id) ON DELETE CASCADE,
    product_id VARCHAR(20) REFERENCES products(id) ON DELETE CASCADE,
    quantity INT NOT NULL DEFAULT 1,
    UNIQUE(cart_id, product_id)
);

CREATE TABLE IF NOT EXISTS orders (
    id VARCHAR(20) PRIMARY KEY,
    user_id INT REFERENCES users(id) ON DELETE CASCADE,
    status VARCHAR(50) NOT NULL DEFAULT 'en attente',
    cancel_reason TEXT DEFAULT '',
    total_ttc NUMERIC(10, 2) NOT NULL DEFAULT 0.00,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS order_items (
    id SERIAL PRIMARY KEY,
    order_id VARCHAR(20) REFERENCES orders(id) ON DELETE CASCADE,
    product_id VARCHAR(20) REFERENCES products(id) ON DELETE CASCADE,
    quantity INT NOT NULL,
    price_unit_ttc NUMERIC(10, 2) NOT NULL
);

INSERT INTO users (email, password_hash, role, is_confirmed) 
VALUES ('admin@ecommerce.com', '$2a$10$wN3Gvg0R5uR9tJ7L1G0N7.3x1M/vWqF8H.g/7S.cW.r.xJ2N7lS2K', 'admin', TRUE)
ON CONFLICT (email) DO NOTHING;

INSERT INTO products (id, name, description, category, price_ht, price_ttc, stock) VALUES
('PDT-7D2K8N', 'Clavier Mécanique RGB', 'Clavier switchs blue réactif', 'Périphériques', 83.33, 100.00, 25),
('PDT-8X9L2M', 'Souris Gaming Sans Fil', 'Capteur optique 16000 DPI', 'Périphériques', 41.67, 50.00, 40),
('PDT-3P4Q5R', 'Écran 27" 144Hz IPS', 'Résolution QHD 2560x1440', 'Écrans', 208.33, 250.00, 15)
ON CONFLICT (id) DO NOTHING;
