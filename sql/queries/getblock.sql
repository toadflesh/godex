-- name: GetBlock :one
SELECT * FROM blocks WHERE blockhash=$1;