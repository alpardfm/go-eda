CREATE TABLE IF NOT EXISTS orders (
    id         TEXT PRIMARY KEY,
    customer   TEXT NOT NULL,
    amount     NUMERIC(12, 2) NOT NULL,
    status     TEXT NOT NULL DEFAULT 'pending',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_orders_status ON orders(status);
CREATE INDEX idx_orders_created_at ON orders(created_at DESC);
