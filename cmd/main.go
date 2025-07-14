package main

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/shopspring/decimal"
	"github.com/toadflesh/godex/internal/database"
)

type Config struct {
	RPCURL  string
	RPCUser string
	RPCPass string
	DBURL   string
}

type RPCRequest struct {
	Version string `json:"version"`
	ID      string `json:"id"`
	Method  string `json:"method"`
	Params  []any  `json:"params"`
}

type BlockchainInfo struct {
	Chain                string  `json:"chain"`
	Blocks               int64   `json:"blocks"`
	Headers              int64   `json:"headers"`
	BestBlockHash        string  `json:"bestblockhash"`
	Difficulty           float64 `json:"difficulty"`
	MedianTime           int64   `json:"medianttime"`
	VerificationProgress float64 `json:"verificationprogress"`
	InitialBlockDownload bool    `json:"initialblockdownload"`
	Chainwork            string  `json:"chainwork"`
	SizeOnDisk           int64   `json:"size_on_disk"`
	Pruned               bool    `json:"pruned"`
	PruneHeight          int64   `json:"pruneheight"`
	AutomaticPruning     bool    `json:"automatic_pruning"`
	PruneTargetSize      int64   `json:"prune_target_size"`
	SoftForks            []struct {
		ForkName struct {
			Type string `json:"type"`
			Bip9 struct {
				Status     string `json:"status"`
				Bit        int64  `json:"bit"`
				StartTime  int64  `json:"start_time"`
				Timeout    int64  `json:"timeout"`
				Since      int64  `json:"since"`
				Statistics struct {
					Period    int64 `json:"period"`
					Threshold int64 `json:"threshold"`
					Elapsed   int64 `json:"elapsed"`
					Count     int64 `json:"count"`
					Possible  bool  `json:"possible"`
				} `json:"statistics"`
			} `json:"bip9"`
			Height int64 `json:"height"`
			Active bool  `json:"active"`
		} `json:"xxxx"`
	} `json:"softforks"`
	Warnings []string `json:"warnings"`
}

type Block struct {
	BlockHash         string          `json:"hash"`
	Confirmations     int64           `json:"confirmations"`
	Height            int64           `json:"height"`
	Version           int64           `json:"version"`
	VersionHex        string          `json:"versionHex"`
	MerkleRoot        string          `json:"merkleroot"`
	Time              int64           `json:"time"`
	MedianTime        int64           `json:"mediantime"`
	Nonce             int64           `json:"nonce"`
	Bits              string          `json:"bits"`
	Difficulty        decimal.Decimal `json:"difficulty"`
	Chainwork         string          `json:"chainwork"`
	NTX               int64           `json:"nTx"`
	PreviousBlockHash *string         `json:"previousblockhash"`
	NextBlockHash     *string         `json:"nextblockhash"`
	StrippedSize      int64           `json:"strippedsize"`
	Size              int64           `json:"size"`
	Weight            int64           `json:"weight"`
	Transactions      []Transaction   `json:"tx"`
}

type Transaction struct {
	TxID     string `json:"txid"`
	Hash     string `json:"hash"`
	Version  int64  `json:"version"`
	Size     int64  `json:"size"`
	VSize    int64  `json:"vsize"`
	Weight   int64  `json:"weight"`
	Locktime int64  `json:"locktime"`
	Vin      []struct {
		Coinbase  *string `json:"coinbase"`
		TxID      *string `json:"txid"`
		Vout      int64   `json:"vout"`
		ScriptSig struct {
			Asm *string `json:"asm"`
			Hex *string `json:"hex"`
		} `json:"scriptSig"`
		TxInWitness []string `json:"txinwitness"`
		Prevout     struct {
			Generated    bool             `json:"generated"`
			PrevHeight   *int64           `json:"height"`
			Value        *decimal.Decimal `json:"value"`
			ScriptPubKey struct {
				Asm     *string `json:"asm"`
				Desc    *string `json:"desc"`
				Hex     *string `json:"hex"`
				Address *string `json:"address"`
				Type    *string `json:"type"`
			} `json:"scriptPubKey"`
		} `json:"prevout"`
		Sequence int64 `json:"sequence"`
	} `json:"vin"`
	Vout []struct {
		Value        decimal.Decimal `json:"value"`
		N            int64           `json:"n"`
		ScriptPubKey struct {
			Asm     string  `json:"asm"`
			Desc    string  `json:"desc"`
			Hex     string  `json:"hex"`
			Address *string `json:"address"`
			Type    string  `json:"type"`
		} `json:"scriptPubKey"`
	} `json:"vout"`
	Fee decimal.Decimal `json:"fee"`
	Hex string          `json:"hex"`
}

