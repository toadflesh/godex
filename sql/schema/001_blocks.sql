-- +goose Up
CREATE TABLE blocks (
    blockhash VARCHAR(64),
    confirmations BIGINT,
    height BIGINT,
    version INTEGER,
    version_hex VARCHAR(64),
    merkle_root VARCHAR(64),
    time TIMESTAMPTZ,
    median_time TIMESTAMPTZ,
    nonce BIGINT,
    bits VARCHAR(64),
    difficulty DECIMAL(20,8),
    chainwork VARCHAR(64),
    ntx INTEGER,
    previous_block_hash VARCHAR(64),
    next_block_hash VARCHAR(64),
    stripped_size BIGINT,
    size BIGINT,
    weight BIGINT
);

-- +goose Down
DROP TABLE blocks;