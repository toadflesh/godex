-- +goose Up
CREATE TABLE transactions (
    txid VARCHAR(64),
    hash VARCHAR(64),
    segwit BOOLEAN,
    replace_by_fee BOOLEAN,
    version INTEGER,
    size BIGINT,
    vsize BIGINT,
    weight BIGINT,
    locktime BIGINT,
    fee DECIMAL,
    hex TEXT,
    blockhash VARCHAR(64),
    blockheight BIGINT,
    time TIMESTAMPTZ
);

-- +goose Down
DROP TABLE transactions;