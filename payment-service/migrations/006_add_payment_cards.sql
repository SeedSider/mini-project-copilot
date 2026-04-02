CREATE TABLE IF NOT EXISTS payment_card (
    id              VARCHAR(100) PRIMARY KEY DEFAULT 'card-' || gen_random_uuid()::text,
    account_id      VARCHAR(100) NOT NULL,
    holder_name     VARCHAR(200) NOT NULL,
    card_label      VARCHAR(100) NOT NULL DEFAULT '',
    masked_number   VARCHAR(30)  NOT NULL DEFAULT '',
    balance         BIGINT       NOT NULL DEFAULT 0,
    currency        VARCHAR(10)  NOT NULL DEFAULT 'USD',
    brand           VARCHAR(20)  NOT NULL DEFAULT 'VISA',
    gradient_colors TEXT[]       NOT NULL DEFAULT '{}'
);

CREATE INDEX IF NOT EXISTS idx_payment_card_account_id ON payment_card(account_id);
