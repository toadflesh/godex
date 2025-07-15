-- +goose Up
CREATE TABLE vins (
    txid VARCHAR(64),
    prev_txid VARCHAR(64),
    coinbase TEXT,
    vout INTEGER,
    scriptsig_asm TEXT,
    scriptsig_hex TEXT,
    txinwitness TEXT[],
    prev_blockheight BIGINT,
    value DECIMAL(16,8),
    script_pubkey_asm TEXT,
    script_pubkey_desc TEXT,
    script_pubkey_hex TEXT,
    script_pubkey_address VARCHAR(64),
    script_pubkey_type VARCHAR(64),
    sequence BIGINT
);

-- +goose Down
DROP TABLE vins;