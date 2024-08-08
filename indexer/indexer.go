package indexer

import (
	"eastnode/indexer/repository/bitcoin"
	"eastnode/indexer/repository/db"
	"errors"
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
	bitcoinRepo *bitcoin.BitcoinRepository
}

func NewIndexer(dbRepo *db.DBRepository, bitcoinRepo *bitcoin.BitcoinRepository) *Indexer {
	return &Indexer{dbRepo, bitcoinRepo}
}

func (i *Indexer) flush(blockHeight int32, newBlocks *[]db.Block, newTxs *[]db.Transaction, newOutpoints *[]db.OutPoint) error {
	log.Printf("Flushing blocks to DB. Last block: %d", blockHeight)
	err := i.DbRepo.CreateBlocks(newBlocks)
	if err != nil {
		return err
	}

	err = i.DbRepo.CreateTransactions(newTxs)
	if err != nil {
		return err
	}

	err = i.DbRepo.CreateOutpoints(newOutpoints)
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
	newOutpoints := []db.OutPoint{}

	for {
		// break if current block-height is the latest, no need to index next block
		if !(blockHash != "" && (blockHeight <= toBlockHeight)) {
			break
		}

		block, err := i.bitcoinRepo.GetBlock(blockHash)
		if err != nil {
			return err
		}

		err = i.HandleBlock(blockHeight, block, &newBlocks, &newTxs, &newOutpoints)
		if err != nil {
			return err
		}

		if toBlockHeight-blockHeight <= 1 || len(newBlocks) >= MAX_BLOCK_FLUSH {
			err = i.flush(blockHeight, &newBlocks, &newTxs, &newOutpoints)
			newBlocks = []db.Block{}
			newTxs = []db.Transaction{}
			newOutpoints = []db.OutPoint{}

			if err != nil {
				return err
			}
		}

		// Sleep because of RPC rate limit
		i, err := strconv.Atoi(os.Getenv("INDEXER_SLEEP_TIME"))
		if err != nil {
			return err
		}

		time.Sleep(time.Duration(i) * time.Millisecond)

		blockHeight++
		blockHash = block.Nextblockhash
	}

	return nil
}

func (i *Indexer) HandleBlock(blockHeight int32, block *bitcoin.GetBlock, newBlocks *[]db.Block, newTxs *[]db.Transaction, newOutpoints *[]db.OutPoint) error {
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

		// insert outpoints

		// vouts
		for voutIdx, vout := range transaction.Vout {
			// convert btc value into sat, 1 btc is 100_000_000 sats
			satValue := int64(vout.Value * 100_000_000)
			outpoint := db.OutPoint{
				FundingTxHash:       transaction.Txid,
				FundingTxIndex:      uint32(voutIdx),
				FundingBlockHash:    block.Hash,
				FundingBlockHeight:  uint64(blockHeight),
				FundingBlockTxIndex: uint32(txIdx),
				PkScript:            vout.ScriptPubKey.Hex,
				Value:               satValue,
				Spender:             vout.ScriptPubKey.Address,
			}
			*newOutpoints = append(*newOutpoints, outpoint)
		}

		// vins
		for idxx, vin := range transaction.Vin {
			// coinbase
			if vin.Coinbase != "" {
				outpoint := db.OutPoint{
					SpendingTxHash:       transaction.Txid,
					SpendingTxIndex:      uint32(idxx),
					SpendingBlockHash:    block.Hash,
					SpendingBlockHeight:  uint64(blockHeight),
					SpendingBlockTxIndex: uint32(txIdx),
					Sequence:             uint32(vin.Sequence),
					SignatureScript:      vin.ScriptSig.Hex,
					Witness:              strings.Join(vin.Txinwitness, ","),

					FundingTxHash:  vin.Txid,
					FundingTxIndex: uint32(vin.Vout),
				}
				*newOutpoints = append(*newOutpoints, outpoint)
			} else {
				// if vin.Coinbase == "" {
				// 	err := i.DbRepo.UpdateOutpointSpending(&db.UpdateOutpointSpendingData{
				// 		PreviousTxHash:  vin.Txid,
				// 		PreviousTxIndex: uint32(vin.Vout),

				// 		SpendingTxHash:       transaction.Txid,
				// 		SpendingTxIndex:      uint32(idxx),
				// 		SpendingBlockHash:    block.Hash,
				// 		SpendingBlockHeight:  uint64(blockHeight),
				// 		SpendingBlockTxIndex: uint32(txIdx),
				// 		Sequence:             uint32(vin.Sequence),
				// 		SignatureScript:      vin.ScriptSig.Hex,
				// 		Witness:              strings.Join(vin.Txinwitness, ","),
				// 	})
				// 	if err != nil {
				// 		return err
				// 	}
				// }
			}
		}
	}

	return nil
}

// TODO
func (i *Indexer) FindReorgHeight() {
	panic(errors.New("not implemented yet"))
}

// TODO
func (i *Indexer) Reorg() {
	panic(errors.New("not implemented yet"))

}
