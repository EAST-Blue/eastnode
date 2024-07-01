package repository

import (
	"eastnode/indexer/model"
	"strings"

	"github.com/libsv/go-bt/v2/bscript"
	"gorm.io/gorm"
)

type IndexerRepository struct {
	db *gorm.DB
}

func NewIndexerRepository(db *gorm.DB) *IndexerRepository {
	return &IndexerRepository{db: db}
}

// TODO: refactor this
// <sig> ... <redeem script>
func ParseP2shSigHexToAsms(hex string) (*model.P2shAsmScripts, error) {
	bs, err := bscript.NewFromHexString(hex)
	if err != nil {
		return &model.P2shAsmScripts{}, err
	}

	asm, err := bs.ToASM()
	if err != nil {
		return &model.P2shAsmScripts{}, err
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
			return &model.P2shAsmScripts{}, err
		}

		asm, err := bs.ToASM()
		if err != nil {
			return &model.P2shAsmScripts{}, err
		}

		lockScripts = strings.Split(asm, " ")
	}

	if unlockHex != "" {
		bs, err := bscript.NewFromHexString(unlockHex)
		if err != nil {
			return &model.P2shAsmScripts{}, err
		}

		asm, err := bs.ToASM()
		if err != nil {
			return &model.P2shAsmScripts{}, err
		}

		if strings.HasSuffix(asm, "[error]") {
			unlockScripts = []string{unlockHex}
		} else {
			unlockScripts = strings.Split(asm, " ")
		}
	}

	return &model.P2shAsmScripts{
		LockScripts:   lockScripts,
		UnlockScripts: unlockScripts,
	}, nil
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

	// TODO: handle this error
	for i, v := range outpoints {
		if v.Type == "scripthash" && v.SignatureScript != "" {
			scripts, err := ParseP2shSigHexToAsms(v.SignatureScript)
			if err != nil {
				continue
			}
			outpoints[i].P2shAsmScripts = scripts
		}

		if v.PkScript != "" {
			bs, err := bscript.NewFromHexString(v.PkScript)
			if err != nil {
				continue
			}

			asm, err := bs.ToASM()
			if err != nil {
				continue
			}
			scripts := strings.Split(asm, " ")
			outpoints[i].PkAsmScripts = &scripts
		}

		if v.Witness != "" {
			bs, err := bscript.NewFromHexString(v.Witness)
			if err != nil {
				continue
			}

			asm, err := bs.ToASM()
			if err != nil {
				continue
			}
			scripts := strings.Split(asm, " ")
			outpoints[i].WitnessAsmScripts = &scripts
		}
	}

	return outpoints, nil
}
