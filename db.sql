CREATE TABLE IF NOT EXISTS users (
    id serial PRIMARY KEY,
    name varchar(255) NOT NULL
);

CREATE TABLE IF NOT EXISTS balances (
    id serial PRIMARY KEY,
    total DECIMAL(9,2),
    user_id INTEGER REFERENCES users (id)
);

CREATE TABLE IF NOT EXISTS history (
    id serial PRIMARY KEY,
    amount DECIMAL(9,2),
    comment varchar(255) NOT NULL,
    date timestamp NOT NULL,
    balance_id INTEGER REFERENCES balances (id)
);

INSERT INTO users (
    name
)
VALUES
('Иванов Иван'),
('Семенов Семен'),
('Петров Петр');

INSERT INTO balances (
    user_id,
    total
)
VALUES
(1, 0.00),
(2, 5000.00),
(3, 100000.00);