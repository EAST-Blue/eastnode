package main

import (
	"eastnode/indexer"
	"eastnode/indexer/repository/bitcoin"
	indexerDb "eastnode/indexer/repository/db"
	storeDB "eastnode/utils/store"

	"eastnode/chain"
	"eastnode/jsonrpc"
	"log"
	"net/http"

	_ "github.com/dolthub/driver"
	"github.com/gorilla/mux"
	"github.com/gorilla/rpc"
	"github.com/gorilla/rpc/json"
)

func main() {
	// indexer
	bitcoinRepo := bitcoin.NewBitcoinRepo("http://localhost:18443", "east", "east")
	s := storeDB.GetInstance(storeDB.IndexerDB)
	indexerDbRepo := indexerDb.NewDBRepository(s.Gorm)
	indexerRepo := indexer.NewIndexer(indexerDbRepo, bitcoinRepo)
	scheduler := indexer.NewScheduler(indexerRepo)

	go func() {
		scheduler.CheckBlock()
	}()

	// rpc
	blockchain := new(chain.Chain)
	bc := blockchain.Init(indexerDbRepo)

	rpcServer := rpc.NewServer()

	rpcServer.RegisterCodec(json.NewCodec(), "application/json")
	rpcServer.RegisterCodec(json.NewCodec(), "application/json;charset=UTF-8")

	runtimeServer := &jsonrpc.RuntimeServer{
		Chain: bc,
	}
	commonServer := &jsonrpc.CommonServer{
		Chain: bc,
	}

	rpcServer.RegisterService(runtimeServer, "Runtime")
	rpcServer.RegisterService(commonServer, "Common")

	router := mux.NewRouter()
	router.Handle("/", rpcServer)

	log.Println("rpc is running")
	err := http.ListenAndServe(":4000", router)
	if err != nil {
		panic(err)
	}
}
