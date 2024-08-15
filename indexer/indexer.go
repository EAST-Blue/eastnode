package indexer

import (
	"eastnode/indexer/repository/bitcoin"
	"eastnode/indexer/repository/db"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

type Indexer struct {
	DbRepo      *db.DBRepository
	bitcoinRepo bitcoin.BitcoinRepositoryInterface
}

func NewIndexer(dbRepo *db.DBRepository, bitcoinRepo bitcoin.BitcoinRepositoryInterface) *Indexer {
	return &Indexer{dbRepo, bitcoinRepo}
}

func (i *Indexer) flush(blockHeight int32, newBlocks *[]db.Block, newTxs *[]db.Transaction, newVins *[]db.Vin, newVouts *[]db.Vout) error {
	log.Printf("Flushing blocks to DB. Last block: %d", blockHeight)
	err := i.DbRepo.CreateBlocks(newBlocks)
	if err != nil {
		return err
	}

	err = i.DbRepo.CreateTransactions(newTxs)
	if err != nil {
		return err
	}

	err = i.DbRepo.CreateVins(newVins)
	if err != nil {
		return err
	}

	err = i.DbRepo.CreateVouts(newVouts)
	if err != nil {
		return err
	}

	// update outpoint spending
	err = i.DbRepo.SetLastHeight(blockHeight)
	if err != nil {
		return err
	}
	commitMessage := fmt.Sprintf("Indexed block %d", blockHeight)
	i.DbRepo.Db.Exec("CALL DOLT_COMMIT('--allow-empty', '-Am', ?);", commitMessage)

	return nil
}

func (i *Indexer) IndexBlocks(fromBlockHeight int32, toBlockHeight int32) error {
	log.Printf("index new blocks from %d to %d", fromBlockHeight, toBlockHeight)

	blockHeight := fromBlockHeight
	blockHash, err := i.bitcoinRepo.GetBlockHash(fromBlockHeight)
	if err != nil {
		return err
	}

	newBlocks := []db.Block{}
	newTxs := []db.Transaction{}
	newVins := []db.Vin{}
	newVouts := []db.Vout{}

	for {
		// Break if toBlockHeight is empty (0) or if we've reached it
		if blockHeight > toBlockHeight {
			break
		}

		block, err := i.bitcoinRepo.GetBlock(blockHash)
		if err != nil {
			return err
		}

		err = i.HandleBlock(blockHeight, block, &newBlocks, &newTxs, &newVins, &newVouts)
		if err != nil {
			return err
		}

		// Flush data if we've reached toBlockHeight
		if blockHeight == toBlockHeight {
			err = i.flush(blockHeight, &newBlocks, &newTxs, &newVins, &newVouts)
			if err != nil {
				return err
			}
			newBlocks = []db.Block{}
			newTxs = []db.Transaction{}
			newVins = []db.Vin{}
			newVouts = []db.Vout{}
		}

		i := 10

		if os.Getenv("INDEXER_SLEEP_TIME") != "" {
			i, err = strconv.Atoi(os.Getenv("INDEXER_SLEEP_TIME"))
			if err != nil {
				return err
			}
		}

		// Sleep because of RPC rate limit
		time.Sleep(time.Duration(i) * time.Millisecond)

		blockHeight++
		blockHash = block.Nextblockhash
		if blockHash == "" {
			break // End of chain reached
		}
	}

	return nil
}

func (i *Indexer) HandleBlock(blockHeight int32, block *bitcoin.GetBlock, newBlocks *[]db.Block, newTxs *[]db.Transaction, newVins *[]db.Vin, newVout *[]db.Vout) error {
	log.Printf("handle block height %d, hash %s", blockHeight, block.Hash)

	// insert block
	newBlock := db.Block{
		Hash:          block.Hash,
		Height:        uint64(blockHeight),
		PreviousBlock: block.Previousblockhash,
		Version:       int32(block.Version),
		Nonce:         uint32(block.Nonce),
		Timestamp:     uint32(block.Time),
		Bits:          block.Bits,
		MerkleRoot:    block.Merkleroot,
	}
	*newBlocks = append(*newBlocks, newBlock)

	// fill the txhash using txid instead of txhash, for the witness tx the id is different from the hash
	// https://bitcoin.stackexchange.com/questions/77699/whats-the-difference-between-txid-and-hash-getrawtransaction-bitcoind

	for txIdx, transaction := range block.Tx {
		// insert transaction
		newTx := db.Transaction{
			Hash:        transaction.Txid,
			LockTime:    uint32(transaction.Locktime),
			Version:     int32(transaction.Version),
			Safe:        false,
			BlockHash:   block.Hash,
			BlockHeight: uint64(blockHeight),
			BlockIndex:  uint32(txIdx),
		}
		*newTxs = append(*newTxs, newTx)

		// vouts
		for voutIdx, vout := range transaction.Vout {
			// convert btc value into sat, 1 btc is 100_000_000 sats
			satValue := int64(vout.Value * 100_000_000)
			vout := db.Vout{
				TxHash:       transaction.Txid,
				TxIndex:      uint32(voutIdx),
				BlockHash:    block.Hash,
				BlockHeight:  uint64(blockHeight),
				BlockTxIndex: uint32(txIdx),
				PkScript:     vout.ScriptPubKey.Hex,
				Value:        satValue,
				Spender:      vout.ScriptPubKey.Address,
			}
			*newVout = append(*newVout, vout)
		}

		// vins
		for idxx, vin := range transaction.Vin {
			satValue := int64(vin.PrevOutput.Value * 100_000_000)
			vin := db.Vin{
				TxHash:          transaction.Txid,
				TxIndex:         uint32(idxx),
				BlockHash:       block.Hash,
				BlockHeight:     uint64(blockHeight),
				BlockTxIndex:    uint32(txIdx),
				Sequence:        uint32(vin.Sequence),
				SignatureScript: vin.ScriptSig.Hex,

				PkScript: vin.PrevOutput.ScriptPubKey.Hex,
				Value:    satValue,
				Spender:  vin.PrevOutput.ScriptPubKey.Address,

				Witness: strings.Join(vin.Txinwitness, ","),
			}
			*newVins = append(*newVins, vin)
		}
	}

	return nil
}

func (i *Indexer) FindReorgHeight(fromHeight int32, depth int32) (int32, error) {
	for d := depth; d > 0; d-- {
		height := fromHeight - int32(d)

		if height <= 0 {
			break // Stop if we reach genesis block
		}

		// Get block hash from our database
		dbBlock, err := i.DbRepo.GetBlockByHeight(int64(height))
		if err != nil {
			return 0, err
		}

		// Get block hash from Bitcoin node
		btcBlockHash, err := i.bitcoinRepo.GetBlockHash(height)
		if err != nil {
			return 0, err
		}

		// Compare hashes
		if dbBlock.Hash != btcBlockHash {
			// Reorg detected, return the height where it occurred
			return height, nil
		}
	}

	// No reorg detected in the last 6 blocks
	return 0, nil
}

func (i *Indexer) Reorg(fromHeight int32, depth int32) (int32, error) {
	// Find the height where the reorg occurred
	reorgHeight, err := i.FindReorgHeight(fromHeight, depth)
	if err != nil {
		return 0, fmt.Errorf("failed to find reorg height: %w", err)
	}

	if reorgHeight == 0 {
		// No reorg detected
		return 0, nil
	}

	log.Printf("Reorg detected at height %d. Starting reorganization process.", reorgHeight)

	// Delete blocks from reorg height onwards
	err = i.DbRepo.DeleteBlocksFrom(reorgHeight)
	if err != nil {
		return 0, fmt.Errorf("failed to delete blocks from height %d: %w", reorgHeight, err)
	}
	err = i.DbRepo.SetLastHeight(reorgHeight - 1)
	if err != nil {
		return 0, fmt.Errorf("failed to delete blocks from height %d: %w", reorgHeight, err)
	}

	return reorgHeight, nil
}
