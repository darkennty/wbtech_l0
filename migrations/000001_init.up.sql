CREATE TABLE IF NOT EXISTS "order" (
    order_uid UUID PRIMARY KEY,
    track_number VARCHAR(14) NOT NULL,
    entry VARCHAR(14) NOT NULL,
    locale VARCHAR(2) NOT NULL,
    internal_signature VARCHAR(20),
    customer_id VARCHAR(20) NOT NULL,
    delivery_service VARCHAR(20) NOT NULL,
    shardkey VARCHAR(8) NOT NULL,
    sm_id INTEGER NOT NULL,
    date_created TIMESTAMP NOT NULL,
    oof_shard VARCHAR(8)
);

CREATE TABLE IF NOT EXISTS delivery (
    order_uid UUID PRIMARY KEY REFERENCES "order" (order_uid) ON DELETE CASCADE,
    "name" VARCHAR(40) NOT NULL,
    phone VARCHAR(12) NOT NULL,
    zip VARCHAR(7) NOT NULL,
    city VARCHAR(32) NOT NULL,
    address VARCHAR(64) NOT NULL,
    region VARCHAR(32) NOT NULL,
    email VARCHAR(256) NOT NULL
);

CREATE TABLE IF NOT EXISTS payment (
    order_uid UUID PRIMARY KEY REFERENCES "order" (order_uid) ON DELETE CASCADE,
    "transaction" VARCHAR(20) NOT NULL,
    request_id VARCHAR(10),
    currency VARCHAR(3) NOT NULL,
    provider VARCHAR(20) NOT NULL,
    amount INTEGER NOT NULL,
    payment_dt NUMERIC(10,0) NOT NULL,
    bank VARCHAR(20) NOT NULL,
    delivery_cost INTEGER NOT NULL,
    goods_total INTEGER NOT NULL,
    custom_fee INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS item (
    order_uid UUID NOT NULL REFERENCES "order" (order_uid) ON DELETE CASCADE,
    chrt_id INTEGER NOT NULL,
    track_number VARCHAR(14) NOT NULL,
    price INTEGER NOT NULL,
    rid VARCHAR(30) NOT NULL,
    "name" VARCHAR(60) NOT NULL,
    sale INTEGER NOT NULL,
    "size" VARCHAR(8) NOT NULL,
    total_price INTEGER NOT NULL,
    nm_id INTEGER NOT NULL,
    brand VARCHAR(20) NOT NULL,
    status INTEGER NOT NULL,
    PRIMARY KEY (order_uid, chrt_id)
);