func (rpcRequest *RPCRequest) SendRPCRequest(config *Config, payload []byte) (RPCResponse, error) {
	var rpcResponse RPCResponse
	request, err := http.NewRequest("POST", config.RPCURL, bytes.NewBuffer(payload))
	if err != nil {
		log.Fatalf("error creating a request for getblockchaininfo payload: %v", err)
	}
	request.SetBasicAuth(config.RPCUser, config.RPCPass)
	request.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		httpResponseError := fmt.Sprintf("there was a problem getting a %s response: %v", rpcRequest.Method, err)
		return rpcResponse, errors.New(httpResponseError)
	}

	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		log.Fatalf("there was a problem reading the %s response: %v", rpcRequest.Method, err)
	}

	err = json.Unmarshal(body, &rpcResponse)
	if err != nil {
		log.Fatalf("there was a problem unmarshaling the RPCResponse: %v", err)
	}

	if rpcResponse.Error != nil {
		rpcResponseError := fmt.Sprintf("there was a problem with the %s response: %v", rpcRequest.Method, rpcResponse.Error)
		return rpcResponse, errors.New(rpcResponseError)
	}

	return rpcResponse, nil
}

func (rpcRequest *RPCRequest) GetBlockchainInfo(config *Config, ctx *context.Context, db *database.Queries) (BlockchainInfo, error) {
	var blockchainInfo BlockchainInfo

	payload, err := json.Marshal(rpcRequest)
	if err != nil {
		log.Fatalf("error marshaling %s request, %v", rpcRequest.Method, err)
	}

	rpcResponse, err := rpcRequest.SendRPCRequest(config, payload)
	if err != nil {
		rpcResponseError := fmt.Sprintf("there was a problem with the %s RPC request: %v", rpcRequest.Method, err)
		return blockchainInfo, errors.New(rpcResponseError)
	}

	err = json.Unmarshal(rpcResponse.Result, &blockchainInfo)
	if err != nil {
		blockchainInfoError := fmt.Sprintf("there was a problem unmarshaling the blockchaininfo result: %v", err)
		return blockchainInfo, errors.New(blockchainInfoError)
	}

	return blockchainInfo, nil
}

func (rpcRequest *RPCRequest) GetBlock(config *Config, ctx *context.Context, db *database.Queries) (Block, error) {
	var block Block
	payload, err := json.Marshal(rpcRequest)
	if err != nil {
		log.Fatalf("error marshalling %s request, %v", rpcRequest.Method, err)
	}

	rpcResponse, err := rpcRequest.SendRPCRequest(config, payload)
	if err != nil {
		rpcResponseError := fmt.Sprintf("there was a problem with the %s RPC request: %v", rpcRequest.Method, err)
		return block, errors.New(rpcResponseError)
	}

	err = json.Unmarshal(rpcResponse.Result, &block)
	if err != nil {
		blockError := fmt.Sprintf("there was a problem unmarshaling the %s result: %v", rpcRequest.Method, err)
		return block, errors.New(blockError)
	}
	// fmt.Printf("coinbase: %s\n", block.Transactions[0].Vin[0].Coinbase)
	return block, nil
}

func (rpcRequest *RPCRequest) GetBlockHash(config *Config, ctx *context.Context, db *database.Queries) (string, error) {
	var blockhash string
	payload, err := json.Marshal(rpcRequest)
	if err != nil {
		log.Fatalf("error marshalling %s request, %v", rpcRequest.Method, err)
	}

	rpcResponse, err := rpcRequest.SendRPCRequest(config, payload)
	if err != nil {
		rpcResponseError := fmt.Sprintf("there was a problem with the %s RPC request: %v", rpcRequest.Method, err)
		return blockhash, errors.New(rpcResponseError)
	}

	err = json.Unmarshal(rpcResponse.Result, &blockhash)
	if err != nil {
		blockhashError := fmt.Sprintf("there was a problem unmarshaling the %s result: %v", rpcRequest.Method, err)
		return blockhash, errors.New(blockhashError)
	}

	return blockhash, nil
}

