package main

import (
	"eastnode/indexer"
	"eastnode/indexer/repository/bitcoin"
	"eastnode/indexer/repository/db"
	storeDB "eastnode/utils/store"
	"os"
)

func main() {
	bitcoinRepo := bitcoin.NewBitcoinRepo(os.Getenv("BTC_RPC_URL"), "east", "east")
	s := storeDB.GetInstance(storeDB.IndexerDB)
	dbRepo := db.NewDBRepository(s.Gorm)
	indexerRepo := indexer.NewIndexer(dbRepo, bitcoinRepo)
	scheduler := indexer.NewScheduler(indexerRepo)

	scheduler.Start()
}
