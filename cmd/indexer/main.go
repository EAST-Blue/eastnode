package main

import (
	"eastnode/indexer"
	"eastnode/indexer/repository/bitcoin"
	"eastnode/indexer/repository/db"
	storeDB "eastnode/utils/store"
)

func main() {
	bitcoinRepo := bitcoin.NewBitcoinRepo("http://localhost:18443", "east", "east")
	s := storeDB.GetInstance(storeDB.IndexerDB)
	dbRepo := db.NewDBRepository(s.Gorm)
	indexerRepo := indexer.NewIndexer(dbRepo, bitcoinRepo)
	scheduler := indexer.NewScheduler(indexerRepo)

	scheduler.Start()
}
