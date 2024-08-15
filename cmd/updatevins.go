package main

import (
	"eastnode/indexer/repository/bitcoin"
	indexerDb "eastnode/indexer/repository/db"
	storeDB "eastnode/utils/store"
	"os"
	"strconv"
	"time"

	"log"

	_ "github.com/dolthub/driver"
	"github.com/joho/godotenv"
)

func main() {
	// indexer
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	log.Println("Initializing...")

	bitcoinRepo := bitcoin.NewBitcoinRepo(os.Getenv("BTC_RPC_URL"), "east", "east")
	s := storeDB.GetInstance(storeDB.IndexerDB)
	indexerDbRepo := indexerDb.NewDBRepository(s.Gorm)

	if err != nil {
		log.Fatal("Error loading .env file")
	}
	blockHeight := 0
	blockHash, err := bitcoinRepo.GetBlockHash(int32(blockHeight))
	if err != nil {
		log.Panic(err)
	}
	for {
		block, err := bitcoinRepo.GetBlock(blockHash)
		if err != nil {
			log.Panic(err)
		}

		log.Printf("handle block height %d, hash %s", blockHeight, block.Hash)
		for _, transaction := range block.Tx {
			for idxx, vin := range transaction.Vin {
				if vin.Coinbase == "" {
					res := indexerDbRepo.Db.Where("`tx_hash` = ? AND `tx_index` = ?", transaction.Txid, idxx).Model(indexerDb.Vin{}).Updates(map[string]interface{}{
						"funding_tx_hash":  vin.Txid,
						"funding_tx_index": vin.Vout,
					})
					log.Println(res.Error)
				}
			}
		}

		i, err := strconv.Atoi(os.Getenv("INDEXER_SLEEP_TIME"))
		if err != nil {
			log.Panic(err)
		}

		// Sleep because of RPC rate limit
		time.Sleep(time.Duration(i) * time.Millisecond)

		blockHeight++
		blockHash = block.Nextblockhash
	}
}
