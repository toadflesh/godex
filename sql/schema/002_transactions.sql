-- +goose Up
CREATE TABLE transactions (
    txid VARCHAR(64),
    hash VARCHAR(64),
    segwit BOOLEAN,
    replace_by_fee BOOLEAN,
    version INTEGER,
    size INTEGER,
    vsize INTEGER,
    weight INTEGER,
    locktime INTEGER,
    fee DECIMAL,
    hex TEXT,
    blockhash VARCHAR(64),
    blockheight INTEGER,
    time TIMESTAMPTZ
);

-- +goose Down
DROP TABLE transactions;