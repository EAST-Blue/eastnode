package bitcoin

import (
	"errors"
	"fmt"
	"math/rand"
	"time"
)

type MockBitcoinRepo struct {
	blocks       map[int32]string
	blockDetails map[string]*GetBlock
	blockCount   int32
}

func NewMockBitcoinRepo() *MockBitcoinRepo {
	repo := &MockBitcoinRepo{
		blocks:       make(map[int32]string),
		blockDetails: make(map[string]*GetBlock),
		blockCount:   0,
	}
	return repo
}

func (m *MockBitcoinRepo) GetBlockHash(height int32) (string, error) {
	if hash, ok := m.blocks[height]; ok {
		return hash, nil
	}
	return "", errors.New("block not found")
}

func (m *MockBitcoinRepo) GetBlock(blockHash string) (*GetBlock, error) {
	if block, ok := m.blockDetails[blockHash]; ok {
		return block, nil
	}
	return nil, errors.New("block not found")
}

func (m *MockBitcoinRepo) GetBlockCount() (int32, error) {
	return m.blockCount, nil
}

func (m *MockBitcoinRepo) AddOrReplaceBlock(height int32) GetBlock {
	// Get previous block hash
	var previousblockhash string
	if height > 0 {
		previousblockhash = m.blocks[height-1]
	}

	// Generate current block hash
	currentHash := fmt.Sprintf("hash_%s", generateRandomHash())

	// Update next block hash for the previous block
	if height > 0 {
		if prevBlock, ok := m.blockDetails[previousblockhash]; ok {
			prevBlock.Nextblockhash = currentHash
		}
	}

	// Set next block hash to empty string, it will be updated when the next block is added
	nextblockhash := ""

	// Update block count if this is a new highest block
	if height > m.blockCount {
		m.blockCount = height
	}

	// Store the block hash
	m.blocks[height] = currentHash

	block := &GetBlock{
		Hash:              currentHash,
		Confirmations:     int(m.blockCount - height + 1),
		Height:            int(height),
		Version:           2,
		VersionHex:        "00000002",
		Merkleroot:        fmt.Sprintf("merkle_%d", height),
		Time:              int(time.Now().Unix()),
		Mediantime:        int(time.Now().Unix()) - 600,
		Nonce:             rand.Intn(100000),
		Bits:              "1d00ffff",
		Difficulty:        1.0,
		Chainwork:         "0000000000000000000000000000000000000000000000000000000100010001",
		NTx:               rand.Intn(1000) + 1,
		Strippedsize:      rand.Intn(1000000) + 500000,
		Size:              rand.Intn(1000000) + 1000000,
		Weight:            rand.Intn(4000000) + 3000000,
		Previousblockhash: previousblockhash,
		Nextblockhash:     nextblockhash,
		Tx: []struct {
			Txid     string `json:"txid"`
			Hash     string `json:"hash"`
			Version  int    `json:"version"`
			Size     int    `json:"size"`
			Vsize    int    `json:"vsize"`
			Weight   int    `json:"weight"`
			Locktime int    `json:"locktime"`
			Vin      []struct {
				Coinbase    string   `json:"coinbase"`
				Txid        string   `json:"txid"`
				Vout        int64    `json:"vout"`
				Txinwitness []string `json:"txinwitness"`
				Sequence    int64    `json:"sequence"`
				ScriptSig   struct {
					Asm string `json:"asm"`
					Hex string `json:"hex"`
				} `json:"scriptSig"`
				PrevOutput struct {
					Value        float64 `json:"value"`
					ScriptPubKey struct {
						Asm     string `json:"asm"`
						Desc    string `json:"desc"`
						Hex     string `json:"hex"`
						Address string `json:"address"`
						Type    string `json:"type"`
					} `json:"scriptPubKey"`
				} `json:"prevout"`
			} `json:"vin"`
			Vout []struct {
				Value        float64 `json:"value"`
				N            int     `json:"n"`
				ScriptPubKey struct {
					Asm     string `json:"asm"`
					Desc    string `json:"desc"`
					Hex     string `json:"hex"`
					Address string `json:"address"`
					Type    string `json:"type"`
				} `json:"scriptPubKey"`
			} `json:"vout"`
			Hex string `json:"hex"`
		}{},
	}

	// Generate random transactions
	numTx := rand.Intn(100) + 1
	for i := 0; i < numTx; i++ {
		tx := struct {
			Txid     string `json:"txid"`
			Hash     string `json:"hash"`
			Version  int    `json:"version"`
			Size     int    `json:"size"`
			Vsize    int    `json:"vsize"`
			Weight   int    `json:"weight"`
			Locktime int    `json:"locktime"`
			Vin      []struct {
				Coinbase    string   `json:"coinbase"`
				Txid        string   `json:"txid"`
				Vout        int64    `json:"vout"`
				Txinwitness []string `json:"txinwitness"`
				Sequence    int64    `json:"sequence"`
				ScriptSig   struct {
					Asm string `json:"asm"`
					Hex string `json:"hex"`
				} `json:"scriptSig"`
				PrevOutput struct {
					Value        float64 `json:"value"`
					ScriptPubKey struct {
						Asm     string `json:"asm"`
						Desc    string `json:"desc"`
						Hex     string `json:"hex"`
						Address string `json:"address"`
						Type    string `json:"type"`
					} `json:"scriptPubKey"`
				} `json:"prevout"`
			} `json:"vin"`
			Vout []struct {
				Value        float64 `json:"value"`
				N            int     `json:"n"`
				ScriptPubKey struct {
					Asm     string `json:"asm"`
					Desc    string `json:"desc"`
					Hex     string `json:"hex"`
					Address string `json:"address"`
					Type    string `json:"type"`
				} `json:"scriptPubKey"`
			} `json:"vout"`
			Hex string `json:"hex"`
		}{
			Txid:     fmt.Sprintf("tx_%d_%d", height, i),
			Hash:     fmt.Sprintf("txhash_%d_%d", height, i),
			Version:  2,
			Size:     rand.Intn(1000) + 200,
			Vsize:    rand.Intn(1000) + 200,
			Weight:   rand.Intn(4000) + 800,
			Locktime: rand.Intn(500000),
		}
		block.Tx = append(block.Tx, tx)
	}

	// Add the block details to the blockDetails map
	m.blockDetails[block.Hash] = block

	return *block
}

func generateRandomHash() string {
	const charset = "abcdef0123456789"
	length := 64
	result := make([]byte, length)
	for i := range result {
		result[i] = charset[rand.Intn(len(charset))]
	}
	return string(result)
}
