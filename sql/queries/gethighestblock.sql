-- name: GetHighestBlock :one
SELECT height FROM blocks ORDER BY height DESC;