package indexer

import (
	"eastnode/indexer/repository/bitcoin"
	"eastnode/indexer/repository/db"
	"errors"
	"log"
	"strings"
)

type Indexer struct {
	DbRepo      *db.DBRepository
	bitcoinRepo *bitcoin.BitcoinRepository
}

func NewIndexer(dbRepo *db.DBRepository, bitcoinRepo *bitcoin.BitcoinRepository) *Indexer {
	return &Indexer{dbRepo, bitcoinRepo}
}

func (i *Indexer) IndexBlocks(fromBlockHeight int32, toBlockHeight int32) error {
	log.Printf("index new blocks from %d to %d", fromBlockHeight, toBlockHeight)

	blockHeight := fromBlockHeight
	blockHash, err := i.bitcoinRepo.GetBlockHash(fromBlockHeight)
	if err != nil {
		return err
	}

	for {
		// break if current block-height is the latest, no need to index next block
		if !(blockHash != "" && (blockHeight <= toBlockHeight)) {
			break
		}

		block, err := i.bitcoinRepo.GetBlock(blockHash)
		if err != nil {
			return err
		}

		err = i.HandleBlock(blockHeight, block)
		if err != nil {
			return err
		}

		blockHeight++
		blockHash = block.Nextblockhash
	}

	return nil
}

func (i *Indexer) HandleBlock(blockHeight int32, block *bitcoin.GetBlock) error {
	log.Printf("handle block height %d, hash %s", blockHeight, block.Hash)

	// insert block
	newBlock := db.Block{
		Hash:          block.Hash,
		Height:        blockHeight,
		PreviousBlock: block.Previousblockhash,
		Version:       int32(block.Version),
		Nonce:         uint32(block.Nonce),
		Timestamp:     uint32(block.Time),
		Bits:          block.Bits,
		MerkleRoot:    block.Merkleroot,
	}
	err := i.DbRepo.CreateBlock(&newBlock)
	if err != nil {
		return err
	}

	// insert transaction
	// fill the txhash using txid instead of txhash, for the witness tx the id is different from the hash
	// https://bitcoin.stackexchange.com/questions/77699/whats-the-difference-between-txid-and-hash-getrawtransaction-bitcoind
	for idx, transaction := range block.Tx {
		newTx := db.Transaction{
			Hash:       transaction.Txid,
			LockTime:   uint32(transaction.Locktime),
			Version:    int32(transaction.Version),
			Safe:       false,
			BlockHash:  block.Hash,
			BlockIndex: uint32(idx),
		}
		err = i.DbRepo.CreateTransaction(&newTx)
		if err != nil {
			return err
		}

		// insert outpoints

		// vouts
		for idxx, vout := range transaction.Vout {
			err = i.DbRepo.CreateOutpoint(&db.OutPoint{
				FundingTxHash:  transaction.Txid,
				FundingTxIndex: uint32(idxx),
				PkScript:       vout.ScriptPubKey.Hex,
				Value:          int64(vout.Value),
				Spender:        vout.ScriptPubKey.Address,
			})
			if err != nil {
				return err
			}
		}

		// vins
		for idxx, vin := range transaction.Vin {
			// coinbase
			if vin.Coinbase != "" {
				err = i.DbRepo.CreateOutpoint(&db.OutPoint{
					SpendingTxHash:  transaction.Txid,
					SpendingTxIndex: uint32(idxx),
					Sequence:        uint32(vin.Sequence),
					SignatureScript: vin.ScriptSig.Hex,
					Witness:         strings.Join(vin.Txinwitness, ","),

					FundingTxHash:  vin.Txid,
					FundingTxIndex: uint32(vin.Vout),
				})
				if err != nil {
					return err
				}
				continue
			}

			err = i.DbRepo.UpdateOutpointSpending(&db.UpdateOutpointSpendingData{
				PreviousTxHash:  vin.Txid,
				PreviousTxIndex: uint32(vin.Vout),

				SpendingTxHash:  transaction.Txid,
				SpendingTxIndex: uint32(idxx),
				Sequence:        uint32(vin.Sequence),
				SignatureScript: vin.ScriptSig.Hex,
				Witness:         strings.Join(vin.Txinwitness, ","),
			})
			if err != nil {
				return err
			}
		}
	}

	return i.DbRepo.SetLastHeight(blockHeight)
}

// TODO
func (i *Indexer) FindReorgHeight() {
	panic(errors.New("not implemented yet"))
}

// TODO
func (i *Indexer) Reorg() {
	panic(errors.New("not implemented yet"))

}
