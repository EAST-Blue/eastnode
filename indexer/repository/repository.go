package repository

import (
	"eastnode/indexer/model"
	"gorm.io/gorm"
)

type IndexerRepository struct {
	db *gorm.DB
}

func NewIndexerRepository(db *gorm.DB) *IndexerRepository {
	return &IndexerRepository{db: db}
}

func (i *IndexerRepository) GetBlockByHeight(height int64) (*model.Block, error) {
	block := &model.Block{}
	if resp := i.db.First(block, "height = ?", height); resp.Error != nil {
		return block, resp.Error
	}
	return block, nil
}

func (i *IndexerRepository) GetTransactionsByBlockHash(blockHash string) ([]*model.Transaction, error) {
	transactions := []*model.Transaction{}
	if resp := i.db.Order("block_index").Find(&transactions, "block_hash = ?", blockHash); resp.Error != nil {
		return nil, resp.Error
	}

	return transactions, nil
}

func (i *IndexerRepository) GetOutpointsByTransactionHash(transactionHash string) ([]*model.OutPoint, error) {
	outpoints := []*model.OutPoint{}
	if resp := i.db.Where("spending_tx_hash = ? OR funding_tx_hash = ?", transactionHash, transactionHash).Find(&outpoints); resp.Error != nil {
		return nil, resp.Error
	}

	return outpoints, nil
}
