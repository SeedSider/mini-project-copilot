CREATE TABLE IF NOT EXISTS exchange_rate (
    id           VARCHAR(50)   PRIMARY KEY,
    country      VARCHAR(100)  NOT NULL,
    currency     VARCHAR(10)   NOT NULL,
    country_code VARCHAR(10)   NOT NULL,
    buy          DECIMAL(15,4) NOT NULL,
    sell         DECIMAL(15,4) NOT NULL
);
