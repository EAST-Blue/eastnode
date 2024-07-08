package main

import (
	"eastnode/indexer/peer"
	"eastnode/indexer/repository"
	"eastnode/indexer/store"
	storeDB "eastnode/utils/store"

	"github.com/btcsuite/btcd/chaincfg"

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
	s := storeDB.GetInstance(storeDB.IndexerDB)
	str := store.NewStorage(&chaincfg.RegressionNetParams, s.Gorm)
	p, err := peer.NewPeer("localhost:18444", str)
	if err != nil {
		panic(err)
	}

	go func() {
		err = p.Run()
		if err != nil {
			panic(err)
		}
	}()
	indexerRepo := repository.NewIndexerRepository(s.Gorm)

	// rpc
	blockchain := new(chain.Chain)
	bc := blockchain.Init(indexerRepo)

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
	err = http.ListenAndServe(":4000", router)
	if err != nil {
		panic(err)
	}
}
