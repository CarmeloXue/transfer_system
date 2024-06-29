-- create_databases.sql
\c postgres

CREATE DATABASE account_db;

CREATE DATABASE transaction_db;


\c account_db

CREATE TABLE IF NOT EXISTS account_tab (
    id SERIAL PRIMARY KEY,
    account_id INT UNIQUE NOT NULL,
    balance BIGINT NOT NULL DEFAULT 0,
    out_balance BIGINT NOT NULL DEFAULT 0,
    in_balance BIGINT NOT NULL DEFAULT 0, 
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);


CREATE TABLE IF NOT EXISTS fund_movement_tab (
    id SERIAL PRIMARY KEY,
    transaction_id CHAR(36) UNIQUE NOT NULL,
    stage INT NOT NULL,
    source_account_id INT NOT NULL,
    destination_account_id INT NOT NULL,
    amount BIGINT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);


\c transaction_db

CREATE TABLE IF NOT EXISTS transaction_tab (
    id SERIAL PRIMARY KEY,
    transaction_id CHAR(36) UNIQUE NOT NULL,
    source_account_id INT NOT NULL,
    destination_account_id INT NOT NULL,
    amount BIGINT NOT NULL,
    transaction_status INT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    expired_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_transactions_status ON transaction_tab(transaction_status);
CREATE INDEX idx_transactions_expired_status ON transaction_tab(expired_at, transaction_status);