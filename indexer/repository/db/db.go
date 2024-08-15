package db

import (
	"strconv"
	"strings"

	"github.com/libsv/go-bt/v2/bscript"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type DBRepository struct {
	Db *gorm.DB
}

func NewDBRepository(db *gorm.DB) *DBRepository {
	return &DBRepository{db}
}

// TODO: refactor this
// <sig> ... <redeem script>
func ParseP2shSigHexToAsms(hex string) (*P2shAsmScripts, error) {
	bs, err := bscript.NewFromHexString(hex)
	if err != nil {
		return nil, err
	}

	asm, err := bs.ToASM()
	if err != nil {
		return nil, err
	}

	lockHex := ""
	unlockHex := ""
	asms := strings.Split(asm, " ")
	switch len(asms) {
	case 2: // without a public key
		unlockHex = asms[0]
		lockHex = asms[1]
	case 3: // if the sig has public key
		unlockHex = asms[0]
		lockHex = asms[2]
	}

	lockScripts := []string{}
	unlockScripts := []string{}

	if lockHex != "" {
		bs, err := bscript.NewFromHexString(lockHex)
		if err != nil {
			return nil, err
		}

		asm, err := bs.ToASM()
		if err != nil {
			return nil, err
		}

		lockScripts = strings.Split(asm, " ")
	}

	if unlockHex != "" {
		bs, err := bscript.NewFromHexString(unlockHex)
		if err != nil {
			return nil, err
		}

		asm, err := bs.ToASM()
		if err != nil {
			return nil, err
		}

		if strings.HasSuffix(asm, "[error]") {
			unlockScripts = []string{unlockHex}
		} else {
			unlockScripts = strings.Split(asm, " ")
		}
	}

	return &P2shAsmScripts{
		LockScripts:   lockScripts,
		UnlockScripts: unlockScripts,
	}, nil
}

func (d *DBRepository) SetLastHeight(height int32) error {
	indexer := Indexer{
		Key:   INDEXER_LAST_HEIGHT_KEY,
		Value: strconv.Itoa(int(height)),
	}

	res := d.Db.Model(Indexer{}).Where("`key` = ?", INDEXER_LAST_HEIGHT_KEY).Updates(&indexer)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected > 0 {
		return nil
	}

	res = d.Db.Save(&indexer)
	return res.Error
}

func (d *DBRepository) SetLastHeightWithTx(tx *gorm.DB, height int32) error {
	indexer := Indexer{
		Key:   INDEXER_LAST_HEIGHT_KEY,
		Value: strconv.Itoa(int(height)),
	}

	res := tx.Model(Indexer{}).Where("`key` = ?", INDEXER_LAST_HEIGHT_KEY).Updates(&indexer)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected > 0 {
		return nil
	}

	res = tx.Save(&indexer)
	return res.Error
}

func (d *DBRepository) GetLastHeight() (int32, error) {
	indexer := Indexer{
		Key: INDEXER_LAST_HEIGHT_KEY,
	}
	res := d.Db.First(&indexer)
	if res.Error != nil {
		if res.Error == gorm.ErrRecordNotFound {
			return -1, nil
		}

		return 0, res.Error
	}

	height, _ := strconv.Atoi(indexer.Value)
	return int32(height), nil
}

func (d *DBRepository) CreateBlock(block *Block) error {
	err := d.Db.Create(block).Error
	if err == gorm.ErrDuplicatedKey {
		return nil
	}

	return err
}

func (d *DBRepository) CreateBlockWithTx(tx *gorm.DB, block *Block) error {
	return tx.Create(block).Error
}

func (d *DBRepository) CreateTransaction(transaction *Transaction) error {
	err := d.Db.Create(transaction).Error
	if err == gorm.ErrDuplicatedKey {
		return nil
	}

	return err
}

func (d *DBRepository) CreateTransactions(transactions *[]Transaction) error {
	err := d.Db.Clauses(clause.OnConflict{DoNothing: true}).CreateInBatches(transactions, 1024).Error
	if err == gorm.ErrDuplicatedKey {
		return nil
	}

	return err
}

func (d *DBRepository) CreateBlocks(blocks *[]Block) error {
	err := d.Db.Clauses(clause.OnConflict{DoNothing: true}).CreateInBatches(blocks, 1024).Error
	if err == gorm.ErrDuplicatedKey {
		return nil
	}

	return err
}

func (d *DBRepository) CreateTransactionWithTx(tx *gorm.DB, transaction *Transaction) error {
	return tx.Create(transaction).Error
}

func (d *DBRepository) CreateVin(outpoint *Vin) error {
	err := d.Db.Create(outpoint).Error
	if err == gorm.ErrDuplicatedKey {
		return nil
	}

	return err
}

func (d *DBRepository) CreateVins(outpoints *[]Vin) error {
	err := d.Db.CreateInBatches(outpoints, 1024).Error
	if err == gorm.ErrDuplicatedKey {
		return nil
	}

	return err
}

func (d *DBRepository) CreateVinWithTx(tx *gorm.DB, outpoint *Vin) error {
	return tx.Create(outpoint).Error
}

func (d *DBRepository) CreateVout(outpoint *Vout) error {
	err := d.Db.Create(outpoint).Error
	if err == gorm.ErrDuplicatedKey {
		return nil
	}

	return err
}

func (d *DBRepository) CreateVouts(outpoints *[]Vout) error {
	err := d.Db.CreateInBatches(outpoints, 1024).Error
	if err == gorm.ErrDuplicatedKey {
		return nil
	}

	return err
}

func (d *DBRepository) CreateVoutWithTx(tx *gorm.DB, outpoint *Vout) error {
	return tx.Create(outpoint).Error
}

func (d *DBRepository) GetBlockByHeight(height int64) (*Block, error) {
	block := &Block{}
	if resp := d.Db.First(block, "height = ?", height); resp.Error != nil {
		return block, resp.Error
	}
	return block, nil
}

func (d *DBRepository) GetTransactionsByBlockHash(blockHash string) ([]*Transaction, error) {
	transactions := []*Transaction{}
	if resp := d.Db.Order("block_index").Find(&transactions, "block_hash = ?", blockHash); resp.Error != nil {
		return nil, resp.Error
	}

	return transactions, nil
}

func (d *DBRepository) GetOutpointsByTransactionHash(transactionHash string) ([]*OutPoint, error) {
	outpoints := []*OutPoint{}

	// find and merge vin and vout
	vins := []*Vin{}
	vouts := []*Vout{}
	if resp := d.Db.Order("tx_index asc").Where("tx_hash = ? ", transactionHash).Find(&vins); resp.Error != nil {
		return nil, resp.Error
	}

	if resp := d.Db.Order("tx_index asc").Where("tx_hash = ? ", transactionHash).Find(&vouts); resp.Error != nil {
		return nil, resp.Error
	}

	for _, outpoint := range vins {
		outpoints = append(outpoints, &OutPoint{
			SpendingTxHash:       outpoint.TxHash,
			SpendingTxIndex:      outpoint.TxIndex,
			SpendingBlockHash:    outpoint.BlockHash,
			SpendingBlockHeight:  outpoint.BlockHeight,
			SpendingBlockTxIndex: outpoint.BlockTxIndex,
			Sequence:             outpoint.Sequence,
			SignatureScript:      outpoint.SignatureScript,
			Witness:              outpoint.Witness,
			PkScript:             outpoint.PkScript,
			Value:                outpoint.Value,
			Spender:              outpoint.Spender,
			Type:                 outpoint.Type,
			P2shAsmScripts:       outpoint.P2shAsmScripts,
			PkAsmScripts:         outpoint.PkAsmScripts,
			WitnessAsmScripts:    outpoint.WitnessAsmScripts,
		})
	}

	for _, outpoint := range vouts {
		outpoints = append(outpoints, &OutPoint{
			FundingTxHash:       outpoint.TxHash,
			FundingTxIndex:      outpoint.TxIndex,
			FundingBlockHash:    outpoint.BlockHash,
			FundingBlockHeight:  outpoint.BlockHeight,
			FundingBlockTxIndex: outpoint.BlockTxIndex,
			PkScript:            outpoint.PkScript,
			Value:               outpoint.Value,
			Spender:             outpoint.Spender,
			Type:                outpoint.Type,
			P2shAsmScripts:      outpoint.P2shAsmScripts,
			PkAsmScripts:        outpoint.PkAsmScripts,
		})
	}

	// TODO: handle these errors
	for i, v := range outpoints {
		if v.SignatureScript != "" {
			scripts, err := ParseP2shSigHexToAsms(v.SignatureScript)
			if err == nil {
				outpoints[i].P2shAsmScripts = scripts
			}
		}

		if v.PkScript != "" {
			bs, err := bscript.NewFromHexString(v.PkScript)
			if err == nil {
				asm, err := bs.ToASM()
				if err == nil {
					scripts := strings.Split(asm, " ")
					outpoints[i].PkAsmScripts = &scripts
				}
			}

		}

		if v.Witness != "" {
			witnesses := strings.Split(v.Witness, ",")
			if len(witnesses) == 3 {
				bs, err := bscript.NewFromHexString(witnesses[1])
				if err == nil {
					asm, err := bs.ToASM()
					if err == nil {
						scripts := strings.Split(asm, " ")
						outpoints[i].WitnessAsmScripts = &scripts
					}
				}
			}
		}
	}

	return outpoints, nil
}

func (d *DBRepository) DeleteBlocksFrom(height int32) error {
	// Start a transaction
	tx := d.Db.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	// Delete vouts
	if err := tx.Where("block_height >= ?", height).Delete(&Vout{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Delete vins
	if err := tx.Where("block_height >= ?", height).Delete(&Vin{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Delete transactions
	if err := tx.Where("block_height >= ?", height).Delete(&Transaction{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Delete blocks
	if err := tx.Where("height >= ?", height).Delete(&Block{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Commit the transaction
	if err := tx.Commit().Error; err != nil {
		return err
	}

	return nil
}
