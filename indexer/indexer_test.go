package indexer

import (
	"eastnode/indexer/repository/bitcoin"
	"eastnode/indexer/repository/db"
	utils "eastnode/utils/store"
	"log"
	"os"
	"testing"
)

func clearIndexerTest() {
	if err := os.RemoveAll("db_test"); err != nil {
		log.Panicln(err)
	}
}

func TestScheduler_Start(t *testing.T) {
	clearIndexerTest()

	// Setup
	instance := utils.GetFakeInstance(utils.IndexerDB, "../utils/store/test/doltdump.sql")
	dbRepo := db.NewDBRepository(instance.Gorm)
	mockBitcoinRepo := bitcoin.NewMockBitcoinRepo()

	indexer := NewIndexer(dbRepo, mockBitcoinRepo)

	startHeight := 0
	endHeight := 250
	for i := startHeight; i <= endHeight; i++ {
		mockBitcoinRepo.AddOrReplaceBlock(int32(i))
	}

	indexer.SyncBlocks(int32(startHeight), int32(endHeight))

	// Check the last block hash from indexed blocks and bitcoin repo
	lastIndexedBlock, err := indexer.DbRepo.GetBlockByHeight(int64(endHeight))
	if err != nil {
		t.Fatalf("Failed to get last indexed block: %v", err)
	}

	lastBitcoinBlockHash, err := indexer.bitcoinRepo.GetBlockHash(int32(endHeight))
	if err != nil {
		t.Fatalf("Failed to get last bitcoin block hash: %v", err)
	}

	if lastIndexedBlock.Hash != lastBitcoinBlockHash {
		t.Errorf("Last block hash mismatch. Indexed: %s, Bitcoin: %s", lastIndexedBlock.Hash, lastBitcoinBlockHash)
	}

	indexedBlockCount, err := indexer.bitcoinRepo.GetBlockCount()
	if err != nil {
		t.Fatalf("Failed to get block count: %v", err)
	}

	if indexedBlockCount != int32(endHeight-startHeight) {
		t.Errorf("Expected %d blocks to be indexed, got %d", endHeight-startHeight+1, indexedBlockCount)
	}

	clearIndexerTest()
}

func TestScheduler_ReorgSimulation(t *testing.T) {
	clearIndexerTest()

	// Setup
	instance := utils.GetFakeInstance(utils.IndexerDB, "../utils/store/test/doltdump.sql")
	dbRepo := db.NewDBRepository(instance.Gorm)
	mockBitcoinRepo := bitcoin.NewMockBitcoinRepo()

	indexer := NewIndexer(dbRepo, mockBitcoinRepo)

	// Initial blockchain setup
	startHeight := 0
	endHeight := 9
	for i := startHeight; i <= endHeight; i++ {
		mockBitcoinRepo.AddOrReplaceBlock(int32(i))
	}

	// Index initial blocks
	err := indexer.SyncBlocks(int32(startHeight), int32(endHeight))
	if err != nil {
		t.Fatalf("Failed to sync initial blocks: %v", err)
	}

	wrongBlockHash := []string{}
	// Save wrong block hash before reorg
	for i := 8; i <= 9; i++ {
		blockHash, err := mockBitcoinRepo.GetBlockHash(int32(i))
		if err != nil {
			t.Fatalf("Failed to get block hash for height %d: %v", i, err)
		}
		wrongBlockHash = append(wrongBlockHash, blockHash)
	}

	log.Println("===== Simulate reorg at block height 8 when syncing latest block 10 (n-2) =====")
	// Simulate reorg: replace blocks 8 and 9, add new block 10
	mockBitcoinRepo.AddOrReplaceBlock(8)
	mockBitcoinRepo.AddOrReplaceBlock(9)
	mockBitcoinRepo.AddOrReplaceBlock(10)

	// Run scheduler to handle reorg
	err = indexer.SyncBlocks(int32(endHeight+1), 10)
	if err != nil {
		t.Fatalf("Failed to sync blocks after reorg: %v", err)
	}

	// Verify reorg occurred
	for i := 8; i <= 10; i++ {
		blockHash, err := indexer.bitcoinRepo.GetBlockHash(int32(i))
		if err != nil {
			t.Fatalf("Failed to get block hash for height %d: %v", i, err)
		}

		// Check that the new block hash is different from the wrong block hash
		if i-8 < len(wrongBlockHash) && blockHash == wrongBlockHash[i-8] {
			t.Errorf("Block hash at height %d should have changed after reorg, but it's still %s", i, blockHash)
		}

		if i <= 9 {
			// Check if the block is marked as orphan
			orphanBlock, err := indexer.DbRepo.GetBlockByHeightWithIsOrphan(int64(i), true)
			if err != nil {
				t.Fatalf("Failed to get orphan block for height %d: %v", i, err)
			}
			if orphanBlock == nil {
				t.Errorf("Expected block at height %d to be marked as orphan, but it wasn't", i)
			} else {
				// Verify that the orphan block hash matches the wrong block hash
				if orphanBlock.Hash != wrongBlockHash[i-8] {
					t.Errorf("Orphan block hash mismatch at height %d. Expected: %s, Got: %s", i, wrongBlockHash[i-8], orphanBlock.Hash)
				}
			}
		}

		// Verify that the new block is correctly indexed and not marked as orphan
		indexedBlock, err := indexer.DbRepo.GetBlockByHeight(int64(i))
		if err != nil {
			t.Fatalf("Failed to get indexed block hash for height %d: %v", i, err)
		}

		if blockHash != indexedBlock.Hash {
			t.Errorf("Block hash mismatch at height %d. Expected: %s, Got: %s", i, blockHash, indexedBlock.Hash)
		}
	}

	// Verify final block count
	indexedBlockCount, err := indexer.DbRepo.GetLastHeight()
	if err != nil {
		t.Fatalf("Failed to get last indexed height: %v", err)
	}

	if indexedBlockCount != 10 {
		t.Errorf("Expected 11 blocks to be indexed after reorg, got %d", indexedBlockCount+1)
	}

	clearIndexerTest()
}
