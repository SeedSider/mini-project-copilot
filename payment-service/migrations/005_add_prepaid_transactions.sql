CREATE TABLE IF NOT EXISTS prepaid_transaction (
    id              VARCHAR(100) PRIMARY KEY,
    card_id         VARCHAR(100) NOT NULL,
    phone           VARCHAR(20) NOT NULL,
    amount          BIGINT NOT NULL,
    status          VARCHAR(20) NOT NULL DEFAULT 'SUCCESS',
    message         TEXT NOT NULL DEFAULT '',
    idempotency_key VARCHAR(100) UNIQUE NOT NULL,
    created_at      TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_prepaid_transaction_idempotency ON prepaid_transaction(idempotency_key);
