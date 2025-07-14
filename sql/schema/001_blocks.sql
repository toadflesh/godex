-- +goose Up
CREATE TABLE blocks (
    blockhash VARCHAR(64),
    confirmations INTEGER,
    height INTEGER,
    version INTEGER,
    version_hex VARCHAR(64),
    merkle_root VARCHAR(64),
    time TIMESTAMPTZ,
    median_time TIMESTAMPTZ,
    nonce INTEGER,
    bits VARCHAR(64),
    difficulty DECIMAL(20,8),
    chainwork VARCHAR(64),
    ntx INTEGER,
    previous_block_hash VARCHAR(64),
    next_block_hash VARCHAR(64),
    stripped_size INTEGER,
    size INTEGER,
    weight INTEGER
);

-- +goose Down
DROP TABLE blocks;