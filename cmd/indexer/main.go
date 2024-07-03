package main

import (
	"eastnode/indexer/peer"
	"eastnode/indexer/store"
	storeDB "eastnode/utils/store"

	"github.com/btcsuite/btcd/chaincfg"
	// "eastnode/indexer/repository"
	// storeDB "eastnode/utils/store"
	// "fmt"
)

func main() {
	s := storeDB.GetInstance(storeDB.IndexerDB)
	str := store.NewStorage(&chaincfg.RegressionNetParams, s.Gorm)
	p, err := peer.NewPeer("localhost:18444", str)
	if err != nil {
		panic(err)
	}
	err = p.Run()
	if err != nil {
		panic(err)
	}

	// s := storeDB.GetInstance(storeDB.IndexerDB)
	// indexerRepo := repository.NewIndexerRepository(s.Gorm)
	//
	// outpoints, err := indexerRepo.GetOutpointsByTransactionHash("4564e73f976aea029f9614eb81f6cb2b3ac07f008eb0f5ec0149e6b07b037ba6")
	// if err != nil {
	// 	panic(err)
	// }
	//
	// for _, v := range outpoints {
	// 	fmt.Println(v.WitnessAsmScripts)
	// }
}