func (block *Block) CopyBlock(ctx *context.Context, conn *pgx.Conn, highestBlock int64) error {
	//********************************************************************//
	// Start a transaction                                                //
	//********************************************************************//
	dbtx, err := conn.Begin(*ctx)
	if err != nil {
		dbtxErr := fmt.Sprintf("ERROR: error beginning pgx transaction: %v", err)
		return errors.New(dbtxErr)
	}

	// If there is an error, rollback changes
	defer dbtx.Rollback(*ctx)

	//********************************************************************//
	// COPY blocks                                                        //
	//********************************************************************//

	// convert timestamps to time.Time
	blockTime := time.Unix(int64(block.Time), 0).UTC()
	blockMedianTime := time.Unix(int64(block.MedianTime), 0).UTC()

	// Generate csv
	blocksCSV := fmt.Sprintf(
		"%s,%d,%d,%d,%s,%s,%s,%s,%d,%s,%s,%s,%d,%s,%s,%d,%d,%d\n",
		block.BlockHash,
		block.Confirmations,
		block.Height,
		block.Version,
		block.VersionHex,
		block.MerkleRoot,
		blockTime.Format("2006-01-02 15:04:05-07:00"),
		blockMedianTime.Format("2006-01-02 15:04:05-07:00"),
		block.Nonce,
		block.Bits,
		block.Difficulty,
		block.Chainwork,
		block.NTX,
		NullString(block.PreviousBlockHash),
		NullString(block.NextBlockHash),
		block.StrippedSize,
		block.Size,
		block.Weight,
	)
	// log.Printf("%d,%d,%d,%d,%d,%d,%d,%d\n", block.Confirmations, block.Height, block.Version, block.Nonce, block.NTX, block.StrippedSize, block.Size, block.Weight)
	_, err = dbtx.Conn().PgConn().CopyFrom(*ctx, strings.NewReader(blocksCSV), `
		COPY blocks (
		blockhash, confirmations, height, version, version_hex, merkle_root, 
		time, median_time, nonce, bits, difficulty, chainwork, ntx, 
		previous_block_hash, next_block_hash, stripped_size, size, weight
		) FROM STDIN WITH (FORMAT csv, NULL '')
	`)
	if err != nil {
		copyBlocksErr := fmt.Sprintf("ERROR: error with COPY blocks: %v", err)
		return errors.New(copyBlocksErr)
	}

	//********************************************************************//
	// COPY transactions                                                  //
	//********************************************************************//

	transactionsCSV := strings.Builder{}
	for _, t := range block.Transactions {
		segwit := false

		// Determine if segwit
		if t.TxID != t.Hash {
			segwit = true
		}

		// Determine if RBF is enabled
		// Check each vin sequence
		// - if sequence < 0xfffffffe
		// then RBF is enabled
		rbf := false
		for _, in := range t.Vin {
			if in.Sequence < 0xfffffffe {
				rbf = true
				break
			}
		}
		fmt.Fprintf(&transactionsCSV, "%s,%s,%t,%t,%d,%d,%d,%d,%d,%s,%s,%s,%d,%s\n",
			t.TxID,
			t.Hash,
			segwit,
			rbf,
			t.Version,
			t.Size,
			t.VSize,
			t.Weight,
			t.Locktime,
			t.Fee.String(),
			t.Hex,
			block.BlockHash,
			block.Height,
			blockTime.Format("2006-01-02 15:04:05-07:00"),
		)

		//********************************************************************//
		// COPY vins                                                          //
		//********************************************************************//
		vinsCSV := strings.Builder{}
		for _, v := range t.Vin {
			witness := "{" + strings.Join(v.TxInWitness, ",") + "}"
			fmt.Fprintf(&vinsCSV, "%s,%s,%s,%d,%s,%s,%q,%s,%s,%s,%s,%s,%s,%s,%d\n",
				t.TxID,
				NullString(v.TxID),
				NullString(v.Coinbase),
				v.Vout,
				NullString(v.ScriptSig.Asm),
				NullString(v.ScriptSig.Hex),
				witness,
				NullInt64(v.Prevout.PrevHeight),
				NullDecimal(v.Prevout.Value),
				NullString(v.Prevout.ScriptPubKey.Asm),
				NullString(v.Prevout.ScriptPubKey.Desc),
				NullString(v.Prevout.ScriptPubKey.Hex),
				NullString(v.Prevout.ScriptPubKey.Address),
				NullString(v.Prevout.ScriptPubKey.Type),
				v.Sequence,
			)
		}
		_, err = dbtx.Conn().PgConn().CopyFrom(*ctx, strings.NewReader(vinsCSV.String()), `
			COPY vins (
				txid, prev_txid, coinbase, vout, scriptsig_asm, scriptsig_hex, txinwitness,
				prev_blockheight, value, script_pubkey_asm, script_pubkey_desc,
				script_pubkey_hex, script_pubkey_address, script_pubkey_type, sequence
			) FROM STDIN WITH (FORMAT csv, NULL '', QUOTE '"', ESCAPE '\')
		`)
		if err != nil {
			vinsCSVError := fmt.Sprintf("ERROR: error COPY vins to db: %v", err)
			return errors.New(vinsCSVError)
		}

		//********************************************************************//
		// COPY vins                                                          //
		//********************************************************************//
		voutsCSV := strings.Builder{}
		for _, v := range t.Vout {
			fmt.Fprintf(&voutsCSV, "%s,%s,%d,%s,%s,%s,%s,%s\n",
				t.TxID,
				v.Value.String(),
				v.N,
				v.ScriptPubKey.Asm,
				v.ScriptPubKey.Desc,
				v.ScriptPubKey.Hex,
				NullString(v.ScriptPubKey.Address),
				v.ScriptPubKey.Type,
			)
		}
		_, err = dbtx.Conn().PgConn().CopyFrom(*ctx, strings.NewReader(voutsCSV.String()), `
			COPY vouts (
				txid, value, n, script_pubkey_asm, script_pubkey_desc,
				script_pubkey_hex, script_pubkey_address, script_pubkey_type
			) FROM STDIN WITH (FORMAT csv, NULL '', QUOTE '"', ESCAPE '\')
		`)
		if err != nil {
			voutsCSVError := fmt.Sprintf("ERROR: error COPY vouts to db: %v", err)
			return errors.New(voutsCSVError)
		}
	}
	_, err = dbtx.Conn().PgConn().CopyFrom(*ctx, strings.NewReader(transactionsCSV.String()), `
	COPY transactions (
		txid, hash, segwit, replace_by_fee, version, size, vsize, weight,
		locktime, fee, hex, blockhash, blockheight, time
		) FROM STDIN WITH (FORMAT csv, NULL '')
	`)
	if err != nil {
		transactionCSVError := fmt.Sprintf("ERROR: error copying transactions: %v", err)
		return errors.New(transactionCSVError)
	}

	// Commit transaction
	// if no errors occurred we are ready to commit to db
	err = dbtx.Commit(*ctx)
	if err != nil {
		commitError := fmt.Sprintf("ERROR: error committing changes to db: %v", err)
		return errors.New(commitError)
	}

	percentComplete := float64(block.Height) / float64(highestBlock) * 100
	log.Printf("Block %d written to db { blocktime: %s } // { %6d / %6d } // -> %.2f%% Complete\n", block.Height, blockTime, block.Height, highestBlock, percentComplete)

	return nil
}

