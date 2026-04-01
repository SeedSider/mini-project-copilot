CREATE TABLE IF NOT EXISTS currency (
    id    UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code  VARCHAR(10) NOT NULL,
    label VARCHAR(100) NOT NULL,
    rate  NUMERIC(15,4) NOT NULL
);
