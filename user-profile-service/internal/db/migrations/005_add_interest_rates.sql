CREATE TABLE IF NOT EXISTS interest_rate (
    id      VARCHAR(50)  PRIMARY KEY,
    kind    VARCHAR(20)  NOT NULL,
    deposit VARCHAR(10)  NOT NULL,
    rate    DECIMAL(5,2) NOT NULL
);
