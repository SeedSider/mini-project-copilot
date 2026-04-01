CREATE TABLE IF NOT EXISTS beneficiary (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    account_id VARCHAR(100) NOT NULL,
    name       VARCHAR(100) NOT NULL,
    phone      VARCHAR(20) NOT NULL,
    avatar     VARCHAR(500) DEFAULT ''
);

CREATE INDEX IF NOT EXISTS idx_beneficiary_account_id ON beneficiary(account_id);
