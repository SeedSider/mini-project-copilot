CREATE TABLE IF NOT EXISTS branch (
    id        VARCHAR(50)   PRIMARY KEY,
    name      VARCHAR(200)  NOT NULL,
    distance  VARCHAR(20)   NOT NULL,
    latitude  DECIMAL(10,6) NOT NULL,
    longitude DECIMAL(10,6) NOT NULL
);
