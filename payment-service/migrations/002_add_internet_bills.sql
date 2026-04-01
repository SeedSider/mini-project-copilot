CREATE TABLE IF NOT EXISTS internet_bill (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id      UUID NOT NULL,
    customer_id  VARCHAR(50) NOT NULL,
    name         VARCHAR(100) NOT NULL,
    address      TEXT NOT NULL,
    phone_number VARCHAR(50) NOT NULL,
    code         VARCHAR(50) NOT NULL,
    bill_from    VARCHAR(20) NOT NULL,
    bill_to      VARCHAR(20) NOT NULL,
    internet_fee VARCHAR(50) NOT NULL,
    tax          VARCHAR(50) NOT NULL,
    total        VARCHAR(50) NOT NULL
);
