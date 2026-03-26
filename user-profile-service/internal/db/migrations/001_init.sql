CREATE TABLE IF NOT EXISTS profile (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID UNIQUE,
    bank            VARCHAR(50)  NOT NULL,
    branch          VARCHAR(50)  NOT NULL,
    name            VARCHAR(100) NOT NULL,
    card_number     VARCHAR(20)  NOT NULL,
    card_provider   VARCHAR(50)  NOT NULL,
    balance         BIGINT       NOT NULL DEFAULT 0,
    currency        VARCHAR(3)   NOT NULL DEFAULT 'IDR',
    account_type    VARCHAR(20)  NOT NULL DEFAULT 'REGULAR',
    image           TEXT         NOT NULL DEFAULT ''
);

CREATE TABLE IF NOT EXISTS menu (
    id        VARCHAR(50)  PRIMARY KEY,
    "index"   INTEGER      UNIQUE NOT NULL,
    type      VARCHAR(20)  NOT NULL,
    title     VARCHAR(50)  NOT NULL,
    icon_url  TEXT         NOT NULL,
    is_active BOOLEAN      NOT NULL DEFAULT TRUE
);
