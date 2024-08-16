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

const MAX_BLOCK_FLUSH = 500

type Indexer struct {
	DbRepo      *db.DBRepository
	bitcoinRepo bitcoin.BitcoinRepositoryInterface
}

func NewIndexer(dbRepo *db.DBRepository, bitcoinRepo bitcoin.BitcoinRepositoryInterface) *Indexer {
	return &Indexer{dbRepo, bitcoinRepo}
}

func (i *Indexer) SyncBlocks(startHeight int32, endHeight int32) error {
	if endHeight-startHeight > MAX_BLOCK_FLUSH {
		// Many blocks to sync, log the sync process
		log.Printf("Syncing %d blocks from height %d to %d", endHeight-startHeight, startHeight, endHeight)

		for h := startHeight; h <= endHeight; h += MAX_BLOCK_FLUSH {
			endBlock := h + MAX_BLOCK_FLUSH
			if endBlock > endHeight {
				endBlock = endHeight
			}

			err := i.IndexBlocks(h, endBlock)
			if err != nil {
				return err
			}
			// Increment i after the block is indexed
			h++
		}
	} else {
		// Just a few blocks to add, sync one by one
		for h := startHeight; h <= endHeight; h++ {
			log.Printf("Indexing block %d", h)

			// Check for reorg before indexing each block
			reorgHeight, err := i.Reorg(h, REORG_DEPTH_CHECK)
			if err != nil {
				return err
			}

			if reorgHeight > 0 {
				h = reorgHeight
			}

			err = i.IndexBlocks(h, h)
			if err != nil {
				return err
			}
		}
	}
	return nil
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

				FundingTxHash:  vin.Txid,
				FundingTxIndex: uint32(vin.Vout),

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
	if fromHeight == 0 {
		return 0, nil
	}

	for currentHeight := fromHeight; currentHeight > 0 && currentHeight >= fromHeight-depth; currentHeight-- {
		// Get block from Bitcoin node
		btcBlockHash, err := i.bitcoinRepo.GetBlockHash(int32(currentHeight))
		if err != nil {
			return 0, fmt.Errorf("failed to get block from Bitcoin node at height %d: %w", currentHeight, err)
		}

		btcBlock, err := i.bitcoinRepo.GetBlockWithVerbosity(btcBlockHash, 2)
		if err != nil {
			return 0, fmt.Errorf("failed to get block from Bitcoin node at height %d: %w", currentHeight, err)
		}

		// Get previous block from our database
		dbBlock, err := i.DbRepo.GetBlockByHeight(int64(currentHeight - 1))
		if err != nil {
			return 0, fmt.Errorf("failed to get block from DB at height %d: %w", currentHeight, err)
		}

		if btcBlock.Previousblockhash == dbBlock.Hash {
			if currentHeight == fromHeight {
				return 0, nil // No reorg detected
			}
			return currentHeight, nil // Reorg starts at this height
		}
	}

	return 0, nil // No reorg detected within the specified depth
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
	err = i.DbRepo.UpdateBlocksAsOrphan(reorgHeight)
	if err != nil {
		return 0, fmt.Errorf("failed to delete blocks from height %d: %w", reorgHeight, err)
	}
	err = i.DbRepo.SetLastHeight(reorgHeight - 1)
	if err != nil {
		return 0, fmt.Errorf("failed to delete blocks from height %d: %w", reorgHeight, err)
	}

	return reorgHeight, nil
}
