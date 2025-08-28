CREATE TABLE IF NOT EXISTS orders (
    order_uid TEXT PRIMARY KEY,
    track_number TEXT NOT NULL,
    entry TEXT NOT NULL,
    locale TEXT,
    internal_signature TEXT,
    customer_id TEXT NOT NULL,
    delivery_service TEXT,
    shardkey TEXT,
    sm_id INT,
    date_created TIMESTAMPTZ NOT NULL,
    oof_shard TEXT
);

CREATE TABLE IF NOT EXISTS deliveries (
    order_uid TEXT PRIMARY KEY REFERENCES orders(order_uid) ON DELETE CASCADE,
    name TEXT NOT NULL,
    phone TEXT NOT NULL,
    zip TEXT,
    city TEXT NOT NULL,
    address TEXT NOT NULL,
    region TEXT,
    email TEXT
    );

CREATE TABLE IF NOT EXISTS payments (
    order_uid TEXT PRIMARY KEY REFERENCES orders(order_uid) ON DELETE CASCADE,
    transaction TEXT NOT NULL,
    request_id TEXT,
    currency TEXT NOT NULL,
    provider TEXT NOT NULL,
    amount INT NOT NULL,
    payment_dt BIGINT NOT NULL,
    bank TEXT,
    delivery_cost INT,
    goods_total INT,
    custom_fee INT
    );

CREATE TABLE IF NOT EXISTS items (
    id BIGSERIAL PRIMARY KEY,
    order_uid TEXT NOT NULL REFERENCES orders(order_uid) ON DELETE CASCADE,
    chrt_id INT NOT NULL,
    track_number TEXT NOT NULL,
    price INT NOT NULL,
    rid TEXT NOT NULL,
    name TEXT NOT NULL,
    sale INT NOT NULL,
    size TEXT,
    total_price INT NOT NULL,
    nm_id INT,
    brand TEXT,
    status INT,
    CONSTRAINT uq_item_per_order UNIQUE(order_uid, rid)
    );

CREATE INDEX IF NOT EXISTS idx_items_order_uid ON items(order_uid);
CREATE INDEX IF NOT EXISTS idx_orders_date_created ON orders(date_created DESC);