func NullString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func NullInt64(n *int64) string {
	if n == nil {
		return ""
	}
	return fmt.Sprintf("%d", *n)
}

func NullDecimal(d *decimal.Decimal) string {
	if d == nil {
		return ""
	}
	return d.String()
}

type RPCResponse struct {
	Result json.RawMessage `json:"result"`
	Error  any             `json:"error"`
	ID     any             `json:"id"`
}

func main() {
	//********************************************************************//
	// Load env variables using godotenv.Load()                           //
	//********************************************************************//
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("there was an error loading .env file: %v", err)
	}

	//********************************************************************//
	// Add env variables to the Config struct                             //
	//********************************************************************//
	config := Config{
		RPCURL:  os.Getenv("BITCOIN_RPC_URL"),
		RPCUser: os.Getenv("BITCOIN_RPC_USERNAME"),
		RPCPass: os.Getenv("BITCOIN_RPC_PASSWORD"),
		DBURL:   os.Getenv("POSTGRESQL_CONNECTION_STRING"),
	}

	//********************************************************************//
	// open connection to postgresql database                             //
	//********************************************************************//
	db, err := sql.Open("postgres", config.DBURL)
	if err != nil {
		log.Fatalf("error connecting to postgresql database: %v", err)
	}
	dbQueries := database.New(db)

	// create context
	ctx := context.Background()

	//********************************************************************//
	// pgx database connection for COPY                                   //
	//********************************************************************//
	conn, err := pgx.Connect(ctx, config.DBURL)
	if err != nil {
		log.Fatalf("error connecting to postgresql database using pgx: %v", err)
	}
	defer conn.Close(ctx)

	//********************************************************************//
	// Once the connection has been made, query the highest block         //
	// that has been inserted into the database                           //
	//********************************************************************//
	startingIndex, err := dbQueries.GetHighestBlock(ctx)
	if err != nil {
		if err == sql.ErrNoRows {
			startingIndex.Int32 = 0 // If there are no rows the starting block is 0
		}
	}

	//********************************************************************//
	// Infinite for loop to continue processing blocks until the current  //
	// block height is equal to the highest block in the database         //
	// - once this occurs we can move to phase two of the project         //
	// - which will include checking for new blocks and adding them to    //
	// - the database once they have been confirmed                       //
	//********************************************************************//
	for {
		getBlockChainInfoRPCRequest := RPCRequest{
			Version: "1.0",
			ID:      "1",
			Method:  "getblockchaininfo",
			Params:  []any{},
		}
		blockchainInfo, err := getBlockChainInfoRPCRequest.GetBlockchainInfo(&config, &ctx, dbQueries)
		if err != nil {
			log.Fatalf("%v", err)
		}
		currentHeight := blockchainInfo.Blocks
		var targetBlock int64
		highestBlockHeight, err := dbQueries.GetHighestBlock(ctx)
		if err != nil {
			if err == sql.ErrNoRows {
				targetBlock = 0
			}
		}

		//********************************************************************//
		// if the current height on the network does not equal the block      //
		// height of the database, get the next block after the blockheight   //
		// in the database.                                                   //
		//********************************************************************//
		if currentHeight != int64(highestBlockHeight.Int32) {
			targetBlock = int64(highestBlockHeight.Int32) + 1

			// RPC headers for getblockhash - targetBlock as params target
			// this will return the blockhash to query getblock with
			getBlockHashRPCRequest := RPCRequest{
				Version: "1.0",
				ID:      "1",
				Method:  "getblockhash",
				Params:  []any{targetBlock},
			}
			targetBlockhash, err := getBlockHashRPCRequest.GetBlockHash(&config, &ctx, dbQueries)
			if err != nil {
				log.Fatalf("%v", err)
			}

			// RPC headers for getblock - we will use verbosity 3 to include previous vin data
			getBlockRPCRequest := RPCRequest{
				Version: "1.0",
				ID:      "1",
				Method:  "getblock",
				Params:  []any{targetBlockhash, 3},
			}

			// Sends the request and returns a block struct to process further
			block, err := getBlockRPCRequest.GetBlock(&config, &ctx, dbQueries)
			if err != nil {
				log.Fatalf("%v", err)
			}

			// Copy the whole block to the database using COPY statements
			err = block.CopyBlock(&ctx, conn, currentHeight)
			if err != nil {
				log.Fatalf("%v", err)
			}
		} else {
			fmt.Printf("database is up to date \ncurrent block height=%d\nhighest block in the database=%d\n", currentHeight, highestBlockHeight.Int32)
		}
		// Increment index by one to continue until up-to-date
		startingIndex.Int32++
	}
}
