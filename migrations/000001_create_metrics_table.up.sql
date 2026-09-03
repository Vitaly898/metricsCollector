CREATE TABLE IF NOT EXISTS metrics (
    id    bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name  varchar(255) NOT NULL,
    type  varchar(10) NOT NULL,
    delta bigint,
    value double precision,
    UNIQUE (name, type)
);
