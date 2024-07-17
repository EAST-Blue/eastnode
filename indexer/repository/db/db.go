package db

import (
	"strconv"
	"strings"

	"github.com/libsv/go-bt/v2/bscript"
	"gorm.io/gorm"
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

func (d *DBRepository) CreateTransactionWithTx(tx *gorm.DB, transaction *Transaction) error {
	return tx.Create(transaction).Error
}

func (d *DBRepository) CreateOutpoint(outpoint *OutPoint) error {
	err := d.Db.Create(outpoint).Error
	if err == gorm.ErrDuplicatedKey {
		return nil
	}

	return err
}

func (d *DBRepository) CreateOutpointWithTx(tx *gorm.DB, outpoint *OutPoint) error {
	return tx.Create(outpoint).Error
}

func (d *DBRepository) UpdateOutpointSpending(data *UpdateOutpointSpendingData) error {
	res := d.Db.Where("`funding_tx_hash` = ? AND `funding_tx_index` = ?", data.PreviousTxHash, data.PreviousTxIndex).Model(OutPoint{}).Updates(map[string]interface{}{
		"spending_tx_id":    data.SpendingTxID,
		"spending_tx_hash":  data.SpendingTxHash,
		"spending_tx_index": data.SpendingTxIndex,
		"sequence":          data.Sequence,
		"signature_script":  data.SignatureScript,
		"witness":           data.Witness,
	})
	return res.Error
}

func (d *DBRepository) UpdateOutpointSpendingWithTx(tx *gorm.DB, data *UpdateOutpointSpendingData) error {
	res := tx.Where("`funding_tx_hash` = ? AND `funding_tx_index` = ?", data.PreviousTxHash, data.PreviousTxIndex).Model(OutPoint{}).Updates(map[string]interface{}{
		"spending_tx_id":    data.SpendingTxID,
		"spending_tx_hash":  data.SpendingTxHash,
		"spending_tx_index": data.SpendingTxIndex,
		"sequence":          data.Sequence,
		"signature_script":  data.SignatureScript,
		"witness":           data.Witness,
	})
	return res.Error
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
	if resp := d.Db.Where("spending_tx_hash = ? OR funding_tx_hash = ?", transactionHash, transactionHash).Find(&outpoints); resp.Error != nil {
		return nil, resp.Error
	}

	// TODO: handle these errors
	for i, v := range outpoints {
		if v.Type == "scripthash" && v.SignatureScript != "" {
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
