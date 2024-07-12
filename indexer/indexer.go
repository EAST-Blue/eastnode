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

	tx := i.DbRepo.Db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

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
	err := i.DbRepo.CreateBlockWithTx(tx, &newBlock)
	if err != nil {
		tx.Rollback()
		return err
	}

	// insert transaction
	for idx, transaction := range block.Tx {
		newTx := db.Transaction{
			Hash:       transaction.Hash,
			LockTime:   uint32(transaction.Locktime),
			Version:    int32(transaction.Version),
			Safe:       false,
			BlockID:    newBlock.ID,
			BlockHash:  block.Hash,
			BlockIndex: uint32(idx),
		}
		err = i.DbRepo.CreateTransactionWithTx(tx, &newTx)
		if err != nil {
			tx.Rollback()
			return err
		}

		// insert outpoint

		// vins
		for idxx, vin := range transaction.Vin {
			err = i.DbRepo.CreateOutpointWithTx(tx, &db.OutPoint{
				SpendingTxID:    newTx.ID,
				SpendingTxHash:  transaction.Hash,
				SpendingTxIndex: uint32(idxx),
				Sequence:        uint32(vin.Sequence),
				SignatureScript: vin.ScriptSig.Hex,
				Witness:         strings.Join(vin.Txinwitness, ","),
			})
			if err != nil {
				tx.Rollback()
				return err
			}
		}

		// vouts
		for idxx, vout := range transaction.Vout {
			err = i.DbRepo.CreateOutpointWithTx(tx, &db.OutPoint{
				FundingTxID:    newTx.ID,
				FundingTxHash:  transaction.Hash,
				FundingTxIndex: uint32(idxx),
				PkScript:       vout.ScriptPubKey.Hex,
				Value:          int64(vout.Value),
				Spender:        vout.ScriptPubKey.Address,
			})
			if err != nil {
				tx.Rollback()
				return err
			}
		}
	}

	err = i.DbRepo.SetLastHeightWithTx(tx, blockHeight)
	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

// TODO
func (i *Indexer) FindReorgHeight() {
	panic(errors.New("not implemented yet"))
}

// TODO
func (i *Indexer) Reorg() {
	panic(errors.New("not implemented yet"))

}
