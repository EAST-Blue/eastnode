package main

import (
	"eastnode/chain"
	"eastnode/indexer/repository"
	"eastnode/jsonrpc"
	store "eastnode/utils/store"
	"log"
	"net/http"

	_ "github.com/dolthub/driver"
	"github.com/gorilla/mux"
	"github.com/gorilla/rpc"
	"github.com/gorilla/rpc/json"
)

func main() {
	indexerInstance := store.GetInstance(store.IndexerDB)
	indexerRepo := repository.NewIndexerRepository(indexerInstance.Gorm)

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
	http.ListenAndServe(":8080", router)
}
