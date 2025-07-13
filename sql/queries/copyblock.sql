-- name: CopyBlock :one
BEGIN;

COPY blocks (
    blockhash, 
    confirmations, 
    height, 
    version, 
    version_hex, 
    merkle_root, 
    time, 
    median_time, 
    nonce,
    bits,
    difficulty,
    chainwork,
    ntx,
    previous_block_hash,
    next_block_hash,
    stripped_size,
    size,
    weight
)
FROM 'bulk/blocks.csv'
WITH (FORMAT csv, HEADER true, DELIMITER ',', NULL '');

COPY transactions (
    txid, 
    hash, 
    segwit, 
    replace_by_fee, 
    version, 
    size, 
    vsize, 
    weight,
    locktime,
    fee,
    hex,
    blockhash,
    blockheight,
    time
)
FROM 'bulk/transactions.csv'
WITH (FORMAT csv, HEADER true, DELIMITER ',', NULL '');

COPY vins (
    txid,
    prev_txid,
    coinbase,
    vout,
    scriptsig_asm,
    scriptsig_hex,
    txinwitness,
    prev_blockheight,
    value,
    script_pubkey_asm,
    script_pubkey_desc,
    script_pubkey_hex,
    script_pubkey_address,
    script_pubkey_type,
    sequence
)
FROM 'bulk/vins.csv'
WITH (FORMAT csv, HEADER true, DELIMITER ',', NULL '', QUOTE '"', ESCAPE '\');

COPY vouts (
    txid,
    value,
    n,
    script_pubkey_asm,
    script_pubkey_desc,
    script_pubkey_hex,
    script_pubkey_address,
    script_pubkey_type
)
FROM 'bulk/vouts.csv'
WITH (FORMAT csv, HEADER true, DELIMITER ',', NULL '', QUOTE '"', ESCAPE '\');
COMMIT;