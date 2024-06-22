USE `transfer_db`;

CREATE TABLE IF NOT EXISTS `account_tab` (
    id SERIAL PRIMARY KEY,
    account_id INT UNIQUE NOT NULL,
    balance DECIMAL(20, 8) NOT NULL DEFAULT 0,
    enabled int NOT NULL DEFAULT 1,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_account_id ON account_tab (account_id);


CREATE TABLE IF NOT EXISTS `order_tab` (
    order_id SERIAL PRIMARY KEY,
    source_account_id INT NOT NULL,
    destination_account_id INT NOT NULL,
    order_status INT NOT NULL DEFAULT 0,
    amount DECIMAL(20, 8) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    expired_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_source_status_updated ON order_tab (source_account_id, order_status, updated_at DESC);


CREATE TABLE IF NOT EXISTS `transaction_tab` (
    transaction_id SERIAL PRIMARY KEY,
    order_id INT NOT NULL,
    transaction_type INT NOT NULL,
    transaction_status INT NOT NULL,
    amount DECIMAL(20, 8) NOT NULL,
    owner_id INT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_transaction_order_id ON transaction_tab (order_id);
CREATE INDEX IF NOT EXISTS idx_transaction_owner_id ON transaction_tab (owner_id);

