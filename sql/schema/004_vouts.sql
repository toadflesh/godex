-- +goose Up
CREATE TABLE vouts (
    txid VARCHAR(64),
    value DECIMAL(16,8),
    n INTEGER,
    script_pubkey_asm TEXT,
    script_pubkey_desc TEXT,
    script_pubkey_hex TEXT,
    script_pubkey_address VARCHAR(64),
    script_pubkey_type VARCHAR(64)
);

-- +goose Down
DROP TABLE vouts;